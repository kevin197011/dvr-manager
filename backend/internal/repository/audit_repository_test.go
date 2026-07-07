package repository

import (
	"testing"
	"time"

	"dvr-vod-system/pkg/db"
)

func TestDeleteOlderThan_removesExpiredRows(t *testing.T) {
	dir := t.TempDir()
	if err := db.InitDB(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewAuditRepository()
	if err := repo.Insert("play", "u", "user", "127.0.0.1", "r1", "old", "success"); err != nil {
		t.Fatal(err)
	}

	database := db.GetDB()
	if _, err := database.Exec(`UPDATE audit_log SET created_at = ? WHERE resource = ?`,
		"2020-01-01 00:00:00", "r1"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Insert("play", "u", "user", "127.0.0.1", "r2", "new", "success"); err != nil {
		t.Fatal(err)
	}

	cutoff := time.Now().AddDate(0, -3, 0)
	n, err := repo.DeleteOlderThan(cutoff)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("deleted=%d want 1", n)
	}

	list, total, err := repo.List(nil, nil, "", "", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(list) != 1 || list[0].Resource != "r2" {
		t.Fatalf("list=%+v total=%d", list, total)
	}
}
