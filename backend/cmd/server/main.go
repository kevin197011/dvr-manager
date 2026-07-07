package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"dvr-vod-system/internal/audit"
	"dvr-vod-system/internal/auth"
	"dvr-vod-system/internal/config"
	"dvr-vod-system/internal/repository"
	"dvr-vod-system/internal/router"
	"dvr-vod-system/pkg/db"
)

func main() {
	dataDir := "../data"
	if dataDirEnv := os.Getenv("DATA_DIR"); dataDirEnv != "" {
		dataDir = dataDirEnv
	} else if _, err := os.Stat("/app/data"); err == nil {
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
	go runAuditDailyCleanup(auditRepo)

	cacheTTLDays := recordingCacheTTLDays()
	recordingCacheRepo := repository.NewRecordingCacheRepository()
	if n, err := recordingCacheRepo.DeleteExpired(time.Now()); err != nil {
		log.Printf("[RecordingCache] startup cleanup warning: %v", err)
	} else if n > 0 {
		log.Printf("[RecordingCache] startup cleanup: removed %d expired entries", n)
	}
	go runRecordingCacheDailyCleanup(recordingCacheRepo)

	configRepo := repository.NewConfigRepository()
	dvrRepo := repository.NewDVRRepository()
	cfg, err := configRepo.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config from database: %v", err)
	}
	if servers, err := dvrRepo.GetAll(); err == nil && len(servers) > 0 {
		cfg.DVRServers = servers
	}
	config.SetConfig(cfg)

	jwt := auth.NewJWT(os.Getenv("JWT_SECRET"))

	log.Printf("Loaded %d DVR servers from database", len(cfg.DVRServers))
	log.Printf("Recording cache TTL: %d days (RECORD_CACHE_TTL_DAYS)", cacheTTLDays)
	if config.RequireAuthForPlayEnabled() {
		log.Printf("Play/stream endpoints require authentication (REQUIRE_AUTH_FOR_PLAY)")
	}

	r := router.NewRouter(cfg, cacheTTLDays, jwt)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	readTimeout := cfg.Server.Timeout
	if readTimeout <= 0 {
		readTimeout = 30 * time.Second
	}

	log.Printf("========================================")
	log.Printf("DVR VOD System starting...")
	log.Printf("Server Address: http://localhost%s", addr)
	log.Printf("DVR Servers: %d", len(cfg.DVRServers))
	for i, server := range cfg.DVRServers {
		log.Printf("  [%d] %s", i+1, server)
	}
	log.Printf("========================================")

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: readTimeout,
		ReadTimeout:       readTimeout,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}

func runAuditDailyCleanup(auditRepo repository.AuditRepository) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		d := time.Until(next)
		if d < 0 {
			d = 0
		}
		time.Sleep(d)
		cutoff := audit.RetentionCutoff()
		n, err := auditRepo.DeleteOlderThan(cutoff)
		if err != nil {
			log.Printf("[Audit] daily cleanup error: %v", err)
		} else {
			log.Printf("[Audit] daily cleanup: deleted=%d cutoff_before=%s",
				n, cutoff.Format("2006-01-02"))
		}
	}
}

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
