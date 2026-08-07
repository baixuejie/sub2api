package imagegeneration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	imagegenerationhandler "github.com/Wei-Shaw/sub2api/internal/extensions/image-generation/handler"
	imagegenerationservice "github.com/Wei-Shaw/sub2api/internal/extensions/image-generation/service"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	core "github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type routeSecurityPolicy struct {
	prepared     *imagegenerationservice.PreparedGeneration
	prepareCalls int
	preparedFor  int64
}

func (p *routeSecurityPolicy) GetOptions(context.Context, int64) (imagegenerationservice.Options, error) {
	return imagegenerationservice.Options{}, nil
}

func (p *routeSecurityPolicy) Prepare(_ context.Context, userID int64, _ imagegenerationservice.GenerationRequest) (*imagegenerationservice.PreparedGeneration, error) {
	p.prepareCalls++
	p.preparedFor = userID
	return p.prepared, nil
}

type routeSecurityGateway struct {
	calls       int
	authChecked bool
}

func (g *routeSecurityGateway) Images(c *gin.Context) {
	g.calls++
	value, _ := c.Get("test.api_key_auth_executed")
	g.authChecked, _ = value.(bool)
	c.JSON(http.StatusOK, gin.H{"data": []gin.H{{"b64_json": "safe-image"}}})
}

func routeSecurityPrepared() *imagegenerationservice.PreparedGeneration {
	groupID := int64(7)
	return &imagegenerationservice.PreparedGeneration{
		APIKey: &core.APIKey{
			ID:      12,
			UserID:  42,
			Key:     "sk-server-only-secret",
			GroupID: &groupID,
			Status:  core.StatusAPIKeyActive,
			Group: &core.Group{
				ID:                   groupID,
				Platform:             core.PlatformOpenAI,
				Status:               core.StatusActive,
				AllowImageGeneration: true,
				Hydrated:             true,
			},
			User: &core.User{ID: 42, Status: core.StatusActive},
		},
		Request: imagegenerationservice.GenerationRequest{
			GroupID:      groupID,
			Model:        "gpt-image-2",
			Prompt:       "draw a mountain",
			N:            1,
			Size:         "1024x1024",
			Quality:      "auto",
			OutputFormat: "png",
			Background:   "auto",
			Moderation:   "auto",
		},
	}
}

func TestImageGenerationRouteExecutesAPIKeyAuthAndRestoresPanelCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prepared := routeSecurityPrepared()
	policy := &routeSecurityPolicy{prepared: prepared}
	gateway := &routeSecurityGateway{}
	h := imagegenerationhandler.NewHandler(policy, gateway)

	var apiKeyAuthCalls int
	jwtAuth := middleware.JWTAuthMiddleware(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42, Concurrency: 2})
		c.Next()
	})
	apiKeyAuth := middleware.APIKeyAuthMiddleware(func(c *gin.Context) {
		apiKeyAuthCalls++
		require.Equal(t, "Bearer sk-server-only-secret", c.GetHeader("Authorization"))
		require.Empty(t, c.GetHeader("x-api-key"))
		require.Empty(t, c.GetHeader("x-goog-api-key"))
		require.Equal(t, "/api/v1/image-generation/generate", c.Request.URL.Path)
		c.Set("test.api_key_auth_executed", true)
		c.Set(string(middleware.ContextKeyAPIKey), prepared.APIKey)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42, Concurrency: 2})
		c.Next()
	})

	router := gin.New()
	RegisterRoutes(router.Group("/api/v1"), h, jwtAuth, apiKeyAuth, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-generation/generate", strings.NewReader(`{"group_id":7,"model":"gpt-image-2","prompt":"draw"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer panel-jwt")
	req.Header.Set("x-api-key", "original-x-api-key")
	req.Header.Set("x-goog-api-key", "original-google-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 1, policy.prepareCalls)
	require.Equal(t, int64(42), policy.preparedFor)
	require.Equal(t, 1, apiKeyAuthCalls)
	require.Equal(t, 1, gateway.calls)
	require.True(t, gateway.authChecked)
	require.Equal(t, "Bearer panel-jwt", req.Header.Get("Authorization"))
	require.Equal(t, "original-x-api-key", req.Header.Get("x-api-key"))
	require.Equal(t, "original-google-key", req.Header.Get("x-goog-api-key"))
	require.NotContains(t, w.Body.String(), prepared.APIKey.Key)
}

func TestImageGenerationRouteRejectsJWTAndAPIKeyUserMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prepared := routeSecurityPrepared()
	policy := &routeSecurityPolicy{prepared: prepared}
	gateway := &routeSecurityGateway{}
	h := imagegenerationhandler.NewHandler(policy, gateway)

	jwtAuth := middleware.JWTAuthMiddleware(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42, Concurrency: 2})
		c.Next()
	})
	apiKeyAuth := middleware.APIKeyAuthMiddleware(func(c *gin.Context) {
		otherUserKey := *prepared.APIKey
		otherUserKey.UserID = 99
		otherUserKey.User = &core.User{ID: 99, Status: core.StatusActive}
		c.Set("test.api_key_auth_executed", true)
		c.Set(string(middleware.ContextKeyAPIKey), &otherUserKey)
		// APIKeyAuthMiddleware replaces the JWT subject with the key owner.
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 99, Concurrency: 2})
		c.Next()
	})

	router := gin.New()
	RegisterRoutes(router.Group("/api/v1"), h, jwtAuth, apiKeyAuth, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-generation/generate", strings.NewReader(`{"group_id":7,"model":"gpt-image-2","prompt":"draw"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer panel-jwt")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Zero(t, gateway.calls)
	require.Equal(t, "Bearer panel-jwt", req.Header.Get("Authorization"))
	require.NotContains(t, w.Body.String(), prepared.APIKey.Key)
}

func TestImageGenerationRouteRestoresPanelCredentialsWhenAPIKeyAuthRejects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prepared := routeSecurityPrepared()
	policy := &routeSecurityPolicy{prepared: prepared}
	gateway := &routeSecurityGateway{}
	h := imagegenerationhandler.NewHandler(policy, gateway)

	jwtAuth := middleware.JWTAuthMiddleware(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42, Concurrency: 2})
		c.Next()
	})
	apiKeyAuth := middleware.APIKeyAuthMiddleware(func(c *gin.Context) {
		require.Equal(t, "Bearer sk-server-only-secret", c.GetHeader("Authorization"))
		c.AbortWithStatus(http.StatusForbidden)
	})

	router := gin.New()
	RegisterRoutes(router.Group("/api/v1"), h, jwtAuth, apiKeyAuth, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-generation/generate", strings.NewReader(`{"group_id":7,"model":"gpt-image-2","prompt":"draw"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer panel-jwt")
	req.Header.Set("x-api-key", "original-x-api-key")
	req.Header.Set("x-goog-api-key", "original-google-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Zero(t, gateway.calls)
	require.Equal(t, "Bearer panel-jwt", req.Header.Get("Authorization"))
	require.Equal(t, "original-x-api-key", req.Header.Get("x-api-key"))
	require.Equal(t, "original-google-key", req.Header.Get("x-goog-api-key"))
	require.NotContains(t, w.Body.String(), prepared.APIKey.Key)
}
