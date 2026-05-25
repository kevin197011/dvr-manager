package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"dvr-vod-system/internal/repository"
	"dvr-vod-system/internal/service"

	"github.com/crewjam/saml"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// SSOHandler SSO 公共处理器（提供商列表 / OIDC & SAML 登录回调）
type SSOHandler struct {
	ssoService  service.SSOService
	authService service.AuthService
	auditRepo   repository.AuditRepository
	jwtSecret   []byte
}

// NewSSOHandler 创建 SSO 处理器
func NewSSOHandler(ssoService service.SSOService, authService service.AuthService, auditRepo repository.AuditRepository, jwtSecret string) *SSOHandler {
	if jwtSecret == "" {
		jwtSecret = "dvr-vod-system-secret-key-change-in-production"
	}
	return &SSOHandler{
		ssoService:  ssoService,
		authService: authService,
		auditRepo:   auditRepo,
		jwtSecret:   []byte(jwtSecret),
	}
}

// ListProviders GET /api/auth/sso/providers 列出已启用的 SSO 提供商（公开）
func (h *SSOHandler) ListProviders(c *gin.Context) {
	providers, err := h.ssoService.ListProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "providers": providers})
}

// parseID 从路径参数提取 provider id
func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的提供商 ID"})
		return 0, false
	}
	return id, true
}

// finishLogin 颁发 JWT、写审计、按需重定向到前端
func (h *SSOHandler) finishLogin(c *gin.Context, user *service.User, source string, redirect bool) {
	token, err := h.generateToken(user.Username, user.Role)
	if err != nil {
		log.Printf("[SSO] 生成令牌失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "生成令牌失败"})
		return
	}
	if h.auditRepo != nil {
		_ = h.auditRepo.Insert("login_success", user.Username, user.Role, c.ClientIP(), source, "SSO 登录成功", "success")
	}
	if redirect {
		// 重定向到前端的 SSO 回调页，把 token 带过去
		q := url.Values{}
		q.Set("token", token)
		q.Set("username", user.Username)
		q.Set("role", user.Role)
		c.Redirect(http.StatusFound, "/sso-callback?"+q.Encode())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"token":    token,
		"username": user.Username,
		"role":     user.Role,
	})
}

