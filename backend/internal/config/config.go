package config

import (
	"time"
)

// Config 配置结构
type Config struct {
	Server     ServerConfig `json:"server"`
	DVR        DVRConfig    `json:"dvr"`
	DVRServers []string     `json:"dvr_servers"`
	CORS       CORSConfig   `json:"cors"`
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

var globalConfig *Config

// SetConfig 设置全局配置
func SetConfig(cfg *Config) {
	globalConfig = cfg
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return globalConfig
}
