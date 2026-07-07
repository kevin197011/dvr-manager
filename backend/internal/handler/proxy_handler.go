package handler

import (
	"log"
	"net/http"
	"strings"

	"dvr-vod-system/internal/repository"
	"dvr-vod-system/internal/service"
	"dvr-vod-system/pkg/cache"

	"github.com/gin-gonic/gin"
)

// ProxyHandler 代理处理器
type ProxyHandler struct {
	proxyService service.ProxyService
	dvrService   service.DVRService
	cache        cache.Cache
	auditRepo    repository.AuditRepository
}

// NewProxyHandler 创建新的代理处理器
func NewProxyHandler(proxyService service.ProxyService, dvrService service.DVRService, cache cache.Cache, auditRepo repository.AuditRepository) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
		dvrService:   dvrService,
		cache:        cache,
		auditRepo:    auditRepo,
	}
}

// Handle 处理视频流代理请求
// 直接 GET /stream/<recordID>.mp4 即可触发 DVR 查询并代理播放（无需先调用 /play）
func (h *ProxyHandler) Handle(c *gin.Context) {
	filename := c.Param("filename")
	recordID := strings.TrimSuffix(filename, ".mp4")
	clientIP := c.ClientIP()

	log.Printf("[INFO] 流代理请求 - IP: %s, 编号: %s", clientIP, recordID)

	// 优先从缓存获取真实 URL，避免重复 DVR 查询
	realURL, exists := h.cache.Get(recordID)

	// 缓存未命中：直接到 DVR 服务器查询
	if !exists {
		if h.dvrService == nil {
			log.Printf("[WARN] 流代理失败 - 编号: %s, 原因: dvrService 未初始化", recordID)
			c.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
			return
		}

		log.Printf("[INFO] 缓存未命中，直接查询 DVR - 编号: %s", recordID)
		username, _ := c.Get("username")
		userStr, _ := username.(string)
		role, _ := c.Get("role")
		roleStr, _ := role.(string)

		url, err := h.dvrService.FindRecording(c.Request.Context(), recordID)
		if err != nil {
			if h.auditRepo != nil {
				_ = h.auditRepo.Insert("play", userStr, roleStr, clientIP, recordID, "流代理: 录像未找到", "fail")
			}
			log.Printf("[WARN] 流代理失败 - 编号: %s, Error: %v", recordID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "recording not found"})
			return
		}

		realURL = url
		h.cache.Set(recordID, realURL)
		if h.auditRepo != nil {
			_ = h.auditRepo.Insert("play", userStr, roleStr, clientIP, recordID, "流代理: 录像已找到", "success")
		}
	}

	rangeHeader := c.GetHeader("Range")

	if err := h.proxyService.ProxyStream(c.Request.Context(), recordID, realURL, c.Writer, rangeHeader); err != nil {
		log.Printf("[ERROR] 流代理失败 - 编号: %s, Error: %v", recordID, err)
		if !c.Writer.Written() {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch video from DVR server"})
		}
		return
	}
}
