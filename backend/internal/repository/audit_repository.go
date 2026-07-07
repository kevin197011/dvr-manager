package repository

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"dvr-manager/internal/audit"
	"dvr-manager/pkg/db"
)

// AuditEntry 单条审计记录
type AuditEntry struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Action    string    `json:"action"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	ClientIP  string    `json:"client_ip"`
	Resource  string    `json:"resource"`
	Detail    string    `json:"detail"`
	Status    string    `json:"status"`
}

// DashboardDayStat 按日统计
type DashboardDayStat struct {
	Date          string `json:"date"`
	QuerySingle   int    `json:"query_single"`
	QueryBatch    int    `json:"query_batch"`
	Stream        int    `json:"stream"`
	QuerySuccess  int    `json:"query_success"`
	QueryFail     int    `json:"query_fail"`
	StreamSuccess int    `json:"stream_success"`
	StreamFail    int    `json:"stream_fail"`
}

// DashboardSummary 区间汇总
type DashboardSummary struct {
	QuerySingle       int     `json:"query_single"`
	QueryBatch        int     `json:"query_batch"`
	QueryBatchRecords int     `json:"query_batch_records"`
	Stream            int     `json:"stream"`
	QuerySuccessRate  float64 `json:"query_success_rate"`
	StreamSuccessRate float64 `json:"stream_success_rate"`
	ActiveUsers       int     `json:"active_users"`
	LoginSuccess      int     `json:"login_success"`
}

// DashboardStats 使用统计
type DashboardStats struct {
	Series    []DashboardDayStat `json:"series"`
	Summary   DashboardSummary   `json:"summary"`
	Truncated bool               `json:"truncated,omitempty"`
}

// AuditRepository 审计仓库接口
type AuditRepository interface {
	Insert(action, username, role, clientIP, resource, detail, status string) error
	List(from, to *time.Time, action, username string, page, pageSize int) ([]AuditEntry, int, error)
	DeleteOlderThan(t time.Time) (int64, error)
	Stats(from, to time.Time) (*DashboardStats, error)
}

type auditRepository struct {
	db *sql.DB
}

var batchDetailRe = regexp.MustCompile(`^批量查询 (\d+) 条`)

// NewAuditRepository 创建审计仓库
func NewAuditRepository() AuditRepository {
	return &auditRepository{db: db.GetDB()}
}

// Insert 写入单条审计记录
func (r *auditRepository) Insert(action, username, role, clientIP, resource, detail, status string) error {
	_, err := r.db.Exec(
		`INSERT INTO audit_log (action, username, role, client_ip, resource, detail, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		action, username, role, clientIP, resource, detail, status,
	)
	return err
}

