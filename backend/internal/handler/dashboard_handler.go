package handler

import (
	"net/http"
	"time"

	"dvr-manager/internal/audit"
	"dvr-manager/internal/repository"

	"github.com/gin-gonic/gin"
)

const maxDashboardRangeDays = 90

// DashboardHandler 使用统计
type DashboardHandler struct {
	auditRepo repository.AuditRepository
}

// NewDashboardHandler 创建 Dashboard 处理器
func NewDashboardHandler(auditRepo repository.AuditRepository) *DashboardHandler {
	return &DashboardHandler{auditRepo: auditRepo}
}

type dashboardQuery struct {
	From        string `form:"from"`
	To          string `form:"to"`
	Granularity string `form:"granularity"`
}

// GetStats 使用统计（按日聚合）
func (h *DashboardHandler) GetStats(c *gin.Context) {
	var q dashboardQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid query"})
		return
	}
	if q.Granularity != "" && q.Granularity != "day" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "only granularity=day is supported"})
		return
	}

	loc := time.Local
	now := time.Now().In(loc)
	toDay := now
	fromDay := now.AddDate(0, 0, -29)

	if q.To != "" {
		t, err := time.ParseInLocation("2006-01-02", q.To, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid to date, use YYYY-MM-DD"})
			return
		}
		toDay = t
	}
	if q.From != "" {
		t, err := time.ParseInLocation("2006-01-02", q.From, loc)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid from date, use YYYY-MM-DD"})
			return
		}
		fromDay = t
	}

	if fromDay.After(toDay) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "from must be before or equal to to"})
		return
	}

	days := int(toDay.Sub(fromDay).Hours()/24) + 1
	if days > maxDashboardRangeDays {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "date range exceeds 90 days"})
		return
	}

	from := time.Date(fromDay.Year(), fromDay.Month(), fromDay.Day(), 0, 0, 0, 0, loc)
	to := time.Date(toDay.Year(), toDay.Month(), toDay.Day(), 23, 59, 59, 999999999, loc)

	if from.Before(audit.RetentionCutoff()) {
		from = audit.RetentionCutoff()
	}

	stats, err := h.auditRepo.Stats(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"range": gin.H{
			"from":        fromDay.Format("2006-01-02"),
			"to":          toDay.Format("2006-01-02"),
			"granularity": "day",
		},
		"summary":   stats.Summary,
		"series":    stats.Series,
		"truncated": stats.Truncated,
	})
}
