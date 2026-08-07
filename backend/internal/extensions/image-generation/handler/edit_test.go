package imagegeneration

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const testPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="

func buildTestEditMultipart(t *testing.T, imageData []byte) ([]byte, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("group_id", "7"))
	require.NoError(t, writer.WriteField("model", "gpt-image-2"))
	require.NoError(t, writer.WriteField("prompt", "make it cinematic"))
	require.NoError(t, writer.WriteField("n", "2"))
	require.NoError(t, writer.WriteField("size", "1024x1024"))
	require.NoError(t, writer.WriteField("quality", "auto"))
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{
		"name":     "image",
		"filename": "..\\source.png",
	}))
	header.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(header)
	require.NoError(t, err)
	_, err = part.Write(imageData)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return body.Bytes(), writer.FormDataContentType()
}

func TestParseEditMultipartRequestValidatesAndNormalizesImageParts(t *testing.T) {
	imageData, err := base64.StdEncoding.DecodeString(testPNGBase64)
	require.NoError(t, err)
	body, contentType := buildTestEditMultipart(t, imageData)

	parsed, err := parseEditMultipartRequest(body, contentType)
	require.NoError(t, err)
	require.Equal(t, int64(7), parsed.Request.GroupID)
	require.Equal(t, "gpt-image-2", parsed.Request.Model)
	require.Equal(t, 2, parsed.Request.N)
	require.Len(t, parsed.Uploads, 1)
	require.Equal(t, "image/png", parsed.Uploads[0].ContentType)
	require.Equal(t, "source.png", parsed.Uploads[0].FileName)

	forwarded, forwardedType, err := buildEditMultipartRequest(testPreparedGeneration().Request, parsed)
	require.NoError(t, err)
	require.NotContains(t, string(forwarded), "group_id")
	require.Contains(t, string(forwarded), "response_format")
	mediaType, params, err := mime.ParseMediaType(forwardedType)
	require.NoError(t, err)
	require.Equal(t, "multipart/form-data", mediaType)
	forwardedReader := multipart.NewReader(bytes.NewReader(forwarded), params["boundary"])
	var imageParts int
	for {
		part, nextErr := forwardedReader.NextPart()
		if nextErr == io.EOF {
			break
		}
		require.NoError(t, nextErr)
		if part.FormName() == "image" && part.FileName() != "" {
			imageParts++
		}
		_ = part.Close()
	}
	require.Equal(t, 1, imageParts)
}

func TestParseEditMultipartRequestRejectsNonImageData(t *testing.T) {
	body, contentType := buildTestEditMultipart(t, []byte("not an image"))
	_, err := parseEditMultipartRequest(body, contentType)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported image file type")
}

func TestImageEditHandlerUsesSanitizedMultipartAndRestoresPanelRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prepared := testPreparedGeneration()
	policy := &fakePolicy{prepared: prepared}
	gateway := &recordingGateway{}
	h := NewHandler(policy, gateway)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42, Concurrency: 2})
		c.Next()
	})
	router.Use(h.PrepareEditAPIKey())
	router.Use(func(c *gin.Context) {
		require.True(t, strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data; boundary="))
		require.Equal(t, "Bearer sk-server-only-secret", c.GetHeader("Authorization"))
		c.Set(string(middleware.ContextKeyAPIKey), prepared.APIKey)
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42, Concurrency: 2})
		c.Next()
	})
	router.POST("/api/v1/image-generation/edit", h.Edit)

	imageData, err := base64.StdEncoding.DecodeString(testPNGBase64)
	require.NoError(t, err)
	body, contentType := buildTestEditMultipart(t, imageData)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-generation/edit", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer panel-jwt")
	req.Header.Set("x-api-key", "panel-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 1, policy.prepareCalls)
	require.Equal(t, "/v1/images/edits", gateway.calledPath)
	require.Contains(t, string(gateway.body), "source.png")
	require.NotContains(t, string(gateway.body), "name=group_id")
	require.Equal(t, "/api/v1/image-generation/edit", req.URL.Path)
	require.Equal(t, "Bearer panel-jwt", req.Header.Get("Authorization"))
	require.Equal(t, "panel-key", req.Header.Get("x-api-key"))
	require.NotContains(t, w.Body.String(), "sk-server-only-secret")
}

func TestReadEditBodyLimitsRequest(t *testing.T) {
	_, err := readEditBody(strings.NewReader(strings.Repeat("x", maxImageEditBodyBytes+1)))
	require.ErrorIs(t, err, errEditBodyTooLarge)
	_, err = readEditBody(io.LimitReader(strings.NewReader(""), 0))
	require.Error(t, err)
}
