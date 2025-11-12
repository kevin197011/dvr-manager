package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Server     ServerConfig `yaml:"server"`
	DVR        DVRConfig    `yaml:"dvr"`
	DVRServers []string     `yaml:"dvr_servers"`
	CORS       CORSConfig   `yaml:"cors"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

// DVRConfig DVR 查询配置
type DVRConfig struct {
	Timeout        time.Duration `yaml:"timeout"`
	Retry          int           `yaml:"retry"`
	SkipTLSVerify  bool          `yaml:"skip_tls_verify"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	Enabled      bool   `yaml:"enabled"`
	AllowOrigins string `yaml:"allow_origins"`
	AllowMethods string `yaml:"allow_methods"`
	AllowHeaders string `yaml:"allow_headers"`
}

var globalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认值
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Timeout == 0 {
		config.Server.Timeout = 30 * time.Second
	}
	if config.DVR.Timeout == 0 {
		config.DVR.Timeout = 10 * time.Second
	}
	if config.DVR.Retry == 0 {
		config.DVR.Retry = 3
	}

	globalConfig = &config
	return &config, nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return globalConfig
}
