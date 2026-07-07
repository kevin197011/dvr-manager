package middleware

import (
	"net/http"

	"dvr-manager/internal/auth"
	"dvr-manager/internal/config"

	"github.com/gin-gonic/gin"
)

func applyClaims(c *gin.Context, claims *auth.Claims) {
	c.Set("username", claims.Username)
	c.Set("role", claims.Role)
}

// AuthMiddleware 强制认证
func AuthMiddleware(jwt *auth.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := auth.ExtractBearer(c.GetHeader("Authorization"))
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未授权，请先登录"})
			c.Abort()
			return
		}
		claims, err := jwt.Verify(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "令牌无效或已过期"})
			c.Abort()
			return
		}
		applyClaims(c, claims)
		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证（有 Token 则解析用户信息）
func OptionalAuthMiddleware(jwt *auth.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := auth.ExtractBearer(c.GetHeader("Authorization"))
		if tokenString == "" {
			c.Next()
			return
		}
		if claims, err := jwt.Verify(tokenString); err == nil {
			applyClaims(c, claims)
		}
		c.Next()
	}
}

// PlayAuthMiddleware 录像播放：默认可选认证；require_auth_for_play=true 时强制登录
func PlayAuthMiddleware(jwt *auth.JWT) gin.HandlerFunc {
	required := AuthMiddleware(jwt)
	optional := OptionalAuthMiddleware(jwt)
	return func(c *gin.Context) {
		if config.RequireAuthForPlayEnabled() {
			required(c)
			return
		}
		optional(c)
	}
}

// AdminMiddleware 管理员权限
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}
