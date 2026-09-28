package pelican

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const safeExample = `<!doctype html><html><head><style>@keyframes sway{from{transform:rotate(-3deg)}to{transform:rotate(3deg)}}.bird{animation:sway 2s infinite alternate;transform-origin:center}</style></head><body><svg viewBox="0 0 960 720" xmlns="http://www.w3.org/2000/svg"><defs><linearGradient id="snow"><stop offset="0" stop-color="#fff"/><stop offset="1" stop-color="#acd"/></linearGradient></defs><rect width="960" height="720" fill="url(#snow)"/><g class="bird"><circle cx="480" cy="350" r="70" fill="white"/><animateTransform attributeName="transform" type="translate" values="0 0;0 4;0 0" dur="3s" repeatCount="indefinite"/></g></svg></body></html>`

func completedPayload(text string) string {
	b, _ := json.Marshal(map[string]any{"id": "resp_test", "model": Model, "status": "completed", "output": []any{
		map[string]any{"type": "reasoning", "summary": "not HTML"},
		map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]string{"type": "output_text", "text": text}}},
	}, "usage": map[string]any{"input_tokens": 20, "output_tokens": 30, "total_tokens": 50, "output_tokens_details": map[string]int{"reasoning_tokens": 10}}})
	return string(b)
}

func TestGeneratorFixedRequestAndFinalText(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Header.Get("Authorization") != "Bearer test-designated-key" {
			t.Error("wrong credential")
		}
		var body map[string]json.RawMessage
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Fatal("invalid JSON")
		}
		if string(body["model"]) != `"gpt-6-astra"` || string(body["reasoning"]) != `{"effort":"high"}` {
			t.Errorf("incorrect fixed parameters: %s %s", body["model"], body["reasoning"])
		}
		if _, ok := body["tools"]; ok {
			t.Error("generation must not provide tools")
		}
		if string(body["store"]) != "false" || string(body["stream"]) != "true" {
			t.Error("wrong stream/store parameters")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"do not duplicate\"}\n\n")
		fmt.Fprintf(w, "data: {\"type\":\"response.completed\",\"response\":%s}\n\n", completedPayload(safeExample))
	}))
	defer srv.Close()
	g := NewGenerator(srv.URL)
	defer g.Close()
	result, err := g.Generate(context.Background(), "test-designated-key", topics[0].prompt(), 16384)
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 || result.Text != safeExample || result.TotalTokens == nil || *result.TotalTokens != 50 {
		t.Fatalf("unexpected generation: calls=%d result=%+v", calls.Load(), result)
	}
}

func TestGeneratorNeverRetriesOrLeaksUpstreamErrors(t *testing.T) {
	for _, status := range []int{401, 429, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(status)
				fmt.Fprint(w, "failure echoed secret-designated-key")
			}))
			defer srv.Close()
			g := NewGenerator(srv.URL)
			defer g.Close()
			_, err := g.Generate(context.Background(), "secret-designated-key", "prompt", 16384)
			if err == nil || strings.Contains(err.Error(), "secret-designated-key") || calls.Load() != 1 {
				t.Fatalf("unsafe error/retry: %v calls=%d", err, calls.Load())
			}
		})
	}
	var redirected atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { redirected.Add(1) }))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	g := NewGenerator(source.URL)
	defer g.Close()
	if _, err := g.Generate(context.Background(), "secret", "prompt", 16384); err == nil || redirected.Load() != 0 {
		t.Fatal("redirect forwarded credentials")
	}
}

func TestResponsesRejectPartialAndUnsafeResults(t *testing.T) {
	for name, stream := range map[string]string{
		"delta only": "data: {\"type\":\"response.output_text.delta\",\"delta\":\"<svg/>\"}\n\n",
		"done only":  "data: [DONE]\n\n",
		"incomplete": "data: {\"type\":\"response.incomplete\"}\n\n",
		"malformed":  "data: {\n\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseStream(strings.NewReader(stream)); err == nil {
				t.Fatal("accepted incomplete stream")
			}
		})
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, completedPayload("<svg>secret-key-echo</svg>"))
	}))
	defer srv.Close()
	g := NewGenerator(srv.URL)
	defer g.Close()
	if _, err := g.Generate(context.Background(), "secret-key-echo", "prompt", 16384); err == nil || err.Error() != "credential_echo" {
		t.Fatalf("credential echo accepted: %v", err)
	}
	if _, err := parseCompleted([]byte(completedPayload(strings.Repeat("a", MaxArtifactBytes+1)))); err == nil {
		t.Fatal("accepted oversized output")
	}
}

