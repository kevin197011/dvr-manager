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

// NewRouter 创建新的路由；cacheTTLDays 为录像 URL 缓存保留天数（0 表示默认 30 天）
func NewRouter(cfg *config.Config, cacheTTLDays int) *gin.Engine {
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
	auditRepo := repository.NewAuditRepository()
	userRepo := repository.NewUserRepository()
	ssoRepo := repository.NewSSORepository()
	recordingCacheRepo := repository.NewRecordingCacheRepository()

	// 初始化服务
	cacheInstance := cache.NewSQLiteCache(recordingCacheRepo, cacheTTLDays)
	dvrService := service.NewDVRService(cfg, dvrRepo)
	proxyService := service.NewProxyService(cfg)
	configService := service.NewConfigService(configRepo, dvrRepo)
	authService := service.NewAuthService(userRepo)
	ssoService := service.NewSSOService(ssoRepo)

	// JWT Secret（从环境变量读取，默认使用固定值）
	jwtSecret := os.Getenv("JWT_SECRET")
	authHandler := handler.NewAuthHandler(authService, jwtSecret, auditRepo)

	// 初始化处理器
	playHandler := handler.NewPlayHandler(dvrService, cacheInstance, auditRepo)
	proxyHandler := handler.NewProxyHandler(proxyService, dvrService, cacheInstance, auditRepo)
	configHandler := handler.NewConfigHandler(cfg)
	healthHandler := handler.NewHealthHandler(cfg)
	adminHandler := handler.NewAdminHandler(configService, auditRepo)
	auditHandler := handler.NewAuditHandler(auditRepo)
	userHandler := handler.NewUserHandler(authService, auditRepo)
	ssoHandler := handler.NewSSOHandler(ssoService, authService, auditRepo, jwtSecret)
	ssoAdminHandler := handler.NewSSOAdminHandler(ssoRepo, ssoService, auditRepo)

	// 认证路由（不需要认证）
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", authHandler.Me)
		auth.POST("/logout", authHandler.Logout)

		// SSO 登录与回调（仅 OIDC）
		auth.GET("/sso/providers", ssoHandler.ListProviders)
		auth.GET("/sso/oidc/:id/login", ssoHandler.OIDCLogin)
		auth.GET("/sso/oidc/:id/callback", ssoHandler.OIDCCallback)
	}

	// 需要登录的认证相关路由
	authProtected := r.Group("/api/auth")
	authProtected.Use(middleware.AuthMiddleware(authHandler))
	{
		authProtected.POST("/change-password", authHandler.ChangePassword)
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
		admin.GET("/audit", auditHandler.GetAudit)
		admin.POST("/audit/cleanup", auditHandler.Cleanup)

		// 用户管理
		admin.GET("/users", userHandler.List)
		admin.POST("/users", userHandler.Create)
		admin.PUT("/users/:id/role", userHandler.UpdateRole)
		admin.POST("/users/:id/reset-password", userHandler.ResetPassword)
		admin.DELETE("/users/:id", userHandler.Delete)

		// SSO 提供商管理
		admin.GET("/sso/providers", ssoAdminHandler.List)
		admin.POST("/sso/providers", ssoAdminHandler.Create)
		admin.PUT("/sso/providers/:id", ssoAdminHandler.Update)
		admin.POST("/sso/providers/:id/toggle", ssoAdminHandler.Toggle)
		admin.DELETE("/sso/providers/:id", ssoAdminHandler.Delete)
	}

	// 视频流代理路由（可选认证，仅用于在审计日志中记录登录用户）
	stream := r.Group("/stream")
	stream.Use(middleware.OptionalAuthMiddleware(authHandler))
	{
		stream.GET("/:filename", proxyHandler.Handle)
	}

	// 健康检查（支持 GET 和 HEAD）
	r.GET("/health", healthHandler.Handle)
	r.HEAD("/health", healthHandler.Handle)

	return r
}
