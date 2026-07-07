package handler

import (
	"log"
	"net/http"

	"dvr-manager/internal/config"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 配置处理器
type ConfigHandler struct{}

// NewConfigHandler 创建配置处理器
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// ConfigResponse 公开配置摘要
type ConfigResponse struct {
	ServerPort   int    `json:"server_port"`
	DVRCount     int    `json:"dvr_count"`
	RetryEnabled bool   `json:"retry_enabled"`
	RetryCount   int    `json:"retry_count"`
	Version      string `json:"version"`
}

// Handle 返回公开配置信息
func (h *ConfigHandler) Handle(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		log.Printf("[CONFIG] 配置未加载 - IP: %s", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config not loaded"})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		ServerPort:   cfg.Server.Port,
		DVRCount:     len(cfg.DVRServers),
		RetryEnabled: cfg.DVR.Retry > 0,
		RetryCount:   cfg.DVR.Retry,
		Version:      "1.0.0",
	})
}
