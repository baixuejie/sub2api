package pelican

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPostgresConfigurationHTTPBoundary(t *testing.T) {
	repo := testRepository(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	admin := router.Group("/api/v1/admin", func(c *gin.Context) {
		if c.GetHeader("X-Test-Role") != "admin" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	})
	user := router.Group("/api/v1", func(c *gin.Context) {
		if c.GetHeader("X-Test-Role") == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	})
	m := &Module{Repo: repo, Encryptor: testEncryptor{}, EncryptionReady: true}
	m.RegisterRoutes(admin, user, func(*gin.Context) int64 { return 42 })
	request := func(method, path, role, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Test-Role", role)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	key := "test-credential-that-must-never-be-returned"
	payload := SaveConfig{Revision: 1, Enabled: true, APIKey: &key, SelectedGroupIDs: []int64{1, 2}, TopicMode: "rotate", FixedTopicID: "pelican-ski", MaxOutputTokens: 16384, TimeoutSeconds: 300, RetentionDays: 30}
	body, _ := json.Marshal(payload)
	if got := request("PUT", "/api/v1/admin/pelican/config", "user", string(body)); got.Code != 403 {
		t.Fatal("ordinary user wrote config")
	}
	response := request("PUT", "/api/v1/admin/pelican/config", "admin", string(body))
	if response.Code != 200 {
		t.Fatalf("save failed: %s", response.Body.String())
	}
	if bytes.Contains(response.Body.Bytes(), []byte(key)) || strings.Contains(response.Body.String(), "encrypted:") {
		t.Fatal("config response exposed credentials")
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("credential response can be cached")
	}
	var count int
	repo.db.QueryRow(`SELECT count(*) FROM pelican_runs`).Scan(&count)
	if count != 0 {
		t.Fatal("saving config launched generation")
	}
	if got := request("GET", "/api/v1/admin/pelican/runs/1/source", "user", ""); got.Code != 403 {
		t.Fatal("ordinary user accessed raw source")
	}
	if got := request("GET", "/api/v1/pelican/runs", "", ""); got.Code != 401 {
		t.Fatal("gallery exposed to unauthenticated caller")
	}
	if got := request("GET", "/api/v1/pelican/runs/1/artifact", "user", ""); got.Code != 404 || strings.Contains(got.Header().Get("Content-Type"), "text/html") {
		t.Fatal("artifact error served executable HTML")
	}
	unknown := strings.TrimSuffix(string(body), "}") + `,"model":"another-model"}`
	if got := request("PUT", "/api/v1/admin/pelican/config", "admin", unknown); got.Code != 400 {
		t.Fatal("accepted configurable model")
	}
	if got := request("PUT", "/api/v1/admin/pelican/config", "admin", `{"api_key":"`+key); got.Code != 400 || strings.Contains(got.Body.String(), key) {
		t.Fatal("invalid JSON leaked credential")
	}
	if got := request("DELETE", "/api/v1/admin/pelican/config/key?revision=1", "admin", ""); got.Code != 409 {
		t.Fatal("stale key deletion accepted")
	}
	if got := request("DELETE", "/api/v1/admin/pelican/config/key?revision=2", "admin", ""); got.Code != 200 {
		t.Fatal("key deletion failed")
	}
	m.EncryptionReady = false
	payload.Revision = 3
	body, _ = json.Marshal(payload)
	if got := request("PUT", "/api/v1/admin/pelican/config", "admin", string(body)); got.Code != 400 {
		t.Fatal("stored credential under an ephemeral encryption key")
	}
}
