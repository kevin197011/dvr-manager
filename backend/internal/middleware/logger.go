package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 请求日志中间件
// 记录所有请求的 IP、方法、路径、状态码、耗时等信息
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		
		// 获取客户端 IP
		clientIP := c.ClientIP()
		
		// 获取请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		
		// 处理请求
		c.Next()
		
		// 计算耗时
		latency := time.Since(start)
		
		// 获取响应状态码
		statusCode := c.Writer.Status()
		
		// 构建日志消息
		if query != "" {
			log.Printf("[REQUEST] %s | %3d | %13v | %15s | %-7s %s?%s",
				method,
				statusCode,
				latency,
				clientIP,
				method,
				path,
				query,
			)
		} else {
			log.Printf("[REQUEST] %s | %3d | %13v | %15s | %-7s %s",
				method,
				statusCode,
				latency,
				clientIP,
				method,
				path,
			)
		}
	}
}
