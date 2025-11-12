package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ConfigResponse 配置响应（仅返回非敏感信息）
type ConfigResponse struct {
	ServerPort   int    `json:"server_port"`
	DVRCount     int    `json:"dvr_count"`
	RetryEnabled bool   `json:"retry_enabled"`
	RetryCount   int    `json:"retry_count"`
	Version      string `json:"version"`
}

// ConfigHandler 返回公开的配置信息
func ConfigHandler(c *gin.Context) {
	config := GetConfig()
	if config == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "config not loaded",
		})
		return
	}

	response := ConfigResponse{
		ServerPort:   config.Server.Port,
		DVRCount:     len(config.DVRServers),
		RetryEnabled: config.DVR.Retry > 0,
		RetryCount:   config.DVR.Retry,
		Version:      "1.0.0",
	}

	c.JSON(http.StatusOK, response)
}
