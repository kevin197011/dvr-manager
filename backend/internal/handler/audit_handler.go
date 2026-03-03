package handler

import (
	"net/http"
	"time"

	"dvr-vod-system/internal/repository"

	"github.com/gin-gonic/gin"
)

// AuditHandler 审计查询处理器
type AuditHandler struct {
	auditRepo repository.AuditRepository
}

// NewAuditHandler 创建审计处理器
func NewAuditHandler(auditRepo repository.AuditRepository) *AuditHandler {
	return &AuditHandler{auditRepo: auditRepo}
}

// ListQuery 查询参数
type ListQuery struct {
	From     string `form:"from"`     // 开始时间 RFC3339
	To       string `form:"to"`       // 结束时间 RFC3339
	Action   string `form:"action"`   // 动作类型
	Username string `form:"username"` // 用户名
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// GetAudit 获取审计日志列表（分页，仅 3 个月内）
func (h *AuditHandler) GetAudit(c *gin.Context) {
	var q ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid query"})
		return
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}

	var from, to *time.Time
	if q.From != "" {
		t, err := time.Parse(time.RFC3339, q.From)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid from time"})
			return
		}
		from = &t
	}
	if q.To != "" {
		t, err := time.Parse(time.RFC3339, q.To)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid to time"})
			return
		}
		to = &t
	}

	list, total, err := h.auditRepo.List(from, to, q.Action, q.Username, q.Page, q.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"list":      list,
		"total":     total,
		"page":      q.Page,
		"page_size": q.PageSize,
	})
}

// Cleanup 清理超过 3 个月的审计记录（仅管理员）
func (h *AuditHandler) Cleanup(c *gin.Context) {
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	n, err := h.auditRepo.DeleteOlderThan(threeMonthsAgo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "cleanup done",
		"deleted": n,
	})
}
