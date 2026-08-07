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

const (
	preparedPromptContextKey = "ext.image_generation.prepared_prompt"
	openAIChatPath           = "/v1/chat/completions"
	maxConfigRequestBytes    = 16 << 10
)

const promptOptimizationSystemMessage = `You improve prompts for image generation. Preserve the user's intent, subjects, constraints, and language. Add useful visual detail such as composition, lighting, color, material, camera, and style only when it helps. Do not add unsafe content or invent requirements that conflict with the request. Return only the optimized prompt, with no title, explanation, markdown, or quotation marks.`

type ConfigurationPolicy interface {
	GetConfigOptions(ctx context.Context, userID int64) (imagegenerationservice.ConfigOptions, error)
	SaveConfig(ctx context.Context, userID int64, config imagegenerationservice.UserImageConfig) (imagegenerationservice.ConfigOptions, error)
	PreparePrompt(ctx context.Context, userID int64, prompt string) (*imagegenerationservice.PreparedPrompt, error)
}

type preparedPromptRequest struct {
	UserID   int64
	GroupID  int64
	APIKeyID int64
	Model    string
	Prompt   string
}

type promptOptimizationResult struct {
	OriginalPrompt  string `json:"original_prompt"`
	OptimizedPrompt string `json:"optimized_prompt"`
}

func (h *Handler) configurationPolicy() (ConfigurationPolicy, bool) {
	if h == nil || h.service == nil {
		return nil, false
	}
	policy, ok := h.service.(ConfigurationPolicy)
	return policy, ok
}

func (h *Handler) Config(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	policy, ok := h.configurationPolicy()
	if !ok {
		response.InternalError(c, "Image generation configuration is unavailable")
		return
	}
	options, err := policy.GetConfigOptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, options)
}

func (h *Handler) SaveConfig(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	policy, ok := h.configurationPolicy()
	if !ok {
		response.InternalError(c, "Image generation configuration is unavailable")
		return
	}
	var requested imagegenerationservice.UserImageConfig
	if err := readBoundedJSON(c.Request, maxConfigRequestBytes, &requested); err != nil {
		response.BadRequest(c, "Invalid image generation configuration")
		return
	}
	options, err := policy.SaveConfig(c.Request.Context(), subject.UserID, requested)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, options)
}

// PrepareOptimizationAPIKey selects the configured server-side credential and
// delegates authentication to the existing API key middleware.
func (h *Handler) PrepareOptimizationAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok := middleware.GetAuthSubjectFromContext(c)
		if !ok || subject.UserID <= 0 {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}
		policy, ok := h.configurationPolicy()
		if !ok || c.Request == nil {
			response.InternalError(c, "Prompt optimization is unavailable")
			c.Abort()
			return
		}
		var requested imagegenerationservice.PromptOptimizationRequest
		body, err := readBoundedJSONBytes(c.Request, maxPanelRequestBytes, &requested)
		if err != nil {
			response.BadRequest(c, "Invalid prompt optimization request")
			c.Abort()
			return
		}
		prepared, err := policy.PreparePrompt(c.Request.Context(), subject.UserID, requested.Prompt)
		if err != nil {
			response.ErrorFrom(c, err)
			c.Abort()
			return
		}
		if prepared == nil || prepared.APIKey == nil || prepared.APIKey.ID <= 0 || strings.TrimSpace(prepared.APIKey.Key) == "" {
			response.ErrorFrom(c, imagegenerationservice.ErrPromptConfig)
			c.Abort()
			return
		}
		c.Set(preparedPromptContextKey, preparedPromptRequest{
			UserID: subject.UserID, GroupID: prepared.GroupID, APIKeyID: prepared.APIKey.ID,
			Model: prepared.Model, Prompt: prepared.Prompt,
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
			restoreHeader(request.Header, "Authorization", originalAuthorization)
			restoreHeader(request.Header, "x-api-key", originalAPIKey)
			restoreHeader(request.Header, "x-goog-api-key", originalGoogleAPIKey)
		}()
		c.Next()
	}
}

