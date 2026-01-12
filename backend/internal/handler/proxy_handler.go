package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"dvr-vod-system/internal/service"
	"dvr-vod-system/pkg/cache"
)

// ProxyHandler 代理处理器
type ProxyHandler struct {
	proxyService service.ProxyService
	cache        cache.Cache
}

// NewProxyHandler 创建新的代理处理器
func NewProxyHandler(proxyService service.ProxyService, cache cache.Cache) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
		cache:        cache,
	}
}

// Handle 处理视频流代理请求
func (h *ProxyHandler) Handle(c *gin.Context) {
	// 从路径中提取录像编号
	filename := c.Param("filename")
	recordID := strings.TrimSuffix(filename, ".mp4")

	log.Printf("[INFO] 流代理请求 - IP: %s, 编号: %s", c.ClientIP(), recordID)

	// 从缓存获取真实 URL
	realURL, exists := h.cache.Get(recordID)
	if !exists {
		log.Printf("[WARN] 流代理失败 - 编号: %s, 原因: URL 未缓存", recordID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "recording not found or expired",
		})
		return
	}

	// 获取 Range 请求头
	rangeHeader := c.GetHeader("Range")

	// 代理视频流
	if err := h.proxyService.ProxyStream(c.Request.Context(), recordID, realURL, c.Writer, rangeHeader); err != nil {
		log.Printf("[ERROR] 流代理失败 - 编号: %s, Error: %v", recordID, err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "failed to fetch video from DVR server",
		})
		return
	}
}
