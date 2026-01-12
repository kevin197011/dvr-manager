package handler

import (
	"log"
	"net/http"
	"time"

	"dvr-vod-system/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminHandler 管理后台处理器
type AdminHandler struct {
	configService service.ConfigService
}

// NewAdminHandler 创建新的管理后台处理器
func NewAdminHandler(configService service.ConfigService) *AdminHandler {
	return &AdminHandler{
		configService: configService,
	}
}

// GetDVRServersRequest 获取 DVR 服务器列表请求
type GetDVRServersResponse struct {
	Success bool     `json:"success"`
	Servers []string `json:"servers"`
	Count   int      `json:"count"`
}

// UpdateDVRServersRequest 更新 DVR 服务器列表请求
type UpdateDVRServersRequest struct {
	Servers []string `json:"servers" binding:"required"`
}

// UpdateDVRServersResponse 更新 DVR 服务器列表响应
type UpdateDVRServersResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// GetDVRServers 获取 DVR 服务器列表
func (h *AdminHandler) GetDVRServers(c *gin.Context) {
	servers := h.configService.GetDVRServers()

	c.JSON(http.StatusOK, GetDVRServersResponse{
		Success: true,
		Servers: servers,
		Count:   len(servers),
	})
}

// UpdateDVRServers 更新 DVR 服务器列表
func (h *AdminHandler) UpdateDVRServers(c *gin.Context) {
	var req UpdateDVRServersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] 请求解析失败 - IP: %s, Error: %v", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, UpdateDVRServersResponse{
			Success: false,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	// 验证服务器列表不为空
	if len(req.Servers) == 0 {
		c.JSON(http.StatusBadRequest, UpdateDVRServersResponse{
			Success: false,
			Message: "servers list cannot be empty",
		})
		return
	}

	// 更新配置
	if err := h.configService.UpdateDVRServers(req.Servers); err != nil {
		log.Printf("[ERROR] 更新配置失败 - IP: %s, Error: %v", c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, UpdateDVRServersResponse{
			Success: false,
			Message: "failed to update config: " + err.Error(),
		})
		return
	}

	log.Printf("[INFO] DVR 服务器列表已更新 - IP: %s, 数量: %d", c.ClientIP(), len(req.Servers))
	c.JSON(http.StatusOK, UpdateDVRServersResponse{
		Success: true,
		Message: "DVR servers updated successfully",
		Count:   len(req.Servers),
	})
}

// GetConfig 获取完整配置
func (h *AdminHandler) GetConfig(c *gin.Context) {
	cfg, err := h.configService.GetConfig()
	if err != nil {
		log.Printf("[ERROR] 获取配置失败 - IP: %s, Error: %v", c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get config: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  cfg,
	})
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	Server     interface{} `json:"server"`
	DVR        interface{} `json:"dvr"`
	DVRServers []string    `json:"dvr_servers"`
	CORS       interface{} `json:"cors"`
}

// UpdateConfig 更新完整配置
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] 请求解析失败 - IP: %s, Error: %v", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	// 获取当前配置
	cfg, err := h.configService.GetConfig()
	if err != nil {
		log.Printf("[ERROR] 获取配置失败 - IP: %s, Error: %v", c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to get config: " + err.Error(),
		})
		return
	}

	// 更新服务器配置
	if req.Server != nil {
		if serverMap, ok := req.Server.(map[string]interface{}); ok {
			if port, ok := serverMap["port"].(float64); ok {
				cfg.Server.Port = int(port)
			}
			if timeout, ok := serverMap["timeout"].(float64); ok {
				cfg.Server.Timeout = time.Duration(timeout) * time.Second
			}
		}
	}

	// 更新 DVR 配置
	if req.DVR != nil {
		if dvrMap, ok := req.DVR.(map[string]interface{}); ok {
			if timeout, ok := dvrMap["timeout"].(float64); ok {
				cfg.DVR.Timeout = time.Duration(timeout) * time.Second
			}
			if retry, ok := dvrMap["retry"].(float64); ok {
				cfg.DVR.Retry = int(retry)
			}
			if skipTLSVerify, ok := dvrMap["skip_tls_verify"].(bool); ok {
				cfg.DVR.SkipTLSVerify = skipTLSVerify
			}
		}
	}

	// 更新 DVR 服务器列表
	if len(req.DVRServers) > 0 {
		cfg.DVRServers = req.DVRServers
	}

	// 保存配置
	if err := h.configService.UpdateConfig(cfg); err != nil {
		log.Printf("[ERROR] 更新配置失败 - IP: %s, Error: %v", c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to update config: " + err.Error(),
		})
		return
	}

	log.Printf("[INFO] 配置已更新 - IP: %s", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "config updated successfully",
	})
}

// ReloadConfig 重新加载配置
func (h *AdminHandler) ReloadConfig(c *gin.Context) {
	if err := h.configService.ReloadConfig(); err != nil {
		log.Printf("[ERROR] 重新加载配置失败 - IP: %s, Error: %v", c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to reload config: " + err.Error(),
		})
		return
	}

	log.Printf("[INFO] 配置已重新加载 - IP: %s", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "config reloaded successfully",
	})
}
