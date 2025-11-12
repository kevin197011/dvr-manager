package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// URL 缓存，用于存储录像编号到真实 URL 的映射
var (
	urlCache      = make(map[string]string)
	urlCacheMutex sync.RWMutex
)

// CacheRecordingURL 缓存录像 URL
func CacheRecordingURL(recordID, url string) {
	urlCacheMutex.Lock()
	defer urlCacheMutex.Unlock()
	urlCache[recordID] = url
}

// GetCachedURL 获取缓存的 URL
func GetCachedURL(recordID string) (string, bool) {
	urlCacheMutex.RLock()
	defer urlCacheMutex.RUnlock()
	url, exists := urlCache[recordID]
	return url, exists
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
	Success bool                `json:"success"`
	Results []RecordingResult   `json:"results"`
	Message string              `json:"message,omitempty"`
}

// RecordingResult 单个录像查询结果
type RecordingResult struct {
	RecordID string `json:"record_id"`
	Found    bool   `json:"found"`
	ProxyURL string `json:"proxy_url,omitempty"`
}

// PlayHandler 处理播放请求（单个录像）
func PlayHandler(c *gin.Context) {
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
		BatchPlayHandler(c, req.RecordIDs)
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
	url, err := FindRecording(ctx, req.RecordID)
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
	CacheRecordingURL(req.RecordID, url)
	
	log.Printf("[SUCCESS] 录像找到 - IP: %s, 编号: %s", c.ClientIP(), req.RecordID)
	c.JSON(http.StatusOK, PlayResponse{
		Success:  true,
		ProxyURL: proxyURL,
		Message:  "recording found",
	})
}

// BatchPlayHandler 处理批量播放请求
func BatchPlayHandler(c *gin.Context, recordIDs []string) {
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

		url, err := FindRecording(ctx, recordID)
		result := RecordingResult{
			RecordID: recordID,
			Found:    err == nil,
		}
		if err == nil {
			// 生成代理 URL
			proxyURL := fmt.Sprintf("/stream/%s.mp4", recordID)
			result.ProxyURL = proxyURL
			
			// 缓存真实 URL 映射
			CacheRecordingURL(recordID, url)
			
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
