package pelican

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// Set PELICAN_TEST_DATABASE_URL to a disposable local PostgreSQL instance.
// Each test owns a new schema and never touches existing application tables.
func testRepository(t *testing.T) *Repository {
	t.Helper()
	dsn := os.Getenv("PELICAN_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PELICAN_TEST_DATABASE_URL is not set")
	}
	root, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("pelican_test_%d", time.Now().UnixNano())
	if _, err = root.Exec("CREATE SCHEMA " + schema); err != nil {
		root.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = root.Exec("DROP SCHEMA " + schema + " CASCADE"); root.Close() })
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			t.Fatal(err)
		}
		q := u.Query()
		q.Set("search_path", schema)
		u.RawQuery = q.Encode()
		dsn = u.String()
	} else {
		dsn += " search_path=" + schema
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(12)
	if _, err = db.Exec(`CREATE TABLE groups(id BIGINT PRIMARY KEY,name TEXT,status TEXT,deleted_at TIMESTAMPTZ);
 INSERT INTO groups VALUES(1,'A','active',NULL),(2,'B','active',NULL),(3,'C','active',NULL);`); err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "241_local_pelican_gallery.sql"))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err = db.Exec(string(migration)); err != nil {
			t.Fatal(err)
		}
	}
	return NewRepository(db)
}

func TestPostgresClaimOnceAndPublishOnlySuccess(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	now := time.Now().UTC()
	in := SaveConfig{Revision: 1, Enabled: true, SelectedGroupIDs: []int64{1, 2, 3}, TopicMode: "rotate", FixedTopicID: "pelican-ski", MaxOutputTokens: 16384, TimeoutSeconds: 300, RetentionDays: 30}
	key := "encrypted:designated-key"
	if err := repo.SaveConfig(ctx, in, &key, 0, now); err != nil {
		t.Fatal(err)
	}
	cfg, err := repo.Config(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.NextRunAt == nil || !cfg.NextRunAt.Equal(nextHour(now)) {
		t.Fatal("enabling generated immediately")
	}
	if err = repo.SaveConfig(ctx, in, &key, 0, now); err != ErrConflict {
		t.Fatalf("stale config accepted: %v", err)
	}
	_, err = repo.db.Exec(`UPDATE pelican_config SET next_run_at=$1`, now.Add(-3*time.Hour).Truncate(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	claims := make(chan *Claim, 10)
	errs := make(chan error, 10)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, e := repo.Claim(ctx, now)
			if e != nil {
				errs <- e
			}
			if c != nil {
				claims <- c
			}
		}()
	}
	wg.Wait()
	close(claims)
	close(errs)
	for e := range errs {
		t.Error(e)
	}
	if len(claims) != 1 {
		t.Fatalf("expected one claim, got %d", len(claims))
	}
	claim := <-claims
	var count, skipped int
	var groups string
	if err = repo.db.QueryRow(`SELECT count(*),max(skipped_hours) FROM pelican_runs`).Scan(&count, &skipped); err != nil {
		t.Fatal(err)
	}
	if count != 1 || skipped != 3 {
		t.Fatalf("replayed downtime: count=%d skipped=%d", count, skipped)
	}
	if err = repo.db.QueryRow(`SELECT selected_groups_snapshot::text FROM pelican_runs WHERE id=$1`, claim.ID).Scan(&groups); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{`"A"`, `"B"`, `"C"`} {
		if !strings.Contains(groups, name) {
			t.Fatalf("missing group tag %s", groups)
		}
	}
	page, err := repo.List(ctx, ListOptions{})
	if err != nil || len(page.Items) != 0 {
		t.Fatalf("running task exposed: %+v %v", page, err)
	}
	preview, err := preparePreview(safeExample)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Finish(ctx, claim, &Generation{Text: safeExample}, preview, "succeeded", "", time.Second); err != nil {
		t.Fatal(err)
	}
	for _, group := range []int64{1, 2, 3} {
		page, err = repo.List(ctx, ListOptions{GroupID: group})
		if err != nil || len(page.Items) != 1 || page.Items[0].ID != claim.ID {
			t.Fatalf("label does not select shared work: %v %v", page, err)
		}
	}
	_, _ = repo.db.Exec(`UPDATE groups SET name='Renamed' WHERE id=1`)
	page, err = repo.List(ctx, ListOptions{})
	if err != nil || page.Items[0].Groups[0].Name != "A" {
		t.Fatal("history label was mutated")
	}
	// Simulate a later hour without waiting or making a model request.
	_, _ = repo.db.Exec(`UPDATE pelican_runs SET scheduled_for=scheduled_for-INTERVAL '1 hour'; UPDATE pelican_config SET next_run_at=date_trunc('hour',NOW())`)
	failed, err := repo.Claim(ctx, time.Now().UTC())
	if err != nil || failed == nil {
		t.Fatalf("next hour not claimed: %v", err)
	}
	if err = repo.Finish(ctx, failed, nil, "", "failed", "http_429", time.Second); err != nil {
		t.Fatal(err)
	}
	page, err = repo.List(ctx, ListOptions{})
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("failure leaked into gallery: %+v %v", page, err)
	}
	if _, err = repo.Artifact(ctx, failed.ID, false); err != ErrNotFound {
		t.Fatalf("failed artifact exposed: %v", err)
	}
	admin, err := repo.List(ctx, ListOptions{Admin: true, Limit: 1})
	if err != nil || len(admin.Items) != 1 || admin.NextCursor == "" {
		t.Fatalf("admin pagination: %+v %v", admin, err)
	}
	older, err := repo.List(ctx, ListOptions{Admin: true, Limit: 1, Cursor: admin.NextCursor})
	if err != nil || len(older.Items) != 1 || older.Items[0].ID != claim.ID {
		t.Fatalf("cursor lost older result: %+v %v", older, err)
	}
}

func TestPostgresExpiredLeaseAndRetention(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	_, err := repo.db.Exec(`UPDATE pelican_config SET enabled=TRUE,api_key_encrypted='encrypted:key',next_run_at=date_trunc('hour',NOW())`)
	if err != nil {
		t.Fatal(err)
	}
	claim, err := repo.Claim(ctx, time.Now().UTC())
	if err != nil || claim == nil {
		t.Fatal(err)
	}
	_, err = repo.db.Exec(`UPDATE pelican_runs SET lease_expires_at=NOW()-INTERVAL '1 second' WHERE id=$1`, claim.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Claim(ctx, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if err = repo.Finish(ctx, claim, &Generation{Text: safeExample}, safeExample, "succeeded", "", 0); err != nil {
		t.Fatal(err)
	}
	var status string
	repo.db.QueryRow(`SELECT status FROM pelican_runs WHERE id=$1`, claim.ID).Scan(&status)
	if status != "interrupted" {
		t.Fatal("late worker overwrote terminal state")
	}
	if _, err = repo.Artifact(ctx, claim.ID, true); err != ErrNotFound {
		t.Fatal("late worker wrote artifact")
	}
	_, _ = repo.db.Exec(`UPDATE pelican_runs SET scheduled_for=NOW()-INTERVAL '31 days'`)
	if err = repo.Cleanup(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	repo.db.QueryRow(`SELECT count(*) FROM pelican_runs`).Scan(&count)
	if count != 0 {
		t.Fatal("retention left expired rows")
	}
}
