package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"dvr-manager/internal/repository"
	"dvr-manager/internal/service"

	"github.com/gin-gonic/gin"
)

// SSOAdminHandler 管理员对 SSO 提供商进行 CRUD
type SSOAdminHandler struct {
	repo       repository.SSORepository
	ssoService service.SSOService
	auditRepo  repository.AuditRepository
}

// NewSSOAdminHandler 创建 SSO 管理处理器
func NewSSOAdminHandler(repo repository.SSORepository, ssoService service.SSOService, auditRepo repository.AuditRepository) *SSOAdminHandler {
	return &SSOAdminHandler{repo: repo, ssoService: ssoService, auditRepo: auditRepo}
}

// SSOProviderRequest 新建/更新请求
type SSOProviderRequest struct {
	Type    string                 `json:"type"` // 仅新建必填
	Name    string                 `json:"name" binding:"required"`
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config" binding:"required"`
}

func (h *SSOAdminHandler) audit(c *gin.Context, action, resource, detail, status string) {
	if h.auditRepo == nil {
		return
	}
	username, _ := c.Get("username")
	userStr, _ := username.(string)
	role, _ := c.Get("role")
	roleStr, _ := role.(string)
	_ = h.auditRepo.Insert(action, userStr, roleStr, c.ClientIP(), resource, detail, status)
}

// List GET /api/admin/sso/providers
func (h *SSOAdminHandler) List(c *gin.Context) {
	list, err := h.repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 把 config_json 转回对象，方便前端编辑（密钥/密码原样返回，前端注意脱敏展示）
	type item struct {
		ID      int64                  `json:"id"`
		Type    string                 `json:"type"`
		Name    string                 `json:"name"`
		Enabled bool                   `json:"enabled"`
		Config  map[string]interface{} `json:"config"`
	}
	items := make([]item, 0, len(list))
	for _, p := range list {
		var cfg map[string]interface{}
		_ = json.Unmarshal([]byte(p.ConfigJSON), &cfg)
		items = append(items, item{
			ID:      p.ID,
			Type:    p.Type,
			Name:    p.Name,
			Enabled: p.Enabled,
			Config:  cfg,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "list": items})
}

// Create POST /api/admin/sso/providers
func (h *SSOAdminHandler) Create(c *gin.Context) {
	var req SSOProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}
	if req.Type != repository.SSOTypeOIDC {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "type 必须为 oidc"})
		return
	}
	if msg := validateProviderConfig(req.Type, req.Config); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": msg})
		return
	}
	raw, err := json.Marshal(req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "config 序列化失败"})
		return
	}
	p := &repository.SSOProvider{
		Type:       req.Type,
		Name:       req.Name,
		Enabled:    req.Enabled,
		ConfigJSON: string(raw),
	}
	created, err := h.repo.Create(p)
	if err != nil {
		h.audit(c, "sso_create", req.Name, err.Error(), "fail")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	_ = h.ssoService.Reload()
	h.audit(c, "sso_create", req.Name, fmt.Sprintf("新建 SSO 提供商（%s）", req.Type), "success")
	c.JSON(http.StatusOK, gin.H{"success": true, "id": created.ID})
}

// Update PUT /api/admin/sso/providers/:id
func (h *SSOAdminHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的 ID"})
		return
	}
	var req SSOProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}
	exist, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}
	if msg := validateProviderConfig(exist.Type, req.Config); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": msg})
		return
	}
	raw, err := json.Marshal(req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "config 序列化失败"})
		return
	}
	exist.Name = req.Name
	exist.Enabled = req.Enabled
	exist.ConfigJSON = string(raw)
	if err := h.repo.Update(exist); err != nil {
		h.audit(c, "sso_update", req.Name, err.Error(), "fail")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	_ = h.ssoService.Reload()
	h.audit(c, "sso_update", req.Name, "更新 SSO 提供商", "success")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Toggle POST /api/admin/sso/providers/:id/toggle
func (h *SSOAdminHandler) Toggle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的 ID"})
		return
	}
	exist, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := h.repo.SetEnabled(id, !exist.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	_ = h.ssoService.Reload()
	state := "已停用"
	if !exist.Enabled {
		state = "已启用"
	}
	h.audit(c, "sso_toggle", exist.Name, state, "success")
	c.JSON(http.StatusOK, gin.H{"success": true, "enabled": !exist.Enabled})
}

// Delete DELETE /api/admin/sso/providers/:id
func (h *SSOAdminHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "无效的 ID"})
		return
	}
	exist, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := h.repo.Delete(id); err != nil {
		h.audit(c, "sso_delete", exist.Name, err.Error(), "fail")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	_ = h.ssoService.Reload()
	h.audit(c, "sso_delete", exist.Name, "删除 SSO 提供商", "success")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// validateProviderConfig 简单校验关键字段
func validateProviderConfig(t string, cfg map[string]interface{}) string {
	if t == repository.SSOTypeOIDC {
		for _, k := range []string{"issuer", "client_id", "client_secret", "redirect_url"} {
			if v, _ := cfg[k].(string); v == "" {
				return "OIDC 缺少必填字段: " + k
			}
		}
	}
	return ""
}
