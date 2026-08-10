package tutorials

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCoverHandlerUploadAndServe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataDir := t.TempDir()
	h := NewCoverHandler(dataDir)
	router := gin.New()
	router.POST("/api/v1/admin/tutorials/covers", h.Upload)
	router.GET("/api/v1/tutorials/covers/:filename", h.Serve)

	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00}
	request := tutorialCoverUploadRequest(t, "../../ignored.png", "application/octet-stream", png)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)

	var result struct {
		Code int `json:"code"`
		Data struct {
			Filename    string `json:"filename"`
			URL         string `json:"url"`
			ContentType string `json:"content_type"`
			Size        int64  `json:"size"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &result))
	require.Equal(t, 0, result.Code)
	require.Regexp(t, tutorialCoverNamePattern, result.Data.Filename)
	require.Equal(t, tutorialCoverURLPrefix+result.Data.Filename, result.Data.URL)
	require.Equal(t, "image/png", result.Data.ContentType)
	require.Equal(t, int64(len(png)), result.Data.Size)

	stored, err := os.ReadFile(filepath.Join(dataDir, TutorialCoversDir, result.Data.Filename))
	require.NoError(t, err)
	require.Equal(t, png, stored)

	readRecorder := httptest.NewRecorder()
	router.ServeHTTP(readRecorder, httptest.NewRequest(http.MethodGet, result.Data.URL, nil))
	require.Equal(t, http.StatusOK, readRecorder.Code)
	require.Equal(t, "image/png", readRecorder.Header().Get("Content-Type"))
	require.Equal(t, "nosniff", readRecorder.Header().Get("X-Content-Type-Options"))
	require.Equal(t, png, readRecorder.Body.Bytes())
}

func TestCoverHandlerRejectsInvalidContentAndPublicTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCoverHandler(t.TempDir())
	router := gin.New()
	router.POST("/api/v1/admin/tutorials/covers", h.Upload)
	router.GET("/api/v1/tutorials/covers/:filename", h.Serve)

	badRecorder := httptest.NewRecorder()
	router.ServeHTTP(badRecorder, tutorialCoverUploadRequest(t, "cover.png", "image/png", []byte("not an image")))
	require.Equal(t, http.StatusBadRequest, badRecorder.Code)

	overSizePNG := append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, bytes.Repeat([]byte{'x'}, TutorialCoverMaxBytes)...)
	overSizeRecorder := httptest.NewRecorder()
	router.ServeHTTP(overSizeRecorder, tutorialCoverUploadRequest(t, "cover.png", "image/png", overSizePNG))
	require.Equal(t, http.StatusBadRequest, overSizeRecorder.Code)

	traversalRecorder := httptest.NewRecorder()
	router.ServeHTTP(traversalRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/tutorials/covers/..%2Fconfig.yaml", nil))
	require.Equal(t, http.StatusNotFound, traversalRecorder.Code)
}

func TestRegisterAdminRoutesUsesExistingAdminMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	admin := v1.Group("/admin")
	admin.Use(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})
	RegisterAdminRoutes(admin, t.TempDir())

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, tutorialCoverUploadRequest(t, "cover.png", "image/png", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}))
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func tutorialCoverUploadRequest(t *testing.T, filename, contentType string, body []byte) *http.Request {
	t.Helper()
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(body)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tutorials/covers", &buffer)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-Test-Declared-Type", contentType)
	return request
}

func TestValidateCoverURLAllowsOnlyGeneratedSameOriginPaths(t *testing.T) {
	valid := tutorialCoverURLPrefix + strings.Repeat("a", 32) + ".png"
	require.NoError(t, validateCoverURL(valid))
	require.Error(t, validateCoverURL(tutorialCoverURLPrefix+"../cover.png"))
}
