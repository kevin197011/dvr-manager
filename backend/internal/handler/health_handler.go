package handler

import (
	"log"

	"dvr-vod-system/internal/config"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Handle 处理健康检查请求
func (h *HealthHandler) Handle(c *gin.Context) {
	cfg := config.GetConfig()
	dvrCount := 0
	if cfg != nil {
		dvrCount = len(cfg.DVRServers)
	}
	log.Printf("[INFO] 健康检查 - IP: %s", c.ClientIP())
	c.JSON(200, gin.H{
		"status":      "ok",
		"dvr_servers": dvrCount,
	})
}
