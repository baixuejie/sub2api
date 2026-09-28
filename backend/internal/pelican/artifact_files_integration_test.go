package pelican

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func fileTestClaim(t *testing.T, repo *Repository) *Claim {
	t.Helper()
	if _, err := repo.db.Exec(`UPDATE pelican_config SET api_key_encrypted='encrypted:key'`); err != nil {
		t.Fatal(err)
	}
	claim, err := repo.ClaimManual(context.Background(), time.Now().UTC())
	if err != nil || claim == nil {
		t.Fatalf("claim %v", err)
	}
	return claim
}

func TestPelicanFilesPersistBlockedSourceAndRebuildWithoutGeneration(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	claim := fileTestClaim(t, repo)
	raw := `<html><body><svg version="1.1"><circle r="4"/><script>alert(1)</script></svg></body></html>`
	if err := repo.Finish(ctx, claim, &Generation{Text: raw, PreviewNotes: []string{"legacy block"}}, "", "preview_blocked", "preview_unavailable", time.Second); err != nil {
		t.Fatal(err)
	}
	source, err := repo.Artifact(ctx, claim.ID, true)
	if err != nil || source != raw {
		t.Fatal("lost blocked source", err)
	}
	page, err := repo.List(ctx, ListOptions{})
	if err != nil || len(page.Items) != 0 {
		t.Fatal("exposed blocked result")
	}
	if _, err = repo.Artifact(ctx, claim.ID, false); err != ErrNotFound {
		t.Fatal("served blocked preview")
	}
	if err = repo.RebuildPreview(ctx, claim.ID); err != nil {
		t.Fatal(err)
	}
	page, err = repo.List(ctx, ListOptions{})
	if err != nil || len(page.Items) != 1 {
		t.Fatal("rebuild did not publish", err)
	}
	preview, err := repo.Artifact(ctx, claim.ID, false)
	if err != nil || strings.Contains(preview, "<script") || !strings.Contains(preview, "<circle") {
		t.Fatal("invalid rebuilt preview", err)
	}
	var legacyRaw, legacyPreview sql.NullString
	var rawPath, previewPath string
	if err = repo.db.QueryRow(`SELECT raw_output,preview_html,raw_path,preview_path FROM pelican_artifacts WHERE run_id=$1`, claim.ID).Scan(&legacyRaw, &legacyPreview, &rawPath, &previewPath); err != nil {
		t.Fatal(err)
	}
	if legacyRaw.Valid || legacyPreview.Valid || rawPath == "" || previewPath == "" {
		t.Fatal("database stores content instead of paths")
	}
	detail, err := repo.Detail(ctx, claim.ID)
	if err != nil || detail["source_available"] != true {
		t.Fatal("missing source availability", err)
	}
	for _, path := range []string{rawPath, previewPath} {
		if _, err = os.Stat(filepath.Join(repo.files.dir, path)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = repo.db.Exec(`UPDATE pelican_runs SET scheduled_for=NOW()-INTERVAL '31 days' WHERE id=$1`, claim.ID); err != nil {
		t.Fatal(err)
	}
	if err = repo.Cleanup(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(repo.files.dir, strconv.FormatInt(claim.ID, 10))); !os.IsNotExist(err) {
		t.Fatal("retention did not remove files")
	}
}

func TestPelicanFilesMigrationIsVerifiedAndRestartable(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	claim := fileTestClaim(t, repo)
	if _, err := repo.db.Exec(`UPDATE pelican_runs SET status='succeeded',finished_at=NOW() WHERE id=$1`, claim.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.db.Exec(`INSERT INTO pelican_artifacts(run_id,raw_output,preview_html,content_sha256,size_bytes,preview_policy_version) VALUES($1,$2,$2,$3,$4,1)`, claim.ID, safeExample, contentHash(safeExample), len(safeExample)); err != nil {
		t.Fatal(err)
	}
	badDir := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(badDir, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	goodDir := repo.files.dir
	repo.files.dir = badDir
	if err := repo.MigrateArtifacts(ctx); err == nil {
		t.Fatal("migration ignored broken storage")
	}
	var raw string
	if err := repo.db.QueryRow(`SELECT raw_output FROM pelican_artifacts WHERE run_id=$1`, claim.ID).Scan(&raw); err != nil || raw != safeExample {
		t.Fatal("migration lost database source", err)
	}
	repo.files.dir = goodDir
	for i := 0; i < 2; i++ {
		if err := repo.MigrateArtifacts(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if got, err := repo.Artifact(ctx, claim.ID, true); err != nil || got != safeExample {
		t.Fatal("migration changed source", err)
	}
	if page, err := repo.List(ctx, ListOptions{}); err != nil || len(page.Items) != 1 {
		t.Fatal("migration hid old work", err)
	}
	var count int
	repo.db.QueryRow(`SELECT count(*) FROM pelican_artifacts WHERE raw_output IS NOT NULL OR preview_html IS NOT NULL`).Scan(&count)
	if count != 0 {
		t.Fatal("legacy content was not cleared after verification")
	}
}

func TestPelicanFilesPreviewRebuildHTTPPermissions(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	claim := fileTestClaim(t, repo)
	raw := `<svg version="1.1"><circle r="4"/></svg>`
	if err := repo.Finish(ctx, claim, &Generation{Text: raw}, "", "preview_blocked", "preview_unavailable", 0); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	admin := router.Group("/admin", func(c *gin.Context) {
		if c.GetHeader("X-Test-Role") != "admin" {
			c.AbortWithStatus(403)
			return
		}
		c.Next()
	})
	user := router.Group("/user")
	// Deliberately no generator/runner: rebuilding must only use saved files.
	module := &Module{Repo: repo}
	module.RegisterRoutes(admin, user, func(*gin.Context) int64 { return 1 })
	for _, role := range []string{"", "user", "admin"} {
		request := httptest.NewRequest(http.MethodPost, "/admin/pelican/runs/"+strconv.FormatInt(claim.ID, 10)+"/preview", nil)
		request.Header.Set("X-Test-Role", role)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		expected := 403
		if role == "admin" {
			expected = 200
		}
		if response.Code != expected {
			t.Fatalf("role %q: %d %s", role, response.Code, response.Body.String())
		}
	}
	var count int
	repo.db.QueryRow(`SELECT count(*) FROM pelican_runs`).Scan(&count)
	if count != 1 {
		t.Fatal("rebuild created a new generation")
	}
}
