package pelican

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	module  *Module
	actorID func(*gin.Context) int64
}

func (m *Module) RegisterRoutes(admin, user *gin.RouterGroup, actorID func(*gin.Context) int64) {
	h := &Handler{module: m, actorID: actorID}
	a := admin.Group("/pelican")
	a.Use(func(c *gin.Context) { c.Header("Cache-Control", "no-store"); c.Next() })
	a.GET("/config", h.config)
	a.PUT("/config", h.saveConfig)
	a.DELETE("/config/key", h.clearKey)
	a.GET("/runs", func(c *gin.Context) { h.list(c, true) })
	a.GET("/runs/:id", h.detail)
	a.GET("/runs/:id/source", func(c *gin.Context) { h.artifact(c, true) })
	u := user.Group("/pelican")
	u.Use(func(c *gin.Context) { c.Header("Cache-Control", "private, no-store"); c.Next() })
	u.GET("/status", h.status)
	u.GET("/groups", h.groups)
	u.GET("/runs", func(c *gin.Context) { h.list(c, false) })
	u.GET("/runs/:id/artifact", func(c *gin.Context) { h.artifact(c, false) })
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "Not found")
	case errors.Is(err, ErrConflict):
		response.Error(c, http.StatusConflict, ErrConflict.Error())
	case errors.Is(err, ErrGroups), errors.Is(err, ErrKeyRequired):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, "Unable to load or save gallery data")
	}
}

func (h *Handler) config(c *gin.Context) {
	cfg, err := h.module.Repo.Config(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	cfg.EncryptionReady = h.module.EncryptionReady
	if cfg.KeyConfigured {
		key, err := h.module.Encryptor.Decrypt(cfg.EncryptedKey)
		cfg.KeyUnavailable = !cfg.EncryptionReady || err != nil || key == ""
	}
	response.Success(c, cfg)
}

func (h *Handler) saveConfig(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<10)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var in SaveConfig
	if err := decoder.Decode(&in); err != nil {
		response.BadRequest(c, "Invalid configuration fields")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		response.BadRequest(c, "Expected one configuration object")
		return
	}
	if err := in.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if (in.APIKey != nil || in.Enabled) && !h.module.EncryptionReady {
		response.BadRequest(c, "Configure a persistent TOTP encryption key before storing the gallery key")
		return
	}
	var encrypted *string
	if in.APIKey != nil {
		value, err := h.module.Encryptor.Encrypt(*in.APIKey)
		if err != nil {
			response.InternalError(c, "Unable to encrypt API key")
			return
		}
		encrypted = &value
	} else if in.Enabled {
		cfg, err := h.module.Repo.Config(c.Request.Context())
		if err != nil {
			writeError(c, err)
			return
		}
		key, err := h.module.Encryptor.Decrypt(cfg.EncryptedKey)
		if err != nil || key == "" {
			response.BadRequest(c, "Replace the unavailable saved API key before enabling")
			return
		}
	}
	if err := h.module.Repo.SaveConfig(c.Request.Context(), in, encrypted, h.actorID(c), time.Now().UTC()); err != nil {
		writeError(c, err)
		return
	}
	h.config(c)
}

func (h *Handler) clearKey(c *gin.Context) {
	revision, err := strconv.ParseInt(c.Query("revision"), 10, 64)
	if err != nil || revision < 1 {
		response.BadRequest(c, "A configuration revision is required")
		return
	}
	if err = h.module.Repo.ClearKey(c.Request.Context(), revision, h.actorID(c)); err != nil {
		writeError(c, err)
		return
	}
	h.config(c)
}

func (h *Handler) list(c *gin.Context, admin bool) {
	limit := 30
	if s := c.Query("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 100 {
			response.BadRequest(c, "Invalid limit")
			return
		}
		limit = n
	}
	opt := ListOptions{Admin: admin, Limit: limit, Cursor: c.Query("cursor"), Topic: c.Query("topic")}
	if opt.Cursor != "" {
		if _, err := parseCursor(opt.Cursor); err != nil {
			response.BadRequest(c, "Invalid cursor")
			return
		}
	}
	if opt.Topic != "" && topicByID(opt.Topic) == nil {
		response.BadRequest(c, "Invalid topic")
		return
	}
	if s := c.Query("group_id"); s != "" {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil || id < 1 {
			response.BadRequest(c, "Invalid group")
			return
		}
		opt.GroupID = id
	}
	page, err := h.module.Repo.List(c.Request.Context(), opt)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, page)
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		response.NotFound(c, "Not found")
		return 0, false
	}
	return id, true
}
func (h *Handler) artifact(c *gin.Context, source bool) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	content, err := h.module.Repo.Artifact(c.Request.Context(), id, source)
	if err != nil {
		writeError(c, err)
		return
	}
	field := "html"
	if source {
		field = "source"
	}
	response.Success(c, map[string]any{field: content, "policy_version": PreviewPolicyVersion})
}
func (h *Handler) detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	data, err := h.module.Repo.Detail(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, data)
}
func (h *Handler) groups(c *gin.Context) {
	data, err := h.module.Repo.Groups(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, data)
}
func (h *Handler) status(c *gin.Context) {
	data, err := h.module.Repo.Status(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, data)
}
