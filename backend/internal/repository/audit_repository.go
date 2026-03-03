package repository

import (
	"database/sql"
	"fmt"
	"time"

	"dvr-vod-system/pkg/db"
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

// AuditRepository 审计仓库接口
type AuditRepository interface {
	Insert(action, username, role, clientIP, resource, detail, status string) error
	List(from, to *time.Time, action, username string, page, pageSize int) ([]AuditEntry, int, error)
	DeleteOlderThan(t time.Time) (int64, error)
}

type auditRepository struct {
	db *sql.DB
}

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

	// 3 个月截止时间
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)

	where := "created_at >= ?"
	args := []interface{}{threeMonthsAgo}
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

	// 总数
	var total int
	countQuery := "SELECT COUNT(*) FROM audit_log WHERE " + where
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit: %w", err)
	}

	// 分页列表
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

// DeleteOlderThan 删除早于 t 的记录，返回删除条数
func (r *auditRepository) DeleteOlderThan(t time.Time) (int64, error) {
	res, err := r.db.Exec("DELETE FROM audit_log WHERE created_at < ?", t)
	if err != nil {
		return 0, fmt.Errorf("delete old audit: %w", err)
	}
	return res.RowsAffected()
}
