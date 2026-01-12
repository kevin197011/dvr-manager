package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"dvr-vod-system/internal/config"
	"dvr-vod-system/pkg/db"
)

// ConfigRepository 配置仓库接口
type ConfigRepository interface {
	GetConfig() (*config.Config, error)
	SaveConfig(cfg *config.Config) error
	GetDefaultConfig() *config.Config
}

// configRepository 配置仓库实现
type configRepository struct {
	db *sql.DB
}

// NewConfigRepository 创建新的配置仓库
func NewConfigRepository() ConfigRepository {
	return &configRepository{
		db: db.GetDB(),
	}
}

// GetDefaultConfig 获取默认配置
func (r *configRepository) GetDefaultConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port:    8080,
			Timeout: 30 * time.Second,
		},
		DVR: config.DVRConfig{
			Timeout:       10 * time.Second,
			Retry:         3,
			SkipTLSVerify: true,
		},
		DVRServers: []string{},
		CORS: config.CORSConfig{
			Enabled:      true,
			AllowOrigins: "*",
			AllowMethods: "POST, GET, OPTIONS",
			AllowHeaders: "Content-Type",
		},
	}
}

// GetConfig 从数据库获取配置
func (r *configRepository) GetConfig() (*config.Config, error) {
	var configJSON string
	err := r.db.QueryRow("SELECT value FROM config WHERE key = ?", "main").Scan(&configJSON)

	if err == sql.ErrNoRows {
		// 配置不存在，返回默认配置
		return r.GetDefaultConfig(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	var cfg config.Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		log.Printf("[WARN] 解析配置失败，使用默认配置: %v", err)
		return r.GetDefaultConfig(), nil
	}

	// 应用默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Timeout == 0 {
		cfg.Server.Timeout = 30 * time.Second
	}
	if cfg.DVR.Timeout == 0 {
		cfg.DVR.Timeout = 10 * time.Second
	}
	if cfg.DVR.Retry == 0 {
		cfg.DVR.Retry = 3
	}
	if cfg.CORS.AllowOrigins == "" {
		cfg.CORS.Enabled = true
		cfg.CORS.AllowOrigins = "*"
		cfg.CORS.AllowMethods = "POST, GET, OPTIONS"
		cfg.CORS.AllowHeaders = "Content-Type"
	}

	return &cfg, nil
}

// SaveConfig 保存配置到数据库
func (r *configRepository) SaveConfig(cfg *config.Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 使用 INSERT OR REPLACE 实现 upsert
	_, err = r.db.Exec(
		"INSERT OR REPLACE INTO config (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)",
		"main",
		string(data),
	)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	log.Printf("[INFO] 配置已保存到数据库")
	return nil
}
