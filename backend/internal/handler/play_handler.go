package handler

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"dvr-vod-system/internal/repository"
	"dvr-vod-system/internal/service"
	"dvr-vod-system/pkg/cache"

	"github.com/gin-gonic/gin"
)

const (
	maxBatchPlaySize = 50
	batchPlayWorkers = 5
)

// PlayHandler 播放处理器
type PlayHandler struct {
	dvrService service.DVRService
	cache      cache.Cache
	auditRepo  repository.AuditRepository
}

// NewPlayHandler 创建播放处理器
func NewPlayHandler(dvrService service.DVRService, cache cache.Cache, auditRepo repository.AuditRepository) *PlayHandler {
	return &PlayHandler{
		dvrService: dvrService,
		cache:      cache,
		auditRepo:  auditRepo,
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

// Handle 处理播放请求
func (h *PlayHandler) Handle(c *gin.Context) {
	var req PlayRequest

	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, PlayResponse{Success: false, Message: "invalid request: " + err.Error()})
			return
		}
	} else {
		req.RecordID = c.PostForm("record_id")
		if req.RecordID == "" {
			req.RecordID = c.Query("record_id")
		}
	}

	if len(req.RecordIDs) > 0 {
		h.HandleBatch(c, req.RecordIDs)
		return
	}

	if strings.TrimSpace(req.RecordID) == "" {
		c.JSON(http.StatusBadRequest, PlayResponse{Success: false, Message: "record_id is required"})
		return
	}

	h.handleSingle(c, strings.TrimSpace(req.RecordID))
}

func (h *PlayHandler) handleSingle(c *gin.Context, recordID string) {
	ctx := c.Request.Context()
	userStr, roleStr := playActor(c)

	url, err := h.dvrService.FindRecording(ctx, recordID)
	if err != nil {
		h.auditPlay(c, userStr, roleStr, recordID, "录像未找到", "fail")
		c.JSON(http.StatusNotFound, PlayResponse{Success: false, Message: "recording not found"})
		return
	}

	proxyURL := fmt.Sprintf("/stream/%s.mp4", recordID)
	h.cache.Set(recordID, url)
	h.auditPlay(c, userStr, roleStr, recordID, "录像已找到", "success")

	c.JSON(http.StatusOK, PlayResponse{
		Success:  true,
		ProxyURL: proxyURL,
		Message:  "recording found",
	})
}

// HandleBatch 批量查询（有限并发）
func (h *PlayHandler) HandleBatch(c *gin.Context, recordIDs []string) {
	if len(recordIDs) > maxBatchPlaySize {
		c.JSON(http.StatusBadRequest, BatchPlayResponse{
			Success: false,
			Message: fmt.Sprintf("batch size exceeds limit of %d", maxBatchPlaySize),
		})
		return
	}

	ctx := c.Request.Context()
	userStr, roleStr := playActor(c)
	results := make([]RecordingResult, len(recordIDs))

	sem := make(chan struct{}, batchPlayWorkers)
	var wg sync.WaitGroup

	for i, id := range recordIDs {
		recordID := strings.TrimSpace(id)
		results[i].RecordID = recordID
		if recordID == "" {
			continue
		}

		wg.Add(1)
		go func(idx int, rid string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				return
			}

			url, err := h.dvrService.FindRecording(ctx, rid)
			if err != nil {
				return
			}
			proxyURL := fmt.Sprintf("/stream/%s.mp4", rid)
			h.cache.Set(rid, url)
			results[idx] = RecordingResult{RecordID: rid, Found: true, ProxyURL: proxyURL}
		}(i, recordID)
	}

	wg.Wait()

	foundCount := 0
	for _, r := range results {
		if r.Found {
			foundCount++
		}
	}

	if h.auditRepo != nil {
		detail := fmt.Sprintf("批量查询 %d 条，找到 %d 条", len(recordIDs), foundCount)
		_ = h.auditRepo.Insert("play_batch", userStr, roleStr, c.ClientIP(), "", detail, "success")
	}

	c.JSON(http.StatusOK, BatchPlayResponse{
		Success: true,
		Results: results,
		Message: "batch query completed",
	})
}

func playActor(c *gin.Context) (username, role string) {
	u, _ := c.Get("username")
	username, _ = u.(string)
	r, _ := c.Get("role")
	role, _ = r.(string)
	return username, role
}

func (h *PlayHandler) auditPlay(c *gin.Context, user, role, recordID, detail, status string) {
	if h.auditRepo == nil {
		return
	}
	_ = h.auditRepo.Insert("play", user, role, c.ClientIP(), recordID, detail, status)
}
