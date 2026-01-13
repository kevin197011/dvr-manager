package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	// 数据库文件名
	DBFileName = "dvr-manager.db"
)

var (
	db *sql.DB
)

// InitDB 初始化数据库
func InitDB(dataDir string) error {
	dbPath := filepath.Join(dataDir, DBFileName)

	// 检查是否存在旧的数据库文件，如果存在且不是 SQLite 格式，则删除
	if _, err := os.Stat(dbPath); err == nil {
		// 文件存在，检查是否是有效的 SQLite 文件
		file, err := os.Open(dbPath)
		if err == nil {
			header := make([]byte, 16)
			if n, err := file.Read(header); err == nil && n == 16 {
				// SQLite 文件以 "SQLite format 3" 开头
				sqliteHeader := "SQLite format 3\x00"
				if string(header) != sqliteHeader {
					// 不是 SQLite 格式，删除旧文件
					file.Close()
					log.Printf("[WARN] 检测到旧的数据库文件，正在删除: %s", dbPath)
					if err := os.Remove(dbPath); err != nil {
						log.Printf("[WARN] 删除旧数据库文件失败: %v", err)
					}
				} else {
					file.Close()
				}
			} else {
				file.Close()
			}
		}
	}

	var err error
	// 使用 SQLite 连接参数优化，避免锁定问题
	db, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=1")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接参数
	db.SetMaxOpenConns(1) // SQLite 只支持单连接写入
	db.SetMaxIdleConns(1)

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// 创建表结构
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Printf("[INFO] 数据库初始化成功: %s", dbPath)
	return nil
}

// createTables 创建数据库表
func createTables() error {
	queries := []string{
		// 配置表
		`CREATE TABLE IF NOT EXISTS config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// DVR 服务器表
		`CREATE TABLE IF NOT EXISTS dvr_servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			server TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// 创建索引
		`CREATE INDEX IF NOT EXISTS idx_config_key ON config(key)`,
		`CREATE INDEX IF NOT EXISTS idx_dvr_servers_server ON dvr_servers(server)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *sql.DB {
	return db
}

// Close 关闭数据库
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
