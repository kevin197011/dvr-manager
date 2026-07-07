package repository

import (
	"database/sql"
	"time"

	"dvr-manager/pkg/db"
)

// RecordingCacheRepository 录像局号 → DVR 真实 URL 缓存
type RecordingCacheRepository interface {
	Get(recordID string) (realURL string, ok bool)
	Set(recordID, realURL string, ttlDays int) error
	DeleteExpired(before time.Time) (int64, error)
}

type recordingCacheRepository struct {
	db *sql.DB
}

// NewRecordingCacheRepository 创建录像缓存仓库
func NewRecordingCacheRepository() RecordingCacheRepository {
	return &recordingCacheRepository{db: db.GetDB()}
}

// Get 读取未过期的缓存条目
func (r *recordingCacheRepository) Get(recordID string) (string, bool) {
	var realURL string
	var expiresAt time.Time
	err := r.db.QueryRow(
		`SELECT real_url, expires_at FROM recording_cache WHERE record_id = ?`,
		recordID,
	).Scan(&realURL, &expiresAt)
	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		return "", false
	}
	if time.Now().After(expiresAt) {
		_, _ = r.db.Exec(`DELETE FROM recording_cache WHERE record_id = ?`, recordID)
		return "", false
	}
	return realURL, true
}

// Set 写入或更新缓存，expires_at = now + ttlDays
func (r *recordingCacheRepository) Set(recordID, realURL string, ttlDays int) error {
	if ttlDays <= 0 {
		ttlDays = 30
	}
	now := time.Now()
	expiresAt := now.AddDate(0, 0, ttlDays)
	_, err := r.db.Exec(
		`INSERT INTO recording_cache (record_id, real_url, created_at, expires_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(record_id) DO UPDATE SET
		   real_url = excluded.real_url,
		   created_at = excluded.created_at,
		   expires_at = excluded.expires_at`,
		recordID, realURL, now, expiresAt,
	)
	return err
}

// DeleteExpired 硬删除 expires_at 早于 before 的记录
func (r *recordingCacheRepository) DeleteExpired(before time.Time) (int64, error) {
	res, err := r.db.Exec(`DELETE FROM recording_cache WHERE expires_at < ?`, before)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
