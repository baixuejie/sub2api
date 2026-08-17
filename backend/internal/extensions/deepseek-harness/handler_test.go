package deepseekharness

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDeepSeekHarnessHandlerRequiresAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newInstallHandler(newTestInstallService(activeOpenAIKey(), newMemoryInstallStore()))
	router := gin.New()
	router.GET("/profile", handler.Profile)

	request := httptest.NewRequest(http.MethodGet, "/profile?api_key_id=42", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.NotContains(t, response.Body.String(), "sk-test-secret-value")
}

func TestDeepSeekHarnessHandlerFeatureFlagFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := newInstallService(
		&fakeAPIKeyReader{key: activeOpenAIKey()},
		&fakeSettingsReader{baseURL: "https://api.example.com", enabled: false},
		newMemoryInstallStore(),
	)
	handler := newInstallHandler(service)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 7})
		c.Next()
	})
	router.GET("/profile", handler.Profile)

	request := httptest.NewRequest(http.MethodGet, "/profile?api_key_id=42", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestDeepSeekHarnessHandlerProfileDoesNotExposeRawKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newInstallHandler(newTestInstallService(activeOpenAIKey(), newMemoryInstallStore()))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 7})
		c.Next()
	})
	router.GET("/profile", handler.Profile)

	request := httptest.NewRequest(http.MethodGet, "/profile?api_key_id=42", nil)
	request.Header.Set("Origin", "https://panel.example.com")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "gpt-5.6-sol")
	require.NotContains(t, response.Body.String(), "sk-test-secret-value")
}

func TestDeepSeekHarnessHandlerRejectsUnknownJSONFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newInstallHandler(newTestInstallService(activeOpenAIKey(), newMemoryInstallStore()))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 7})
		c.Next()
	})
	router.POST("/sessions", handler.CreateSession)

	request := httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(`{"api_key_id":42,"api_key":"must-not-be-accepted"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.NotContains(t, response.Body.String(), "must-not-be-accepted")
}

func TestDeepSeekHarnessExchangeResponseDisablesCaching(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := newTestInstallService(activeOpenAIKey(), newMemoryInstallStore())
	created, err := service.CreateSession(t.Context(), 7, CreateSessionRequest{APIKeyID: 42}, "https://api.example.com")
	require.NoError(t, err)
	launch, err := url.Parse(created.LaunchURI)
	require.NoError(t, err)

	handler := newInstallHandler(service)
	router := gin.New()
	router.POST("/exchange", handler.Exchange)
	request := httptest.NewRequest(http.MethodPost, "/exchange", strings.NewReader(`{"ticket":"`+launch.Query().Get("ticket")+`"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	require.Equal(t, "no-cache", response.Header().Get("Pragma"))
	require.Contains(t, response.Body.String(), "sk-test-secret-value")
}

func TestDeepSeekHarnessRequestOriginIgnoresBrowserAndForwardedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodGet, "http://panel.example.com/profile", nil)
	context.Request.Host = "panel.example.com"
	context.Request.Header.Set("Origin", "https://attacker.example.com")
	context.Request.Header.Set("X-Forwarded-Proto", "https")

	require.Equal(t, "http://panel.example.com", requestOrigin(context))
}

func TestDeepSeekHarnessHandlerHelperEventRequiresBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newInstallHandler(newTestInstallService(activeOpenAIKey(), newMemoryInstallStore()))
	router := gin.New()
	router.POST("/sessions/:id/events", handler.UpdateSession)

	request := httptest.NewRequest(http.MethodPost, "/sessions/session-1/events", strings.NewReader(`{"status":"installing","progress":20}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusUnauthorized, response.Code)
}
