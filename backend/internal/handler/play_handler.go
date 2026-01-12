package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"dvr-vod-system/internal/service"
	"dvr-vod-system/pkg/cache"

	"github.com/gin-gonic/gin"
)

// PlayHandler 播放处理器
type PlayHandler struct {
	dvrService service.DVRService
	cache      cache.Cache
}

// NewPlayHandler 创建新的播放处理器
func NewPlayHandler(dvrService service.DVRService, cache cache.Cache) *PlayHandler {
	return &PlayHandler{
		dvrService: dvrService,
		cache:      cache,
	}
}

// PlayRequest 播放请求
type PlayRequest struct {
	RecordID  string   `json:"record_id"`
	RecordIDs []string `json:"record_ids"`
}

// PlayResponse 播放响应
type PlayResponse struct {
	Success  bool   `json:"success"`
	URL      string `json:"url,omitempty"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Message  string `json:"message,omitempty"`
}

// BatchPlayResponse 批量播放响应
type BatchPlayResponse struct {
	Success bool              `json:"success"`
	Results []RecordingResult `json:"results"`
	Message string            `json:"message,omitempty"`
}

// RecordingResult 单个录像查询结果
type RecordingResult struct {
	RecordID string `json:"record_id"`
	Found    bool   `json:"found"`
	ProxyURL string `json:"proxy_url,omitempty"`
}

// Handle 处理播放请求（单个录像）
func (h *PlayHandler) Handle(c *gin.Context) {
	var req PlayRequest

	// 支持 JSON 和表单两种方式
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[ERROR] 请求解析失败 - IP: %s, Error: %v", c.ClientIP(), err)
			c.JSON(http.StatusBadRequest, PlayResponse{
				Success: false,
				Message: "invalid request: " + err.Error(),
			})
			return
		}
	} else {
		req.RecordID = c.PostForm("record_id")
		if req.RecordID == "" {
			req.RecordID = c.Query("record_id")
		}
	}

	// 检查是否为批量请求
	if len(req.RecordIDs) > 0 {
		log.Printf("[INFO] 批量查询请求 - IP: %s, 数量: %d", c.ClientIP(), len(req.RecordIDs))
		h.HandleBatch(c, req.RecordIDs)
		return
	}

	if req.RecordID == "" {
		log.Printf("[WARN] 缺少录像编号 - IP: %s", c.ClientIP())
		c.JSON(http.StatusBadRequest, PlayResponse{
			Success: false,
			Message: "record_id is required",
		})
		return
	}

	log.Printf("[INFO] 单个查询请求 - IP: %s, 编号: %s", c.ClientIP(), req.RecordID)

	// 使用请求的上下文，每个请求独立互不干扰
	ctx := c.Request.Context()

	// 查找录像
	url, err := h.dvrService.FindRecording(ctx, req.RecordID)
	if err != nil {
		log.Printf("[WARN] 录像未找到 - IP: %s, 编号: %s", c.ClientIP(), req.RecordID)
		c.JSON(http.StatusNotFound, PlayResponse{
			Success: false,
			Message: "recording not found",
		})
		return
	}

	// 生成代理 URL，隐藏真实地址
	proxyURL := fmt.Sprintf("/stream/%s.mp4", req.RecordID)

	// 缓存真实 URL 映射
	h.cache.Set(req.RecordID, url)

	log.Printf("[SUCCESS] 录像找到 - IP: %s, 编号: %s", c.ClientIP(), req.RecordID)
	c.JSON(http.StatusOK, PlayResponse{
		Success:  true,
		ProxyURL: proxyURL,
		Message:  "recording found",
	})
}

// HandleBatch 处理批量播放请求
func (h *PlayHandler) HandleBatch(c *gin.Context, recordIDs []string) {
	ctx := c.Request.Context()
	startTime := time.Now()

	results := make([]RecordingResult, 0, len(recordIDs))
	foundCount := 0

	for i, recordID := range recordIDs {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			log.Printf("[ERROR] 批量查询超时 - IP: %s, 已处理: %d/%d", c.ClientIP(), i, len(recordIDs))
			c.JSON(http.StatusRequestTimeout, BatchPlayResponse{
				Success: false,
				Results: results,
				Message: "request timeout",
			})
			return
		default:
		}

		url, err := h.dvrService.FindRecording(ctx, recordID)
		result := RecordingResult{
			RecordID: recordID,
			Found:    err == nil,
		}
		if err == nil {
			// 生成代理 URL
			proxyURL := fmt.Sprintf("/stream/%s.mp4", recordID)
			result.ProxyURL = proxyURL

			// 缓存真实 URL 映射
			h.cache.Set(recordID, url)

			foundCount++
			log.Printf("[SUCCESS] 批量查询 [%d/%d] - 编号: %s", i+1, len(recordIDs), recordID)
		} else {
			log.Printf("[WARN] 批量查询 [%d/%d] - 编号: %s 未找到", i+1, len(recordIDs), recordID)
		}
		results = append(results, result)
	}

	duration := time.Since(startTime)
	log.Printf("[INFO] 批量查询完成 - IP: %s, 总数: %d, 找到: %d, 耗时: %v",
		c.ClientIP(), len(recordIDs), foundCount, duration)

	c.JSON(http.StatusOK, BatchPlayResponse{
		Success: true,
		Results: results,
		Message: "batch query completed",
	})
}
