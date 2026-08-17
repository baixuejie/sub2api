package deepseekharness

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

const maxRequestBodyBytes = 16 << 10

type installHandler struct {
	service *installService
}

func newInstallHandler(service *installService) *installHandler {
	return &installHandler{service: service}
}

func (h *installHandler) Profile(c *gin.Context) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	apiKeyID, err := strconv.ParseInt(strings.TrimSpace(c.Query("api_key_id")), 10, 64)
	if err != nil || apiKeyID <= 0 {
		response.BadRequest(c, "Invalid api_key_id")
		return
	}
	profile, err := h.service.Profile(c.Request.Context(), subject.UserID, apiKeyID, requestOrigin(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, profile)
}

func (h *installHandler) CreateSession(c *gin.Context) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var request CreateSessionRequest
	if err := decodeBoundedJSON(c, &request); err != nil || request.APIKeyID <= 0 {
		response.BadRequest(c, "Invalid install session request")
		return
	}
	session, err := h.service.CreateSession(c.Request.Context(), subject.UserID, request, requestOrigin(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Accepted(c, session)
}

func (h *installHandler) GetSession(c *gin.Context) {
	subject, ok := servermiddleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	session, err := h.service.GetSession(c.Request.Context(), subject.UserID, c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, session)
}

func (h *installHandler) Exchange(c *gin.Context) {
	var request ExchangeRequest
	if err := decodeBoundedJSON(c, &request); err != nil {
		response.BadRequest(c, "Invalid exchange request")
		return
	}
	task, err := h.service.Exchange(c.Request.Context(), request)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	response.Success(c, task)
}

func (h *installHandler) UpdateSession(c *gin.Context) {
	eventToken := bearerToken(c.GetHeader("Authorization"))
	if eventToken == "" {
		response.Unauthorized(c, "Invalid helper credential")
		return
	}
	var event InstallEvent
	if err := decodeBoundedJSON(c, &event); err != nil {
		response.BadRequest(c, "Invalid install event")
		return
	}
	session, err := h.service.UpdateSession(c.Request.Context(), c.Param("id"), eventToken, event)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Success(c, session)
}

func decodeBoundedJSON(c *gin.Context, target any) error {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return io.ErrUnexpectedEOF
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func bearerToken(value string) string {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func requestOrigin(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return normalizedOrigin(scheme + "://" + c.Request.Host)
}

func normalizedOrigin(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	return (&url.URL{Scheme: parsed.Scheme, Host: parsed.Host}).String()
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errFeatureDisabled):
		response.NotFound(c, "DeepSeek Harness installation is disabled")
	case errors.Is(err, errAPIKeyNotFound), errors.Is(err, errSessionNotFound):
		response.NotFound(c, "Resource not found")
	case errors.Is(err, errTicketNotFound), errors.Is(err, errInvalidEventToken):
		response.Unauthorized(c, "Invalid or expired helper credential")
	case errors.Is(err, errAPIKeyUnavailable):
		response.Error(c, http.StatusConflict, "API key is not available")
	case errors.Is(err, errUnsupportedGroup):
		response.BadRequest(c, "API key group is not supported")
	case errors.Is(err, errInvalidModel):
		response.BadRequest(c, "Model is not available for this API key group")
	case errors.Is(err, errInvalidBaseURL):
		response.Error(c, http.StatusConflict, "Site API base URL is not configured correctly")
	case errors.Is(err, errInvalidSession), errors.Is(err, errInvalidEvent):
		response.BadRequest(c, "Invalid install session state")
	default:
		if isPublicError(err) {
			response.BadRequest(c, err.Error())
			return
		}
		response.ErrorFrom(c, err)
	}
}
