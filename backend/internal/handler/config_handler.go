package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"dvr-vod-system/internal/config"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	config *config.Config
}

// NewConfigHandler 创建新的配置处理器
func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{
		config: cfg,
	}
}

// ConfigResponse 配置响应（仅返回非敏感信息）
type ConfigResponse struct {
	ServerPort   int    `json:"server_port"`
	DVRCount     int    `json:"dvr_count"`
	RetryEnabled bool   `json:"retry_enabled"`
	RetryCount   int    `json:"retry_count"`
	Version      string `json:"version"`
}

// Handle 返回公开的配置信息
func (h *ConfigHandler) Handle(c *gin.Context) {
	clientIP := c.ClientIP()
	log.Printf("[CONFIG] 获取配置请求 - IP: %s", clientIP)
	
	if h.config == nil {
		log.Printf("[CONFIG] 配置未加载 - IP: %s", clientIP)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "config not loaded",
		})
		return
	}

	response := ConfigResponse{
		ServerPort:   h.config.Server.Port,
		DVRCount:     len(h.config.DVRServers),
		RetryEnabled: h.config.DVR.Retry > 0,
		RetryCount:   h.config.DVR.Retry,
		Version:      "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}
