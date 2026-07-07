package middleware

import (
	"dvr-manager/internal/config"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS 中间件（每次请求读取最新配置）
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()
		if cfg == nil || !cfg.CORS.Enabled {
			c.Next()
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.CORS.AllowOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", cfg.CORS.AllowMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", cfg.CORS.AllowHeaders)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}
