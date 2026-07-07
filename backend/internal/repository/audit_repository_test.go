package repository

import (
	"testing"
	"time"

	"dvr-manager/pkg/db"
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

func TestStats_aggregatesByDay(t *testing.T) {
	dir := t.TempDir()
	if err := db.InitDB(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewAuditRepository()
	now := time.Now()
	day := now.Format("2006-01-02")
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	inserts := []struct {
		action, detail, status string
	}{
		{"play", "录像已找到", "success"},
		{"play", "录像未找到", "fail"},
		{"play_batch", "批量查询 5 条，找到 3 条", "success"},
		{"stream", "流代理: 录像已找到", "success"},
		{"login_success", "登录成功", "success"},
	}
	for _, row := range inserts {
		if err := repo.Insert(row.action, "alice", "user", "127.0.0.1", "", row.detail, row.status); err != nil {
			t.Fatal(err)
		}
	}

	stats, err := repo.Stats(from, to)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Summary.QuerySingle != 2 {
		t.Fatalf("query_single=%d want 2", stats.Summary.QuerySingle)
	}
	if stats.Summary.QueryBatch != 1 {
		t.Fatalf("query_batch=%d want 1", stats.Summary.QueryBatch)
	}
	if stats.Summary.QueryBatchRecords != 5 {
		t.Fatalf("query_batch_records=%d want 5", stats.Summary.QueryBatchRecords)
	}
	if stats.Summary.Stream != 1 {
		t.Fatalf("stream=%d want 1", stats.Summary.Stream)
	}
	if stats.Summary.LoginSuccess != 1 {
		t.Fatalf("login_success=%d want 1", stats.Summary.LoginSuccess)
	}
	if stats.Summary.ActiveUsers != 1 {
		t.Fatalf("active_users=%d want 1", stats.Summary.ActiveUsers)
	}
	if len(stats.Series) != 1 || stats.Series[0].Date != day {
		t.Fatalf("series=%+v want one day %s", stats.Series, day)
	}
}
