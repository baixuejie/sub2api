package imagegeneration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	imagegenerationservice "github.com/Wei-Shaw/sub2api/internal/extensions/image-generation/service"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

const openAIImageGenerationsPath = "/v1/images/generations"

const (
	preparedRequestContextKey = "ext.image_generation.prepared_request"
	maxPanelRequestBytes      = 64 << 10
)

// ImageGateway is the minimal adapter implemented by OpenAIGatewayHandler.
type ImageGateway interface {
	Images(c *gin.Context)
}

type Policy interface {
	GetOptions(ctx context.Context, userID int64) (imagegenerationservice.Options, error)
	Prepare(ctx context.Context, userID int64, req imagegenerationservice.GenerationRequest) (*imagegenerationservice.PreparedGeneration, error)
}

type Handler struct {
	service Policy
	gateway ImageGateway
}

type preparedRequest struct {
	UserID   int64
	GroupID  int64
	APIKeyID int64
	Request  imagegenerationservice.GenerationRequest
}

func NewHandler(service Policy, gateway ImageGateway) *Handler {
	return &Handler{service: service, gateway: gateway}
}

// PrepareAPIKey validates the browser payload, selects a server-side key, and
// lets the existing API-key middleware authenticate that key. The replacement
// Authorization header exists only while the downstream middleware runs.
func (h *Handler) PrepareAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok := middleware.GetAuthSubjectFromContext(c)
		if !ok || subject.UserID <= 0 {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}
		if h == nil || h.service == nil || c.Request == nil || c.Request.Body == nil {
			response.InternalError(c, "Image generation is unavailable")
			c.Abort()
			return
		}

		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxPanelRequestBytes+1))
		if err != nil {
			response.BadRequest(c, "Invalid image generation request")
			c.Abort()
			return
		}
		if len(body) > maxPanelRequestBytes {
			response.Error(c, http.StatusRequestEntityTooLarge, "Image generation request is too large")
			c.Abort()
			return
		}
		if len(bytes.TrimSpace(body)) == 0 {
			response.BadRequest(c, "Image generation request is empty")
			c.Abort()
			return
		}
		var req imagegenerationservice.GenerationRequest
		if err := json.Unmarshal(body, &req); err != nil {
			response.BadRequest(c, "Invalid image generation request")
			c.Abort()
			return
		}
		prepared, err := h.service.Prepare(c.Request.Context(), subject.UserID, req)
		if err != nil {
			response.ErrorFrom(c, err)
			c.Abort()
			return
		}
		if prepared == nil || prepared.APIKey == nil || prepared.APIKey.ID <= 0 || strings.TrimSpace(prepared.APIKey.Key) == "" {
			response.ErrorFrom(c, imagegenerationservice.ErrImageAPIKeyMissing)
			c.Abort()
			return
		}

		// Keep only non-sensitive data in the extension context. The API key
		// itself is reloaded and authenticated by APIKeyAuthMiddleware.
		c.Set(preparedRequestContextKey, preparedRequest{
			UserID:   subject.UserID,
			GroupID:  prepared.Request.GroupID,
			APIKeyID: prepared.APIKey.ID,
			Request:  prepared.Request,
		})

		request := c.Request
		originalBody := request.Body
		originalLength := request.ContentLength
		originalAuthorization := request.Header.Get("Authorization")
		originalAPIKey := request.Header.Get("x-api-key")
		originalGoogleAPIKey := request.Header.Get("x-goog-api-key")
		request.Body = io.NopCloser(bytes.NewReader(body))
		request.ContentLength = int64(len(body))
		request.Header.Set("Authorization", "Bearer "+prepared.APIKey.Key)
		request.Header.Del("x-api-key")
		request.Header.Del("x-goog-api-key")
		defer func() {
			request.Body = originalBody
			request.ContentLength = originalLength
			if originalAuthorization == "" {
				request.Header.Del("Authorization")
			} else {
				request.Header.Set("Authorization", originalAuthorization)
			}
			if originalAPIKey == "" {
				request.Header.Del("x-api-key")
			} else {
				request.Header.Set("x-api-key", originalAPIKey)
			}
			if originalGoogleAPIKey == "" {
				request.Header.Del("x-goog-api-key")
			} else {
				request.Header.Set("x-goog-api-key", originalGoogleAPIKey)
			}
		}()

		c.Next()
	}
}

// Options returns image-capable groups and their actual configured models.
func (h *Handler) Options(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h == nil || h.service == nil {
		response.InternalError(c, "Image generation is unavailable")
		return
	}

	options, err := h.service.GetOptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, options)
}

// Generate validates the panel request and delegates all upstream work and billing
// to the existing OpenAI images gateway. The selected API key stays server-side.
func (h *Handler) Generate(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h == nil || h.service == nil || h.gateway == nil {
		response.InternalError(c, "Image generation is unavailable")
		return
	}

	value, exists := c.Get(preparedRequestContextKey)
	prepared, preparedOK := value.(preparedRequest)
	if !exists || !preparedOK || prepared.UserID != subject.UserID || prepared.GroupID <= 0 || prepared.APIKeyID <= 0 {
		response.Forbidden(c, "Image generation request is not prepared")
		return
	}
	apiKey, apiKeyOK := middleware.GetAPIKeyFromContext(c)
	if !apiKeyOK || apiKey == nil || apiKey.ID != prepared.APIKeyID || apiKey.UserID != subject.UserID || apiKey.GroupID == nil || *apiKey.GroupID != prepared.GroupID || apiKey.Group == nil || !apiKey.Group.IsActive() || !apiKey.Group.AllowImageGeneration {
		response.Forbidden(c, "Image generation authorization is invalid")
		return
	}

	payload := map[string]any{
		"model":           prepared.Request.Model,
		"prompt":          prepared.Request.Prompt,
		"n":               prepared.Request.N,
		"size":            prepared.Request.Size,
		"quality":         prepared.Request.Quality,
		"output_format":   prepared.Request.OutputFormat,
		"background":      prepared.Request.Background,
		"moderation":      prepared.Request.Moderation,
		"response_format": "b64_json",
	}
	if prepared.Request.OutputCompression != nil {
		payload["output_compression"] = *prepared.Request.OutputCompression
	}
	body, err := json.Marshal(payload)
	if err != nil {
		response.InternalError(c, "Failed to prepare image generation request")
		return
	}

	request := c.Request
	originalPath := request.URL.Path
	originalRawPath := request.URL.RawPath
	originalBody := request.Body
	originalLength := request.ContentLength
	originalContentType := request.Header.Get("Content-Type")
	defer func() {
		request.URL.Path = originalPath
		request.URL.RawPath = originalRawPath
		request.Body = originalBody
		request.ContentLength = originalLength
		if originalContentType == "" {
			request.Header.Del("Content-Type")
		} else {
			request.Header.Set("Content-Type", originalContentType)
		}
	}()

	request.URL.Path = openAIImageGenerationsPath
	request.URL.RawPath = ""
	request.Body = io.NopCloser(bytes.NewReader(body))
	request.ContentLength = int64(len(body))
	request.Header.Set("Content-Type", "application/json")
	// The gateway writes an OpenAI-compatible response directly. Keeping this shape
	// avoids copying large base64 image payloads into another response envelope.
	h.gateway.Images(c)
}
