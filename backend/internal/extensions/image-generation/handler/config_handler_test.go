package imagegeneration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	imagegenerationservice "github.com/Wei-Shaw/sub2api/internal/extensions/image-generation/service"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	core "github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type promptPolicy struct {
	prepared *imagegenerationservice.PreparedPrompt
}

func (p *promptPolicy) GetOptions(context.Context, int64) (imagegenerationservice.Options, error) {
	return imagegenerationservice.Options{}, nil
}

func (p *promptPolicy) Prepare(context.Context, int64, imagegenerationservice.GenerationRequest) (*imagegenerationservice.PreparedGeneration, error) {
	return nil, nil
}

func (p *promptPolicy) GetConfigOptions(context.Context, int64) (imagegenerationservice.ConfigOptions, error) {
	return imagegenerationservice.ConfigOptions{}, nil
}

func (p *promptPolicy) SaveConfig(_ context.Context, _ int64, config imagegenerationservice.UserImageConfig) (imagegenerationservice.ConfigOptions, error) {
	return imagegenerationservice.ConfigOptions{Config: config}, nil
}

func (p *promptPolicy) PreparePrompt(_ context.Context, _ int64, prompt string) (*imagegenerationservice.PreparedPrompt, error) {
	result := *p.prepared
	result.Prompt = strings.TrimSpace(prompt)
	return &result, nil
}

type promptRecordingGateway struct {
	path string
	body []byte
}

func (g *promptRecordingGateway) Images(*gin.Context) {}

func (g *promptRecordingGateway) ChatCompletions(c *gin.Context) {
	g.path = c.Request.URL.Path
	g.body, _ = io.ReadAll(c.Request.Body)
	c.JSON(http.StatusOK, gin.H{
		"id":      "chatcmpl-private-id",
		"debug":   "upstream-private-metadata",
		"choices": []gin.H{{"message": gin.H{"role": "assistant", "content": "A detailed cinematic mountain at sunrise"}}},
	})
}

func promptPrepared() *imagegenerationservice.PreparedPrompt {
	groupID := int64(7)
	return &imagegenerationservice.PreparedPrompt{
		APIKey: &core.APIKey{
			ID: 12, UserID: 42, Key: "sk-server-only-secret", GroupID: &groupID, Status: core.StatusAPIKeyActive,
			Group: &core.Group{ID: groupID, Platform: core.PlatformOpenAI, Status: core.StatusActive, Hydrated: true},
			User:  &core.User{ID: 42, Status: core.StatusActive},
		},
		GroupID: groupID,
		Model:   "gpt-4.1-mini",
	}
}

func TestPromptOptimizationUsesGatewayAndReturnsOnlyOptimizedPrompt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prepared := promptPrepared()
	policy := &promptPolicy{prepared: prepared}
	gateway := &promptRecordingGateway{}
	h := NewHandler(policy, gateway)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42, Concurrency: 2})
		c.Next()
	})
	router.Use(h.PrepareOptimizationAPIKey())
	router.Use(func(c *gin.Context) {
		require.Equal(t, "Bearer sk-server-only-secret", c.GetHeader("Authorization"))
		c.Set(string(middleware.ContextKeyAPIKey), prepared.APIKey)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42, Concurrency: 2})
		c.Next()
	})
	router.POST("/api/v1/image-generation/optimize", h.Optimize)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-generation/optimize", strings.NewReader(`{"prompt":"mountain at sunrise"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer panel-jwt")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, openAIChatPath, gateway.path)
	require.Contains(t, string(gateway.body), promptOptimizationSystemMessage)
	require.Contains(t, string(gateway.body), `"stream":false`)
	require.Equal(t, "/api/v1/image-generation/optimize", req.URL.Path)
	require.Equal(t, "Bearer panel-jwt", req.Header.Get("Authorization"))
	require.Contains(t, w.Body.String(), "A detailed cinematic mountain at sunrise")
	require.NotContains(t, w.Body.String(), "chatcmpl-private-id")
	require.NotContains(t, w.Body.String(), "upstream-private-metadata")
	require.NotContains(t, w.Body.String(), prepared.APIKey.Key)
}

func TestPromptOptimizationRejectsMismatchedAPIKeyContext(t *testing.T) {
	prepared := promptPrepared()
	h := NewHandler(&promptPolicy{prepared: prepared}, &promptRecordingGateway{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/image-generation/optimize", strings.NewReader(`{}`))
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
	c.Set(preparedPromptContextKey, preparedPromptRequest{
		UserID: 42, GroupID: prepared.GroupID, APIKeyID: prepared.APIKey.ID, Model: prepared.Model, Prompt: "draw",
	})
	wrong := *prepared.APIKey
	wrong.ID = 99
	c.Set(string(middleware.ContextKeyAPIKey), &wrong)

	h.Optimize(c)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestOptimizedPromptFromResponseSupportsTextParts(t *testing.T) {
	body, err := json.Marshal(gin.H{"choices": []gin.H{{"message": gin.H{"content": []gin.H{
		{"type": "text", "text": "first"}, {"type": "output_text", "text": "second"},
	}}}}})
	require.NoError(t, err)
	value, err := optimizedPromptFromResponse(body)
	require.NoError(t, err)
	require.Equal(t, "first\nsecond", value)
}

func TestImageGenerationConfigHandlersUseJWTContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	policy := &promptPolicy{prepared: promptPrepared()}
	h := NewHandler(policy, &promptRecordingGateway{})
	jwtAuth := middleware.JWTAuthMiddleware(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42, Concurrency: 2})
		c.Next()
	})
	router := gin.New()
	configRoutes := router.Group("/api/v1/image-generation")
	configRoutes.Use(gin.HandlerFunc(jwtAuth))
	configRoutes.GET("/config", h.Config)
	configRoutes.PUT("/config", h.SaveConfig)

	for _, method := range []string{http.MethodGet, http.MethodPut} {
		var body io.Reader
		if method == http.MethodPut {
			body = strings.NewReader(`{"default_n":1}`)
		}
		req := httptest.NewRequest(method, "/api/v1/image-generation/config", body)
		req.Header.Set("Authorization", "Bearer panel-jwt")
		if method == http.MethodPut {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, method)
		require.NotContains(t, w.Body.String(), "sk-server-only-secret")
	}
}

func TestCaptureWriterKeepsTheFirstStatusAndCapturesBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	capture := newCaptureWriter(ctx.Writer)

	require.False(t, capture.Written())
	require.Equal(t, -1, capture.Size())
	capture.Header().Set("Content-Type", "application/json")
	capture.WriteHeader(http.StatusAccepted)
	capture.WriteHeader(http.StatusBadGateway)
	_, err := capture.WriteString(`{"ok":true}`)
	require.NoError(t, err)
	capture.Flush()

	require.Equal(t, http.StatusAccepted, capture.Status())
	require.Equal(t, len(`{"ok":true}`), capture.Size())
	require.True(t, capture.Written())
	require.Equal(t, `{"ok":true}`, capture.body.String())
}