// List 分页查询审计记录，仅返回 3 个月内数据；from/to 为可选筛选
func (r *auditRepository) List(from, to *time.Time, action, username string, page, pageSize int) ([]AuditEntry, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	where := "created_at >= ?"
	args := []interface{}{audit.RetentionCutoff()}
	if from != nil {
		where += " AND created_at >= ?"
		args = append(args, *from)
	}
	if to != nil {
		where += " AND created_at <= ?"
		args = append(args, *to)
	}
	if action != "" {
		where += " AND action = ?"
		args = append(args, action)
	}
	if username != "" {
		where += " AND username = ?"
		args = append(args, username)
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM audit_log WHERE " + where
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit: %w", err)
	}

	args = append(args, pageSize, offset)
	rows, err := r.db.Query(
		`SELECT id, created_at, action, username, role, client_ip, resource, detail, status
		 FROM audit_log WHERE `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit: %w", err)
	}
	defer rows.Close()

	var list []AuditEntry
	for rows.Next() {
		var e AuditEntry
		var username, role, clientIP, resource, detail, status sql.NullString
		if err := rows.Scan(&e.ID, &e.CreatedAt, &e.Action, &username, &role, &clientIP, &resource, &detail, &status); err != nil {
			return nil, 0, err
		}
		e.Username = username.String
		e.Role = role.String
		e.ClientIP = clientIP.String
		e.Resource = resource.String
		e.Detail = detail.String
		e.Status = status.String
		list = append(list, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Stats 按日聚合使用统计（from/to 含当日边界，由调用方设定）
func (r *auditRepository) Stats(from, to time.Time) (*DashboardStats, error) {
	cutoff := audit.RetentionCutoff()
	truncated := from.Before(cutoff)
	if truncated {
		from = cutoff
	}

	rows, err := r.db.Query(`
		SELECT date(created_at) AS d,
			SUM(CASE WHEN action = 'play' AND detail IN ('录像已找到', '录像未找到') THEN 1 ELSE 0 END),
			SUM(CASE WHEN action = 'play_batch' THEN 1 ELSE 0 END),
			SUM(CASE WHEN action = 'stream' OR (action = 'play' AND detail LIKE '流代理:%') THEN 1 ELSE 0 END),
			SUM(CASE WHEN action = 'play' AND detail IN ('录像已找到', '录像未找到') AND status = 'success' THEN 1 ELSE 0 END),
			SUM(CASE WHEN action = 'play' AND detail IN ('录像已找到', '录像未找到') AND status = 'fail' THEN 1 ELSE 0 END),
			SUM(CASE WHEN (action = 'stream' OR (action = 'play' AND detail LIKE '流代理:%')) AND status = 'success' THEN 1 ELSE 0 END),
			SUM(CASE WHEN (action = 'stream' OR (action = 'play' AND detail LIKE '流代理:%')) AND status = 'fail' THEN 1 ELSE 0 END)
		FROM audit_log
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY date(created_at)
		ORDER BY d`,
		from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("stats series: %w", err)
	}
	defer rows.Close()

	byDate := map[string]DashboardDayStat{}
	for rows.Next() {
		var day string
		var s DashboardDayStat
		if err := rows.Scan(&day, &s.QuerySingle, &s.QueryBatch, &s.Stream,
			&s.QuerySuccess, &s.QueryFail, &s.StreamSuccess, &s.StreamFail); err != nil {
			return nil, err
		}
		s.Date = day
		byDate[day] = s
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	series := fillDaySeries(from, to, byDate)

	var summary DashboardSummary
	for _, s := range series {
		summary.QuerySingle += s.QuerySingle
		summary.QueryBatch += s.QueryBatch
		summary.Stream += s.Stream
	}
	summary.QuerySuccessRate = rate(sumField(series, func(s DashboardDayStat) int { return s.QuerySuccess }),
		sumField(series, func(s DashboardDayStat) int { return s.QuerySuccess + s.QueryFail }))
	summary.StreamSuccessRate = rate(sumField(series, func(s DashboardDayStat) int { return s.StreamSuccess }),
		sumField(series, func(s DashboardDayStat) int { return s.StreamSuccess + s.StreamFail }))

	if err := r.db.QueryRow(`
		SELECT COUNT(DISTINCT username) FROM audit_log
		WHERE created_at >= ? AND created_at <= ?
		  AND username IS NOT NULL AND username != ''
		  AND (
		    (action = 'play' AND detail IN ('录像已找到', '录像未找到'))
		    OR action = 'play_batch'
		    OR action = 'stream'
		    OR (action = 'play' AND detail LIKE '流代理:%')
		  )`, from, to).Scan(&summary.ActiveUsers); err != nil {
		return nil, fmt.Errorf("stats active users: %w", err)
	}

	if err := r.db.QueryRow(`
		SELECT COUNT(*) FROM audit_log
		WHERE created_at >= ? AND created_at <= ? AND action = 'login_success'`,
		from, to).Scan(&summary.LoginSuccess); err != nil {
		return nil, fmt.Errorf("stats login: %w", err)
	}

	batchRecords, err := r.sumBatchRecords(from, to)
	if err != nil {
		return nil, err
	}
	summary.QueryBatchRecords = batchRecords

	return &DashboardStats{
		Series:    series,
		Summary:   summary,
		Truncated: truncated,
	}, nil
}

func (r *auditRepository) sumBatchRecords(from, to time.Time) (int, error) {
	rows, err := r.db.Query(
		`SELECT detail FROM audit_log WHERE created_at >= ? AND created_at <= ? AND action = 'play_batch'`,
		from, to,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var detail sql.NullString
		if err := rows.Scan(&detail); err != nil {
			return 0, err
		}
		if m := batchDetailRe.FindStringSubmatch(detail.String); len(m) == 2 {
			if n, err := strconv.Atoi(m[1]); err == nil {
				total += n
			}
		}
	}
	return total, rows.Err()
}

func fillDaySeries(from, to time.Time, byDate map[string]DashboardDayStat) []DashboardDayStat {
	loc := from.Location()
	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, loc)
	end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, loc)

	var series []DashboardDayStat
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		if s, ok := byDate[key]; ok {
			series = append(series, s)
		} else {
			series = append(series, DashboardDayStat{Date: key})
		}
	}
	return series
}

func sumField(series []DashboardDayStat, fn func(DashboardDayStat) int) int {
	n := 0
	for _, s := range series {
		n += fn(s)
	}
	return n
}

func rate(success, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(success) / float64(total)
}

// DeleteOlderThan 删除早于 t 的记录，返回删除条数
func (r *auditRepository) DeleteOlderThan(t time.Time) (int64, error) {
	res, err := r.db.Exec("DELETE FROM audit_log WHERE created_at < ?", t)
	if err != nil {
		return 0, fmt.Errorf("delete old audit: %w", err)
	}
	return res.RowsAffected()
}
