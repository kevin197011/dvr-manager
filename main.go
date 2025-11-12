package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置文件
	config, err := LoadConfig("config.yml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	log.Printf("Loaded %d DVR servers from config", len(config.DVRServers))

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// 配置 CORS（如果启用）
	if config.CORS.Enabled {
		r.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", config.CORS.AllowOrigins)
			c.Writer.Header().Set("Access-Control-Allow-Methods", config.CORS.AllowMethods)
			c.Writer.Header().Set("Access-Control-Allow-Headers", config.CORS.AllowHeaders)

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(200)
				return
			}

			c.Next()
		})
	}

	// 静态文件服务 - 托管 index.html 和图标
	r.StaticFile("/", "./index.html")
	r.StaticFile("/index.html", "./index.html")
	r.StaticFile("/favicon.svg", "./favicon.svg")
	r.StaticFile("/favicon.ico", "./favicon.svg")
	r.StaticFile("/favicon-192.png", "./favicon-192.png")
	r.StaticFile("/favicon-32.png", "./favicon-32.png")

	// API 路由
	r.POST("/api/play", PlayHandler)
	r.GET("/api/play", PlayHandler)

	// 视频流代理路由
	r.GET("/stream/:filename", StreamProxyHandler)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		log.Printf("[INFO] 健康检查 - IP: %s", c.ClientIP())
		c.JSON(200, gin.H{
			"status":      "ok",
			"dvr_servers": len(config.DVRServers),
		})
	})

	// 配置信息（仅返回非敏感信息）
	r.GET("/api/config", ConfigHandler)

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.Server.Port)
	log.Printf("========================================")
	log.Printf("DVR VOD System starting...")
	log.Printf("Server Address: http://localhost%s", addr)
	log.Printf("DVR Servers: %d", len(config.DVRServers))
	for i, server := range config.DVRServers {
		log.Printf("  [%d] %s", i+1, server)
	}
	log.Printf("========================================")
	
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
