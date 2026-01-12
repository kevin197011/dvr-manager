package repository

import (
	"database/sql"
	"fmt"
	"log"

	"dvr-vod-system/pkg/db"
)

// DVRRepository DVR 服务器仓库接口
type DVRRepository interface {
	GetAll() ([]string, error)
	Save(servers []string) error
	Add(server string) error
	Delete(server string) error
}

// dvrRepository DVR 服务器仓库实现
type dvrRepository struct {
	db *sql.DB
}

// NewDVRRepository 创建新的 DVR 仓库
func NewDVRRepository() DVRRepository {
	return &dvrRepository{
		db: db.GetDB(),
	}
}

// GetAll 获取所有 DVR 服务器
func (r *dvrRepository) GetAll() ([]string, error) {
	rows, err := r.db.Query("SELECT server FROM dvr_servers ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to query DVR servers: %w", err)
	}
	defer rows.Close()

	var servers []string
	for rows.Next() {
		var server string
		if err := rows.Scan(&server); err != nil {
			log.Printf("[WARN] 扫描服务器数据失败: %v", err)
			continue
		}
		servers = append(servers, server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return servers, nil
}

// Save 保存 DVR 服务器列表（替换所有）
func (r *dvrRepository) Save(servers []string) error {
	// 开始事务
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 删除所有现有数据
	_, err = tx.Exec("DELETE FROM dvr_servers")
	if err != nil {
		return fmt.Errorf("failed to clear existing servers: %w", err)
	}

	// 插入新数据
	stmt, err := tx.Prepare("INSERT INTO dvr_servers (server) VALUES (?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, server := range servers {
		if _, err := stmt.Exec(server); err != nil {
			return fmt.Errorf("failed to insert server: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[INFO] 保存了 %d 个 DVR 服务器到数据库", len(servers))
	return nil
}

// Add 添加单个 DVR 服务器
func (r *dvrRepository) Add(server string) error {
	// 检查是否已存在
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM dvr_servers WHERE server = ?", server).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing server: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("server already exists: %s", server)
	}

	// 插入新服务器
	_, err = r.db.Exec("INSERT INTO dvr_servers (server) VALUES (?)", server)
	if err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	log.Printf("[INFO] 添加 DVR 服务器到数据库: %s", server)
	return nil
}

// Delete 删除 DVR 服务器
func (r *dvrRepository) Delete(server string) error {
	result, err := r.db.Exec("DELETE FROM dvr_servers WHERE server = ?", server)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server not found: %s", server)
	}

	log.Printf("[INFO] 从数据库删除 DVR 服务器: %s", server)
	return nil
}
