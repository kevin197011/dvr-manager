package handler

import (
	"log"

	"github.com/gin-gonic/gin"
	"dvr-vod-system/internal/config"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	config *config.Config
}

// NewHealthHandler 创建新的健康检查处理器
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		config: cfg,
	}
}

// Handle 处理健康检查请求
func (h *HealthHandler) Handle(c *gin.Context) {
	log.Printf("[INFO] 健康检查 - IP: %s", c.ClientIP())
	c.JSON(200, gin.H{
		"status":      "ok",
		"dvr_servers": len(h.config.DVRServers),
	})
}
