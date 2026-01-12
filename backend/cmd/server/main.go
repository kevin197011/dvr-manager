package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dvr-vod-system/internal/config"
	"dvr-vod-system/internal/repository"
	"dvr-vod-system/internal/router"
	"dvr-vod-system/pkg/db"
)

func main() {
	// 初始化数据库
	// 数据目录：本地开发时使用 ../data，Docker 中使用环境变量或 /app/data
	dataDir := "../data"
	if dataDirEnv := os.Getenv("DATA_DIR"); dataDirEnv != "" {
		dataDir = dataDirEnv
	} else if _, err := os.Stat("/app/data"); err == nil {
		// Docker 环境，使用 /app/data
		dataDir = "/app/data"
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	if err := db.InitDB(dataDir); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Printf("Database initialized: %s", filepath.Join(dataDir, db.DBFileName))

	// 从数据库加载配置
	configRepo := repository.NewConfigRepository()
	cfg, err := configRepo.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config from database: %v", err)
	}

	// 设置全局配置
	config.SetConfig(cfg)

	log.Printf("Loaded %d DVR servers from database", len(cfg.DVRServers))

	// 创建路由
	r := router.NewRouter(cfg)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("========================================")
	log.Printf("DVR VOD System starting...")
	log.Printf("Server Address: http://localhost%s", addr)
	log.Printf("DVR Servers: %d", len(cfg.DVRServers))
	for i, server := range cfg.DVRServers {
		log.Printf("  [%d] %s", i+1, server)
	}
	log.Printf("========================================")

	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
