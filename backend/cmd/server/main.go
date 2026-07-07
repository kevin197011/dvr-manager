package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"dvr-vod-system/internal/audit"
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

	// 审计日志：默认保留 3 个月，硬删除；启动时 + 每日 00:00 后台清理
	auditRetentionMonths := audit.RetentionMonths()
	auditRepo := repository.NewAuditRepository()
	log.Printf("Audit log retention: %d months (AUDIT_RETENTION_MONTHS); startup + daily 00:00 cleanup enabled",
		auditRetentionMonths)
	if n, err := auditRepo.DeleteOlderThan(audit.RetentionCutoff()); err != nil {
		log.Printf("[Audit] startup cleanup warning: %v", err)
	} else {
		log.Printf("[Audit] startup cleanup: deleted=%d cutoff_before=%s",
			n, audit.RetentionCutoff().Format("2006-01-02"))
	}
	go runAuditDailyCleanup(auditRepo, auditRetentionMonths)

	// 录像 URL 缓存：默认保留 30 天，可通过 RECORD_CACHE_TTL_DAYS 调整
	cacheTTLDays := recordingCacheTTLDays()
	recordingCacheRepo := repository.NewRecordingCacheRepository()
	if n, err := recordingCacheRepo.DeleteExpired(time.Now()); err != nil {
		log.Printf("[RecordingCache] startup cleanup warning: %v", err)
	} else if n > 0 {
		log.Printf("[RecordingCache] startup cleanup: removed %d expired entries", n)
	}
	go runRecordingCacheDailyCleanup(recordingCacheRepo)

	// 从数据库加载配置
	configRepo := repository.NewConfigRepository()
	cfg, err := configRepo.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config from database: %v", err)
	}

	// 设置全局配置
	config.SetConfig(cfg)

	log.Printf("Loaded %d DVR servers from database", len(cfg.DVRServers))
	log.Printf("Recording cache TTL: %d days (RECORD_CACHE_TTL_DAYS)", cacheTTLDays)

	// 创建路由
	r := router.NewRouter(cfg, cacheTTLDays)

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
		} else {
			log.Printf("[Audit] daily cleanup: deleted=%d retention_months=%d cutoff_before=%s",
				n, retentionMonths, cutoff.Format("2006-01-02"))
		}
	}
}

// recordingCacheTTLDays 从环境变量读取缓存保留天数，默认 30
func recordingCacheTTLDays() int {
	const defaultDays = 30
	if s := os.Getenv("RECORD_CACHE_TTL_DAYS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
		log.Printf("[RecordingCache] invalid RECORD_CACHE_TTL_DAYS=%q, using default %d", s, defaultDays)
	}
	return defaultDays
}

// runRecordingCacheDailyCleanup 每日午夜删除已过期的录像 URL 缓存
func runRecordingCacheDailyCleanup(repo repository.RecordingCacheRepository) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		d := time.Until(next)
		if d < 0 {
			d = 0
		}
		time.Sleep(d)
		n, err := repo.DeleteExpired(time.Now())
		if err != nil {
			log.Printf("[RecordingCache] daily cleanup error: %v", err)
		} else if n > 0 {
			log.Printf("[RecordingCache] daily cleanup: removed %d expired entries", n)
		}
	}
}
