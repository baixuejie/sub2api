package pelican

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestPelicanPostgresIntervalChangesAndClaims(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Hour).Add(23 * time.Minute)
	key := "encrypted:test-key"
	for _, minutes := range []int{60, 30, 10, 60} {
		cfg, err := repo.Config(ctx)
		if err != nil {
			t.Fatal(err)
		}
		in := SaveConfig{Revision: cfg.Revision, Enabled: true, TopicMode: "rotate", FixedTopicID: "pelican-ski", MaxOutputTokens: 16384, TimeoutSeconds: 300, RetentionDays: 30, IntervalMinutes: minutes}
		if err = repo.SaveConfig(ctx, in, &key, 42, now); err != nil {
			t.Fatal(err)
		}
		updated, err := repo.Config(ctx)
		want := nextScheduledRun(now, minutes)
		if err != nil || updated.IntervalMinutes != minutes || updated.NextRunAt == nil || !updated.NextRunAt.Equal(want) {
			t.Fatalf("interval %d: %+v %v", minutes, updated, err)
		}
		status, err := repo.Status(ctx)
		if err != nil || status["interval_seconds"] != minutes*60 {
			t.Fatalf("wrong public interval: %v %v", status, err)
		}
		if claim, err := repo.Claim(ctx, want.Add(-time.Second)); err != nil || claim != nil {
			t.Fatalf("early claim: %+v %v", claim, err)
		}
	}
	for _, minutes := range []int{10, 30, 60} {
		t.Run(fmt.Sprint(minutes), func(t *testing.T) {
			r := testRepository(t)
			interval := time.Duration(minutes) * time.Minute
			now := time.Now().UTC()
			slot := now.Truncate(interval)
			if _, err := r.db.Exec(`UPDATE pelican_config SET enabled=TRUE,api_key_encrypted='encrypted:key',interval_minutes=$1,next_run_at=$2`, minutes, slot.Add(-3*interval)); err != nil {
				t.Fatal(err)
			}
			claim, err := r.Claim(ctx, now)
			if err != nil || claim == nil {
				t.Fatalf("claim: %v", err)
			}
			var at time.Time
			var missed int
			if err = r.db.QueryRow(`SELECT scheduled_for,skipped_hours FROM pelican_runs WHERE id=$1`, claim.ID).Scan(&at, &missed); err != nil {
				t.Fatal(err)
			}
			if !at.Equal(slot) || missed != 3 {
				t.Fatalf("cadence/skip mismatch: %v %d", at, missed)
			}
			cfg, err := r.Config(ctx)
			if err != nil || !cfg.NextRunAt.Equal(slot.Add(interval)) || cfg.Sequence != 4 {
				t.Fatalf("next slot: %+v %v", cfg, err)
			}
			if duplicate, err := r.Claim(ctx, now); err != nil || duplicate != nil {
				t.Fatalf("duplicate: %+v %v", duplicate, err)
			}
		})
	}
}

func TestPelicanPostgresManualConcurrencyCooldownAndSchedule(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	now := time.Now().UTC()
	if _, err := repo.ClaimManual(ctx, now); !errors.Is(err, ErrKeyRequired) {
		t.Fatalf("missing key: %v", err)
	}
	if _, err := repo.db.Exec(`UPDATE pelican_config SET api_key_encrypted='encrypted:key',interval_minutes=10`); err != nil {
		t.Fatal(err)
	}
	before, err := repo.Config(ctx)
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
			c, e := repo.ClaimManual(ctx, now)
			if e != nil {
				errs <- e
			} else {
				claims <- c
			}
		}()
	}
	wg.Wait()
	close(claims)
	close(errs)
	if len(claims) != 1 || len(errs) != 9 {
		t.Fatalf("duplicate dispatch: claims=%d errors=%d", len(claims), len(errs))
	}
	for err := range errs {
		if !errors.Is(err, ErrRunActive) {
			t.Fatal(err)
		}
	}
	claim := <-claims
	if err = repo.Finish(ctx, claim, nil, "", "failed", "mock_failure", 0); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ClaimManual(ctx, now.Add(time.Second)); !errors.Is(err, ErrRunCooldown) {
		t.Fatalf("missing cooldown: %v", err)
	}
	after, err := repo.Config(ctx)
	if err != nil || after.Enabled || after.NextRunAt != nil || after.Sequence != before.Sequence || after.Revision != before.Revision {
		t.Fatalf("manual modified schedule: %+v %v", after, err)
	}
	second, err := repo.ClaimManual(ctx, now.Add(ManualRunCooldown))
	if err != nil || second == nil {
		t.Fatalf("after cooldown: %v", err)
	}
	if _, err = repo.db.Exec(`UPDATE pelican_runs SET lease_expires_at=NOW()-INTERVAL '1 second' WHERE id=$1`, second.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ClaimManual(ctx, now.Add(2*ManualRunCooldown)); err != nil {
		t.Fatalf("expired lease blocks manual: %v", err)
	}
}

