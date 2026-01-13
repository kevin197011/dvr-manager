package router

import (
	"os"

	"dvr-vod-system/internal/config"
	"dvr-vod-system/internal/handler"
	"dvr-vod-system/internal/middleware"
	"dvr-vod-system/internal/repository"
	"dvr-vod-system/internal/service"
	"dvr-vod-system/pkg/cache"

	"github.com/gin-gonic/gin"
)

// NewRouter 创建新的路由
func NewRouter(cfg *config.Config) *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// 配置请求日志中间件（记录所有请求的 IP）
	r.Use(middleware.LoggerMiddleware())

	// 配置 CORS 中间件
	r.Use(middleware.CORSMiddleware(cfg))

	// 初始化仓库
	configRepo := repository.NewConfigRepository()
	dvrRepo := repository.NewDVRRepository()

	// 初始化服务
	cacheInstance := cache.NewMemoryCache()
	dvrService := service.NewDVRService(cfg, dvrRepo)
	proxyService := service.NewProxyService(cfg)
	configService := service.NewConfigService(configRepo, dvrRepo)
	authService := service.NewAuthService()

	// 初始化处理器
	playHandler := handler.NewPlayHandler(dvrService, cacheInstance)
	proxyHandler := handler.NewProxyHandler(proxyService, cacheInstance)
	configHandler := handler.NewConfigHandler(cfg)
	healthHandler := handler.NewHealthHandler(cfg)
	adminHandler := handler.NewAdminHandler(configService)
	
	// JWT Secret（从环境变量读取，默认使用固定值）
	jwtSecret := os.Getenv("JWT_SECRET")
	authHandler := handler.NewAuthHandler(authService, jwtSecret)

	// 认证路由（不需要认证）
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", authHandler.Me)
		auth.POST("/logout", authHandler.Logout)
	}

	// API 路由（可选认证，用于公开访问）
	api := r.Group("/api")
	api.Use(middleware.OptionalAuthMiddleware(authHandler))
	{
		api.POST("/play", playHandler.Handle)
		api.GET("/play", playHandler.Handle)
		api.GET("/config", configHandler.Handle)
	}

	// 管理后台 API 路由（需要认证和管理员权限）
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware(authHandler))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/config", adminHandler.GetConfig)
		admin.POST("/config", adminHandler.UpdateConfig)
		admin.GET("/dvr-servers", adminHandler.GetDVRServers)
		admin.POST("/dvr-servers", adminHandler.UpdateDVRServers)
		admin.POST("/reload", adminHandler.ReloadConfig)
	}

	// 视频流代理路由
	r.GET("/stream/:filename", proxyHandler.Handle)

	// 健康检查（支持 GET 和 HEAD）
	r.GET("/health", healthHandler.Handle)
	r.HEAD("/health", healthHandler.Handle)

	return r
}