func (h *Handler) Optimize(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h == nil || h.promptGateway == nil {
		response.InternalError(c, "Prompt optimization is unavailable")
		return
	}
	value, exists := c.Get(preparedPromptContextKey)
	prepared, preparedOK := value.(preparedPromptRequest)
	if !exists || !preparedOK || prepared.UserID != subject.UserID || prepared.GroupID <= 0 || prepared.APIKeyID <= 0 || strings.TrimSpace(prepared.Model) == "" || strings.TrimSpace(prepared.Prompt) == "" {
		response.Forbidden(c, "Prompt optimization request is not prepared")
		return
	}
	apiKey, apiKeyOK := middleware.GetAPIKeyFromContext(c)
	if !apiKeyOK || apiKey == nil || apiKey.ID != prepared.APIKeyID || apiKey.UserID != subject.UserID || apiKey.GroupID == nil || *apiKey.GroupID != prepared.GroupID || apiKey.Group == nil || !apiKey.Group.IsActive() {
		response.Forbidden(c, "Prompt optimization authorization is invalid")
		return
	}
	payload := map[string]any{
		"model":  prepared.Model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": promptOptimizationSystemMessage},
			{"role": "user", "content": prepared.Prompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		response.InternalError(c, "Failed to prepare prompt optimization request")
		return
	}

	request := c.Request
	originalPath := request.URL.Path
	originalRawPath := request.URL.RawPath
	originalBody := request.Body
	originalLength := request.ContentLength
	originalContentType := request.Header.Get("Content-Type")
	originalWriter := c.Writer
	capture := newCaptureWriter(originalWriter)
	restored := false
	restore := func() {
		if restored {
			return
		}
		restored = true
		c.Writer = originalWriter
		request.URL.Path = originalPath
		request.URL.RawPath = originalRawPath
		request.Body = originalBody
		request.ContentLength = originalLength
		restoreHeader(request.Header, "Content-Type", originalContentType)
	}
	defer restore()
	c.Writer = capture
	request.URL.Path = openAIChatPath
	request.URL.RawPath = ""
	request.Body = io.NopCloser(bytes.NewReader(body))
	request.ContentLength = int64(len(body))
	request.Header.Set("Content-Type", "application/json")
	h.promptGateway.ChatCompletions(c)
	restore()

	status := capture.Status()
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		if status < http.StatusBadRequest || status > 599 {
			status = http.StatusBadGateway
		}
		response.Error(c, status, "Prompt optimization failed")
		return
	}
	optimized, err := optimizedPromptFromResponse(capture.body.Bytes())
	if err != nil || optimized == "" {
		response.Error(c, http.StatusBadGateway, "Prompt optimization returned an invalid response")
		return
	}
	response.Success(c, promptOptimizationResult{OriginalPrompt: prepared.Prompt, OptimizedPrompt: optimized})
}

func optimizedPromptFromResponse(body []byte) (string, error) {
	var payload struct {
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(body), &payload); err != nil || len(payload.Choices) == 0 {
		return "", err
	}
	content := payload.Choices[0].Message.Content
	var text string
	if json.Unmarshal(content, &text) == nil {
		return strings.TrimSpace(text), nil
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(content, &parts); err != nil {
		return "", err
	}
	var builder strings.Builder
	for _, part := range parts {
		if part.Type == "text" || part.Type == "output_text" || part.Type == "" {
			if value := strings.TrimSpace(part.Text); value != "" {
				if builder.Len() > 0 {
					builder.WriteByte('\n')
				}
				builder.WriteString(value)
			}
		}
	}
	return strings.TrimSpace(builder.String()), nil
}

func readBoundedJSON(request *http.Request, limit int64, destination any) error {
	_, err := readBoundedJSONBytes(request, limit, destination)
	return err
}

func readBoundedJSONBytes(request *http.Request, limit int64, destination any) ([]byte, error) {
	if request == nil || request.Body == nil {
		return nil, io.ErrUnexpectedEOF
	}
	body, err := io.ReadAll(io.LimitReader(request.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit || len(bytes.TrimSpace(body)) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	if err := json.Unmarshal(body, destination); err != nil {
		return nil, err
	}
	return body, nil
}

func restoreHeader(header http.Header, name, value string) {
	if value == "" {
		header.Del(name)
		return
	}
	header.Set(name, value)
}

type captureWriter struct {
	gin.ResponseWriter
	header        http.Header
	status        int
	size          int
	headerWritten bool
	body          bytes.Buffer
}

func newCaptureWriter(parent gin.ResponseWriter) *captureWriter {
	return &captureWriter{ResponseWriter: parent, header: make(http.Header), status: http.StatusOK, size: -1}
}

func (w *captureWriter) Header() http.Header { return w.header }

func (w *captureWriter) WriteHeader(status int) {
	if status > 0 && !w.headerWritten {
		w.status = status
		w.headerWritten = true
		w.size = 0
	}
}

func (w *captureWriter) WriteHeaderNow() {
	if !w.headerWritten {
		w.headerWritten = true
		w.size = 0
	}
}

func (w *captureWriter) Write(value []byte) (int, error) {
	w.WriteHeaderNow()
	n, err := w.body.Write(value)
	w.size += n
	return n, err
}

func (w *captureWriter) WriteString(value string) (int, error) {
	return w.Write([]byte(value))
}

func (w *captureWriter) Status() int { return w.status }

func (w *captureWriter) Size() int { return w.size }

func (w *captureWriter) Written() bool { return w.headerWritten }

func (w *captureWriter) Flush() { w.WriteHeaderNow() }
