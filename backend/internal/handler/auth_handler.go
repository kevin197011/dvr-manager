package handler

import (
	"log"
	"net/http"

	"dvr-manager/internal/auth"
	"dvr-manager/internal/repository"
	"dvr-manager/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
	jwt         *auth.JWT
	auditRepo   repository.AuditRepository
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService service.AuthService, jwt *auth.JWT, auditRepo repository.AuditRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwt:         jwt,
		auditRepo:   auditRepo,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success bool     `json:"success"`
	Token   string   `json:"token,omitempty"`
	User    UserInfo `json:"user,omitempty"`
	Message string   `json:"message,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// VerifyResponse 验证响应
type VerifyResponse struct {
	Success bool     `json:"success"`
	User    UserInfo `json:"user,omitempty"`
}

// Login 处理登录
func (h *AuthHandler) Login(c *gin.Context) {
	clientIP := c.ClientIP()
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{Success: false, Message: "请求参数错误"})
		return
	}

	user, err := h.authService.Authenticate(req.Username, req.Password)
	if err != nil {
		if h.auditRepo != nil {
			_ = h.auditRepo.Insert("login_fail", req.Username, "", clientIP, "", "登录失败", "fail")
		}
		c.JSON(http.StatusUnauthorized, LoginResponse{Success: false, Message: "用户名或密码错误"})
		return
	}

	token, err := h.jwt.Generate(user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, LoginResponse{Success: false, Message: "生成令牌失败"})
		return
	}

	if h.auditRepo != nil {
		_ = h.auditRepo.Insert("login_success", user.Username, user.Role, clientIP, "", "登录成功", "success")
	}
	log.Printf("[AUTH] 登录成功 - IP: %s, 用户名: %s", clientIP, user.Username)
	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Token:   token,
		User:    UserInfo{Username: user.Username, Role: user.Role},
	})
}

// Me 当前用户
func (h *AuthHandler) Me(c *gin.Context) {
	tokenString := auth.ExtractBearer(c.GetHeader("Authorization"))
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, VerifyResponse{Success: false})
		return
	}
	claims, err := h.jwt.Verify(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, VerifyResponse{Success: false})
		return
	}
	c.JSON(http.StatusOK, VerifyResponse{
		Success: true,
		User:    UserInfo{Username: claims.Username, Role: claims.Role},
	})
}

// ChangePasswordRequest 修改密码
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	clientIP := c.ClientIP()
	usernameVal, _ := c.Get("username")
	username, _ := usernameVal.(string)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未登录"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	if err := h.authService.ChangePassword(username, req.OldPassword, req.NewPassword); err != nil {
		if h.auditRepo != nil {
			_ = h.auditRepo.Insert("password_change", username, "", clientIP, "", err.Error(), "fail")
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if h.auditRepo != nil {
		_ = h.auditRepo.Insert("password_change", username, "", clientIP, "", "修改自己的密码", "success")
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "密码修改成功"})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "登出成功"})
}
