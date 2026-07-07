package router

import (
	"dvr-vod-system/internal/auth"
	"dvr-vod-system/internal/config"
	"dvr-vod-system/internal/handler"
	"dvr-vod-system/internal/middleware"
	"dvr-vod-system/internal/repository"
	"dvr-vod-system/internal/service"
	"dvr-vod-system/internal/web"
	"dvr-vod-system/pkg/cache"

	"github.com/gin-gonic/gin"
)

// NewRouter 创建路由
func NewRouter(cfg *config.Config, cacheTTLDays int, jwt *auth.JWT) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware())

	configRepo := repository.NewConfigRepository()
	dvrRepo := repository.NewDVRRepository()
	auditRepo := repository.NewAuditRepository()
	userRepo := repository.NewUserRepository()
	ssoRepo := repository.NewSSORepository()
	recordingCacheRepo := repository.NewRecordingCacheRepository()

	cacheInstance := cache.NewSQLiteCache(recordingCacheRepo, cacheTTLDays)
	dvrService := service.NewDVRService(cfg, dvrRepo)
	proxyService := service.NewProxyService(cfg)
	configService := service.NewConfigService(configRepo, dvrRepo)
	authService := service.NewAuthService(userRepo)
	ssoService := service.NewSSOService(ssoRepo)

	authHandler := handler.NewAuthHandler(authService, jwt, auditRepo)
	playHandler := handler.NewPlayHandler(dvrService, cacheInstance, auditRepo)
	proxyHandler := handler.NewProxyHandler(proxyService, dvrService, cacheInstance, auditRepo)
	configHandler := handler.NewConfigHandler()
	healthHandler := handler.NewHealthHandler()
	adminHandler := handler.NewAdminHandler(configService, auditRepo)
	auditHandler := handler.NewAuditHandler(auditRepo)
	userHandler := handler.NewUserHandler(authService, auditRepo)
	ssoHandler := handler.NewSSOHandler(ssoService, authService, auditRepo, jwt)
	ssoAdminHandler := handler.NewSSOAdminHandler(ssoRepo, ssoService, auditRepo)

	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", authHandler.Me)
		auth.POST("/logout", authHandler.Logout)
		auth.GET("/sso/providers", ssoHandler.ListProviders)
		auth.GET("/sso/oidc/:id/login", ssoHandler.OIDCLogin)
		auth.GET("/sso/oidc/:id/callback", ssoHandler.OIDCCallback)
	}

	authProtected := r.Group("/api/auth")
	authProtected.Use(middleware.AuthMiddleware(jwt))
	{
		authProtected.POST("/change-password", authHandler.ChangePassword)
	}

	playAuth := middleware.PlayAuthMiddleware(jwt)

	api := r.Group("/api")
	api.Use(playAuth)
	{
		api.POST("/play", playHandler.Handle)
		api.GET("/play", playHandler.Handle)
	}
	r.GET("/api/config", configHandler.Handle)

	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthMiddleware(jwt))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/config", adminHandler.GetConfig)
		admin.POST("/config", adminHandler.UpdateConfig)
		admin.GET("/dvr-servers", adminHandler.GetDVRServers)
		admin.POST("/dvr-servers", adminHandler.UpdateDVRServers)
		admin.POST("/reload", adminHandler.ReloadConfig)
		admin.GET("/audit", auditHandler.GetAudit)
		admin.POST("/audit/cleanup", auditHandler.Cleanup)
		admin.GET("/users", userHandler.List)
		admin.POST("/users", userHandler.Create)
		admin.PUT("/users/:id/role", userHandler.UpdateRole)
		admin.POST("/users/:id/reset-password", userHandler.ResetPassword)
		admin.DELETE("/users/:id", userHandler.Delete)
		admin.GET("/sso/providers", ssoAdminHandler.List)
		admin.POST("/sso/providers", ssoAdminHandler.Create)
		admin.PUT("/sso/providers/:id", ssoAdminHandler.Update)
		admin.POST("/sso/providers/:id/toggle", ssoAdminHandler.Toggle)
		admin.DELETE("/sso/providers/:id", ssoAdminHandler.Delete)
	}

	stream := r.Group("/stream")
	stream.Use(playAuth)
	{
		stream.GET("/:filename", proxyHandler.Handle)
	}

	r.GET("/health", healthHandler.Handle)
	r.HEAD("/health", healthHandler.Handle)

	web.Register(r)

	return r
}
