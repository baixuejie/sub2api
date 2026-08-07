package imagegeneration

import (
	"context"
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

type fakePolicy struct {
	prepared     *imagegenerationservice.PreparedGeneration
	err          error
	prepareCalls int
}

func (f *fakePolicy) GetOptions(context.Context, int64) (imagegenerationservice.Options, error) {
	return imagegenerationservice.Options{}, nil
}

func (f *fakePolicy) Prepare(context.Context, int64, imagegenerationservice.GenerationRequest) (*imagegenerationservice.PreparedGeneration, error) {
	f.prepareCalls++
	return f.prepared, f.err
}

type recordingGateway struct {
	calledPath string
	apiKey     *core.APIKey
	sub        *core.UserSubscription
	body       []byte
}

func (g *recordingGateway) Images(c *gin.Context) {
	g.calledPath = c.Request.URL.Path
	g.apiKey, _ = middleware.GetAPIKeyFromContext(c)
	g.sub, _ = middleware.GetSubscriptionFromContext(c)
	g.body, _ = io.ReadAll(c.Request.Body)
	c.JSON(http.StatusOK, gin.H{"data": []gin.H{{"b64_json": "safe-image"}}})
}

func testPreparedGeneration() *imagegenerationservice.PreparedGeneration {
	groupID := int64(7)
	return &imagegenerationservice.PreparedGeneration{
		APIKey: &core.APIKey{
			ID:      12,
			UserID:  42,
			Key:     "sk-server-only-secret",
			GroupID: &groupID,
			Status:  core.StatusAPIKeyActive,
			Group:   &core.Group{ID: groupID, Platform: core.PlatformOpenAI, Status: core.StatusActive, AllowImageGeneration: true, Hydrated: true},
			User:    &core.User{ID: 42, Status: core.StatusActive},
		},
		Request: imagegenerationservice.GenerationRequest{
			GroupID: 7, Model: "gpt-image-2", Prompt: "draw a mountain", N: 1,
			Size: "1024x1024", Quality: "auto", OutputFormat: "png", Background: "auto", Moderation: "auto",
		},
	}
}

func TestImageGenerationHandlerPrepareAPIKeyUsesAuthenticatedContextAndRestoresHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prepared := testPreparedGeneration()
	policy := &fakePolicy{prepared: prepared}
	gateway := &recordingGateway{}
	h := NewHandler(policy, gateway)
	subscription := &core.UserSubscription{ID: 3, UserID: 42, GroupID: 7, Status: core.SubscriptionStatusActive}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42, Concurrency: 2})
		c.Next()
	})
	router.Use(h.PrepareAPIKey())
	router.Use(func(c *gin.Context) {
		require.Equal(t, "Bearer sk-server-only-secret", c.GetHeader("Authorization"))
		// This simulates the trusted output of APIKeyAuthMiddleware. Generate must
		// consume it, not write these values itself.
		c.Set(string(middleware.ContextKeyAPIKey), prepared.APIKey)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: prepared.APIKey.UserID, Concurrency: 2})
		c.Set(string(middleware.ContextKeySubscription), subscription)
		c.Next()
	})
	router.POST("/api/v1/image-generation/generate", h.Generate)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-generation/generate", strings.NewReader(`{"group_id":7,"model":"gpt-image-2","prompt":"draw"}`))
	req.Header.Set("Authorization", "Bearer user-jwt")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 1, policy.prepareCalls)
	require.Equal(t, "/v1/images/generations", gateway.calledPath)
	require.NotNil(t, gateway.apiKey)
	require.Equal(t, "sk-server-only-secret", gateway.apiKey.Key)
	require.Same(t, subscription, gateway.sub)
	require.Contains(t, string(gateway.body), `"response_format":"b64_json"`)
	require.Equal(t, "/api/v1/image-generation/generate", req.URL.Path)
	require.Equal(t, "Bearer user-jwt", req.Header.Get("Authorization"))
	require.NotContains(t, w.Body.String(), "sk-server-only-secret")
}

func TestImageGenerationHandlerPrepareAPIKeyLimitsBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	policy := &fakePolicy{prepared: testPreparedGeneration()}
	h := NewHandler(policy, &recordingGateway{})
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
		c.Next()
	})
	router.Use(h.PrepareAPIKey())
	router.POST("/api/v1/image-generation/generate", h.Generate)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-generation/generate", strings.NewReader(strings.Repeat("x", maxPanelRequestBytes+1)))
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	require.Zero(t, policy.prepareCalls)
}

func TestImageGenerationHandlerGenerateRejectsMismatchedAPIKeyContext(t *testing.T) {
	prepared := testPreparedGeneration()
	h := NewHandler(&fakePolicy{prepared: prepared}, &recordingGateway{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/image-generation/generate", strings.NewReader(`{}`))
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
	c.Set(preparedRequestContextKey, preparedRequest{UserID: 42, GroupID: 7, APIKeyID: 12, Request: prepared.Request})
	wrong := *prepared.APIKey
	wrong.ID = 99
	c.Set(string(middleware.ContextKeyAPIKey), &wrong)
	h.Generate(c)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestImageGenerationHandlerGenerateRequiresAuthentication(t *testing.T) {
	h := NewHandler(&fakePolicy{prepared: testPreparedGeneration()}, &recordingGateway{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/image-generation/generate", strings.NewReader(`{"prompt":"draw"}`))

	h.Generate(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}