// generateToken 生成 JWT
func (h *SSOHandler) generateToken(username, role string) (string, error) {
	claims := Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

// ---------------- OIDC ----------------

// OIDCLogin GET /api/auth/sso/oidc/:id/login 跳转到 IdP
func (h *SSOHandler) OIDCLogin(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	state, err := service.GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	authURL, err := h.ssoService.BuildOIDCAuthURL(id, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	// state 写到 cookie，回调时核对
	c.SetCookie(oidcStateCookieName(id), state, 600, "/", "", c.Request.TLS != nil, true)
	c.Redirect(http.StatusFound, authURL)
}

// OIDCCallback GET /api/auth/sso/oidc/:id/callback
func (h *SSOHandler) OIDCCallback(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	clientIP := c.ClientIP()

	// 错误处理
	if errStr := c.Query("error"); errStr != "" {
		desc := c.Query("error_description")
		if h.auditRepo != nil {
			_ = h.auditRepo.Insert("login_fail", "", "", clientIP, fmt.Sprintf("oidc:%d", id), errStr+": "+desc, "fail")
		}
		c.Redirect(http.StatusFound, "/sso-callback?error="+url.QueryEscape(errStr+": "+desc))
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.Redirect(http.StatusFound, "/sso-callback?error="+url.QueryEscape("缺少 code/state"))
		return
	}

	expectedState, _ := c.Cookie(oidcStateCookieName(id))
	if expectedState == "" || expectedState != state {
		if h.auditRepo != nil {
			_ = h.auditRepo.Insert("login_fail", "", "", clientIP, fmt.Sprintf("oidc:%d", id), "state 校验失败", "fail")
		}
		c.Redirect(http.StatusFound, "/sso-callback?error="+url.QueryEscape("state 校验失败"))
		return
	}
	c.SetCookie(oidcStateCookieName(id), "", -1, "/", "", c.Request.TLS != nil, true)

	username, err := h.ssoService.ExchangeOIDC(c.Request.Context(), id, code)
	if err != nil {
		if h.auditRepo != nil {
			_ = h.auditRepo.Insert("login_fail", "", "", clientIP, fmt.Sprintf("oidc:%d", id), err.Error(), "fail")
		}
		c.Redirect(http.StatusFound, "/sso-callback?error="+url.QueryEscape(err.Error()))
		return
	}

	source := fmt.Sprintf("oidc:%d", id)
	user, err := h.authService.FindOrCreateSSOUser(username, source)
	if err != nil {
		if h.auditRepo != nil {
			_ = h.auditRepo.Insert("login_fail", username, "", clientIP, source, err.Error(), "fail")
		}
		c.Redirect(http.StatusFound, "/sso-callback?error="+url.QueryEscape(err.Error()))
		return
	}
	h.finishLogin(c, user, source, true)
}

func oidcStateCookieName(id int64) string {
	return fmt.Sprintf("oidc_state_%d", id)
}

// ---------------- SAML ----------------

// SAMLMetadata GET /api/auth/sso/saml/:id/metadata 返回 SP metadata XML
func (h *SSOHandler) SAMLMetadata(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	sp, err := h.ssoService.GetSAMLMiddleware(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}
	sp.ServeMetadata(c.Writer, c.Request)
}

// SAMLLogin GET /api/auth/sso/saml/:id/login 触发 SAML AuthnRequest
func (h *SSOHandler) SAMLLogin(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	sp, err := h.ssoService.GetSAMLMiddleware(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}
	// crewjam/saml 通过 HandleStartAuthFlow 自动 binding + relay state
	sp.HandleStartAuthFlow(c.Writer, c.Request)
}

// SAMLACS POST /api/auth/sso/saml/:id/acs 处理 SAML 响应
func (h *SSOHandler) SAMLACS(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	clientIP := c.ClientIP()
	sp, err := h.ssoService.GetSAMLMiddleware(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}
	cfg, _ := h.ssoService.GetSAMLConfig(id)

	if err := c.Request.ParseForm(); err != nil {
		c.Redirect(http.StatusFound, "/sso-callback?error="+url.QueryEscape("解析表单失败"))
		return
	}

	// 校验并解析 SAMLResponse，得到 Assertion
	assertion, err := sp.ServiceProvider.ParseResponse(c.Request, []string{})
	if err != nil {
		// 第二种签名：传入 possibleRequestIDs；空数组在已 IdP-initiated 场景下也合法
		var invalidResp *saml.InvalidResponseError
		if errors.As(err, &invalidResp) {
			err = invalidResp.PrivateErr
		}
		if h.auditRepo != nil {
			_ = h.auditRepo.Insert("login_fail", "", "", clientIP, fmt.Sprintf("saml:%d", id), err.Error(), "fail")
		}
		c.Redirect(http.StatusFound, "/sso-callback?error="+url.QueryEscape("SAML 响应校验失败: "+err.Error()))
		return
	}

	username := extractSAMLUsername(assertion, cfg.UsernameAttr)
	if username == "" {
		c.Redirect(http.StatusFound, "/sso-callback?error="+url.QueryEscape("SAML 响应缺少用户名"))
		return
	}

	source := fmt.Sprintf("saml:%d", id)
	user, err := h.authService.FindOrCreateSSOUser(username, source)
	if err != nil {
		if h.auditRepo != nil {
			_ = h.auditRepo.Insert("login_fail", username, "", clientIP, source, err.Error(), "fail")
		}
		c.Redirect(http.StatusFound, "/sso-callback?error="+url.QueryEscape(err.Error()))
		return
	}
	h.finishLogin(c, user, source, true)
}

func extractSAMLUsername(assertion *saml.Assertion, preferred string) string {
	if assertion == nil {
		return ""
	}
	// 1) 指定属性优先
	if preferred != "" {
		for _, st := range assertion.AttributeStatements {
			for _, a := range st.Attributes {
				if (a.FriendlyName == preferred || a.Name == preferred) && len(a.Values) > 0 && a.Values[0].Value != "" {
					return a.Values[0].Value
				}
			}
		}
	}
	// 2) 常见属性 email / mail / username / uid
	common := []string{"email", "mail", "username", "uid", "preferred_username"}
	for _, st := range assertion.AttributeStatements {
		for _, a := range st.Attributes {
			for _, name := range common {
				if strings.EqualFold(a.FriendlyName, name) || strings.EqualFold(a.Name, name) {
					if len(a.Values) > 0 && a.Values[0].Value != "" {
						return a.Values[0].Value
					}
				}
			}
		}
	}
	// 3) Fallback: NameID
	if assertion.Subject != nil && assertion.Subject.NameID != nil {
		return assertion.Subject.NameID.Value
	}
	return ""
}