func TestPelicanPostgresManualDoesNotConsumeScheduledBoundary(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(10 * time.Minute).Add(10 * time.Minute)
	if _, err := repo.db.Exec(`UPDATE pelican_config SET enabled=TRUE,api_key_encrypted='encrypted:key',interval_minutes=10,next_run_at=$1`, now); err != nil {
		t.Fatal(err)
	}
	manual, err := repo.ClaimManual(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Finish(ctx, manual, nil, "", "failed", "mock_failure", 0); err != nil {
		t.Fatal(err)
	}
	scheduled, err := repo.Claim(ctx, now)
	if err != nil || scheduled == nil || scheduled.ID == manual.ID {
		t.Fatalf("manual consumed automatic time slot: %+v %v", scheduled, err)
	}
}

func TestPelicanPostgresManualHTTPPermissionsAndAsyncResult(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	release := make(chan struct{})
	var once sync.Once
	runner := NewRunner(repo, generatorFunc(func(ctx context.Context, _ string, _ string, _ int) (*Generation, error) {
		select {
		case <-release:
			return &Generation{Text: safeExample}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}), testEncryptor{}, true)
	runner.Start()
	t.Cleanup(func() { once.Do(func() { close(release) }); runner.Stop() })
	m := &Module{Repo: repo, Runner: runner, Encryptor: testEncryptor{}, EncryptionReady: true}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	admin := router.Group("/api/v1/admin", func(c *gin.Context) {
		role := c.GetHeader("X-Test-Role")
		if role == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if role != "admin" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	})
	m.RegisterRoutes(admin, router.Group("/api/v1"), func(*gin.Context) int64 { return 42 })
	request := func(role string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/api/v1/admin/pelican/run", nil)
		r.Header.Set("X-Test-Role", role)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	for role, code := range map[string]int{"": 401, "user": 403, "admin": 400} {
		if w := request(role); w.Code != code {
			t.Fatalf("role %q: %d %s", role, w.Code, w.Body.String())
		}
	}
	if _, err := repo.db.Exec(`UPDATE pelican_config SET api_key_encrypted='unusable-secret'`); err != nil {
		t.Fatal(err)
	}
	if w := request("admin"); w.Code != 400 {
		t.Fatalf("unusable key accepted: %d", w.Code)
	}
	if _, err := repo.db.Exec(`UPDATE pelican_config SET api_key_encrypted='encrypted:mock-key'`); err != nil {
		t.Fatal(err)
	}
	m.EncryptionReady = false
	if w := request("admin"); w.Code != 400 {
		t.Fatalf("no encryption accepted: %d", w.Code)
	}
	m.EncryptionReady = true
	w := request("admin")
	if w.Code != 202 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("submit: %d %s", w.Code, w.Body.String())
	}
	var body struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Data.ID == 0 {
		t.Fatalf("no run ID: %s", w.Body.String())
	}
	if w = request("admin"); w.Code != 409 {
		t.Fatalf("busy: %d", w.Code)
	}
	once.Do(func() { close(release) })
	deadline := time.Now().Add(2 * time.Second)
	for {
		page, err := repo.List(ctx, ListOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Items) == 1 && page.Items[0].ID == body.Data.ID {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("manual output was not published")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if w = request("admin"); w.Code != 429 || w.Header().Get("Retry-After") == "" {
		t.Fatalf("cooldown: %d", w.Code)
	}
}
