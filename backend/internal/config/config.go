package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config 配置结构
type Config struct {
	Server             ServerConfig `json:"server"`
	DVR                DVRConfig    `json:"dvr"`
	DVRServers         []string     `json:"dvr_servers"`
	CORS               CORSConfig   `json:"cors"`
	RequireAuthForPlay bool         `json:"require_auth_for_play"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
}

// DVRConfig DVR 查询配置
type DVRConfig struct {
	Timeout       time.Duration `json:"timeout"`
	Retry         int           `json:"retry"`
	SkipTLSVerify bool          `json:"skip_tls_verify"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	Enabled      bool   `json:"enabled"`
	AllowOrigins string `json:"allow_origins"`
	AllowMethods string `json:"allow_methods"`
	AllowHeaders string `json:"allow_headers"`
}

var (
	mu           sync.RWMutex
	globalConfig *Config
)

// SetConfig 设置全局配置（线程安全）
func SetConfig(cfg *Config) {
	mu.Lock()
	defer mu.Unlock()
	globalConfig = cfg
}

// GetConfig 获取全局配置快照（线程安全）
func GetConfig() *Config {
	mu.RLock()
	defer mu.RUnlock()
	if globalConfig == nil {
		return nil
	}
	cp := *globalConfig
	cp.DVRServers = append([]string(nil), globalConfig.DVRServers...)
	return &cp
}

// RequireAuthForPlayEnabled 播放接口是否强制登录（环境变量优先于 DB 配置）
func RequireAuthForPlayEnabled() bool {
	if v := strings.TrimSpace(os.Getenv("REQUIRE_AUTH_FOR_PLAY")); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	cfg := GetConfig()
	return cfg != nil && cfg.RequireAuthForPlay
}
