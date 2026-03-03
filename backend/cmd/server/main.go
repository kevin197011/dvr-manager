package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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

	// 审计日志：默认保留 3 个月，硬删除，避免数据过大影响性能
	const auditRetentionMonths = 3
	auditRepo := repository.NewAuditRepository()
	auditCutoff := time.Now().AddDate(0, -auditRetentionMonths, 0)
	if n, err := auditRepo.DeleteOlderThan(auditCutoff); err != nil {
		log.Printf("[Audit] startup cleanup warning: %v", err)
	} else if n > 0 {
		log.Printf("[Audit] startup cleanup: removed %d entries older than %d months", n, auditRetentionMonths)
	}

	// 定时任务：每日午夜执行审计日志巡检（删除超过 3 个月的记录）
	go runAuditDailyCleanup(auditRepo, auditRetentionMonths)

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

// runAuditDailyCleanup 每日午夜执行审计日志清理，仅保留 retentionMonths 个月内数据（硬删除）
func runAuditDailyCleanup(auditRepo repository.AuditRepository, retentionMonths int) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		d := time.Until(next)
		if d < 0 {
			d = 0
		}
		time.Sleep(d)
		cutoff := time.Now().AddDate(0, -retentionMonths, 0)
		n, err := auditRepo.DeleteOlderThan(cutoff)
		if err != nil {
			log.Printf("[Audit] daily cleanup error: %v", err)
		} else if n > 0 {
			log.Printf("[Audit] daily cleanup: removed %d entries older than %d months", n, retentionMonths)
		}
	}
}
