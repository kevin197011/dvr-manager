package handler

import (
	"errors"
	"net/http"
	"time"

	"dvr-vod-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
	jwtSecret   []byte
}

// NewAuthHandler 创建新的认证处理器
func NewAuthHandler(authService service.AuthService, jwtSecret string) *AuthHandler {
	if jwtSecret == "" {
		jwtSecret = "dvr-vod-system-secret-key-change-in-production"
	}
	return &AuthHandler{
		authService: authService,
		jwtSecret:   []byte(jwtSecret),
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	User    UserInfo `json:"user,omitempty"`
	Message string `json:"message,omitempty"`
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

// Claims JWT Claims
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Login 处理登录请求
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "请求参数错误",
		})
		return
	}

	// 验证用户名和密码
	user, err := h.authService.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "用户名或密码错误",
		})
		return
	}

	// 生成 JWT Token
	token, err := h.generateToken(user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "生成令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Token:   token,
		User: UserInfo{
			Username: user.Username,
			Role:     user.Role,
		},
	})
}

// Me 获取当前用户信息
func (h *AuthHandler) Me(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, VerifyResponse{
			Success: false,
		})
		return
	}

	// 移除 "Bearer " 前缀
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims, err := h.parseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, VerifyResponse{
			Success: false,
		})
		return
	}

	c.JSON(http.StatusOK, VerifyResponse{
		Success: true,
		User: UserInfo{
			Username: claims.Username,
			Role:     claims.Role,
		},
	})
}

// Logout 处理登出请求
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT 是无状态的，登出只需要客户端删除 token
	// 如果需要服务端控制，可以实现 token 黑名单
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登出成功",
	})
}

// generateToken 生成 JWT Token
func (h *AuthHandler) generateToken(username, role string) (string, error) {
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

// parseToken 解析 JWT Token
func (h *AuthHandler) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return h.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// VerifyToken 验证 Token（用于中间件）
func (h *AuthHandler) VerifyToken(tokenString string) (*Claims, error) {
	return h.parseToken(tokenString)
}