func TestPreviewRetainsAnimationAndBlocksActiveContent(t *testing.T) {
	preview, err := preparePreview("```html\n" + safeExample + "\n```")
	if err != nil {
		t.Fatalf("valid animated SVG rejected: %v", err)
	}
	for _, expected := range []string{"Content-Security-Policy", "@keyframes sway", "animateTransform", "url(#snow)"} {
		if !strings.Contains(preview, expected) {
			t.Errorf("lost %s", expected)
		}
	}
	for name, content := range map[string]string{
		"script":        `<script>alert(1)</script>`,
		"event":         `<svg onload="alert(1)"></svg>`,
		"external use":  `<svg><use href="https://example.com/a.svg#x"/></svg>`,
		"external CSS":  `<style>svg{background:url(https://example.com/track)}</style>`,
		"escaped CSS":   `<style>svg{background:u\72l(https://example.com/x)}</style>`,
		"import":        `<style>@import 'https://example.com/x';</style>`,
		"animated URL":  `<svg><animate attributeName="href" values="javascript:alert(1)"/></svg>`,
		"meta redirect": `<meta http-equiv="refresh" content="0;url=https://example.com">`,
		"frame":         `<iframe srcdoc="hello"></iframe>`,
		"image-set":     `<style>svg{background:image-set('https://example.com/a' 1x)}</style>`,
	} {
		t.Run(name, func(t *testing.T) {
			preview, notes, err := sanitizePreview(strings.Replace(safeExample, "</body>", content+"</body>", 1))
			if err != nil || preview == "" {
				t.Fatalf("discarded otherwise valid drawing: %v", err)
			}
			if len(notes) == 0 {
				t.Fatal("missing sanitization diagnostics")
			}
			for _, unsafe := range []string{"<script", "onload=", "<iframe", "https://example.com", "javascript:", "http-equiv=\"refresh\""} {
				if strings.Contains(preview, unsafe) {
					t.Fatalf("retained active content: %s", unsafe)
				}
			}
		})
	}
}

func TestConfigurationAndSchedule(t *testing.T) {
	c := SaveConfig{Revision: 1, SelectedGroupIDs: []int64{3, 2, 3}, TopicMode: "rotate", FixedTopicID: "pelican-ski", MaxOutputTokens: 16384, TimeoutSeconds: 300, RetentionDays: 30}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(c.SelectedGroupIDs) != 2 {
		t.Fatal("did not deduplicate group labels")
	}
	now := time.Date(2026, 9, 27, 14, 25, 0, 0, time.FixedZone("CST", 8*3600))
	if got := nextHour(now); got.In(now.Location()).Hour() != 15 {
		t.Fatalf("unexpected slot: %v", got)
	}
	if selectTopic(Config{Sequence: 0}, 3).ID != "pelican-ski" {
		t.Fatal("downtime rotation is not stable")
	}
	for _, topic := range topics {
		if !strings.Contains(topic.prompt(), "不检查、不测试、不解释") {
			t.Fatal("missing code-only instruction")
		}
	}
	for host, expected := range map[string]string{"0.0.0.0": "http://127.0.0.1:8080/v1/responses", "::": "http://[::1]:8080/v1/responses", "127.0.0.1": "http://127.0.0.1:8080/v1/responses"} {
		if got := LocalEndpoint(host, 8080); got != expected {
			t.Fatalf("endpoint %q", got)
		}
	}
}

type testEncryptor struct{}

func (testEncryptor) Encrypt(s string) (string, error) { return "encrypted:" + s, nil }
func (testEncryptor) Decrypt(s string) (string, error) {
	if !strings.HasPrefix(s, "encrypted:") {
		return "", errors.New("bad secret")
	}
	return strings.TrimPrefix(s, "encrypted:"), nil
}

type fakeRunStore struct {
	status, code, preview string
	finishes              int
}

func (*fakeRunStore) Claim(context.Context, time.Time) (*Claim, error) { return nil, nil }
func (*fakeRunStore) ClaimManual(context.Context, time.Time) (*Claim, error) {
	return &Claim{ID: 7, Config: Config{EncryptedKey: "encrypted:key", TimeoutSeconds: 60, MaxOutputTokens: 16384}, Prompt: safeExample}, nil
}
func (*fakeRunStore) ClaimActive(context.Context, *Claim) (bool, error) { return true, nil }
func (s *fakeRunStore) Finish(_ context.Context, _ *Claim, _ *Generation, preview, status, code string, _ time.Duration) error {
	s.finishes++
	s.status = status
	s.code = code
	s.preview = preview
	return nil
}
func (*fakeRunStore) Cleanup(context.Context) error { return nil }

type generatorFunc func(context.Context, string, string, int) (*Generation, error)

func (f generatorFunc) Generate(c context.Context, k, p string, n int) (*Generation, error) {
	return f(c, k, p, n)
}

func TestWorkerSuccessFailureAndCancellation(t *testing.T) {
	for _, fail := range []bool{false, true} {
		store := &fakeRunStore{}
		calls := 0
		g := generatorFunc(func(context.Context, string, string, int) (*Generation, error) {
			calls++
			if fail {
				return nil, failure("failed", "http_429")
			}
			return &Generation{Text: safeExample}, nil
		})
		runner := NewRunner(store, g, testEncryptor{}, true)
		runner.execute(&Claim{Config: Config{EncryptedKey: "encrypted:test-key", TimeoutSeconds: 60, MaxOutputTokens: 16384}, Prompt: topics[0].prompt()})
		if calls != 1 || store.finishes != 1 {
			t.Fatal("worker retried or did not finalize once")
		}
		if fail && store.status != "failed" || !fail && (store.status != "succeeded" || store.preview == "") {
			t.Fatalf("unexpected terminal state: %+v", store)
		}
		runner.Stop()
	}
	store := &fakeRunStore{}
	entered := make(chan struct{})
	done := make(chan struct{})
	g := generatorFunc(func(ctx context.Context, _ string, _ string, _ int) (*Generation, error) {
		close(entered)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	runner := NewRunner(store, g, testEncryptor{}, true)
	go func() {
		runner.execute(&Claim{Config: Config{EncryptedKey: "encrypted:key", TimeoutSeconds: 60}})
		close(done)
	}()
	<-entered
	runner.cancel()
	<-done
	if store.status != "interrupted" || store.code != "shutdown" {
		t.Fatalf("shutdown not recorded: %+v", store)
	}
}
