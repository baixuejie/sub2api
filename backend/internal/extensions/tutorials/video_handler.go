package tutorials

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// PublicVideo is the public tutorial video contract enriched with its
// persisted aggregate play count.
type PublicVideo struct {
	Video
	PlayCount int64 `json:"play_count"`
}

// VideoHandler serves the public tutorial list and play-count endpoint.
type VideoHandler struct {
	videos PublicVideosProvider
	counts PlayCountStore
}

func NewVideoHandler(videos PublicVideosProvider, counts PlayCountStore) *VideoHandler {
	return &VideoHandler{videos: videos, counts: counts}
}

// List handles GET /api/v1/tutorials/videos.
func (h *VideoHandler) List(c *gin.Context) {
	videos, err := h.listVideos(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	ids := make([]string, 0, len(videos))
	for _, video := range videos {
		ids = append(ids, video.ID)
	}
	counts := make(map[string]int64, len(ids))
	if h.counts != nil {
		counts, err = h.counts.List(c.Request.Context(), ids)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
	}

	result := make([]PublicVideo, 0, len(videos))
	for _, video := range videos {
		result = append(result, PublicVideo{Video: video, PlayCount: counts[video.ID]})
	}
	response.Success(c, gin.H{"videos": result})
}

// Play handles POST /api/v1/tutorials/videos/:id/play.
func (h *VideoHandler) Play(c *gin.Context) {
	rawID := strings.TrimSpace(c.Param("id"))
	if !strings.HasSuffix(rawID, "/play") {
		response.NotFound(c, "tutorial video not found")
		return
	}
	id := strings.TrimSpace(strings.TrimSuffix(rawID, "/play"))
	id = strings.TrimPrefix(id, "/")
	if id == "" || utf8.RuneCountInString(id) > VideoMaxIDLen {
		response.BadRequest(c, "invalid tutorial video id")
		return
	}

	videos, err := h.listVideos(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	for _, video := range videos {
		if video.ID != id {
			continue
		}
		if h.counts == nil {
			response.InternalError(c, "tutorial play count repository is not configured")
			return
		}
		count, incrementErr := h.counts.Increment(c.Request.Context(), id)
		if incrementErr != nil {
			response.ErrorFrom(c, incrementErr)
			return
		}
		response.Success(c, gin.H{"id": id, "play_count": count})
		return
	}

	response.NotFound(c, "tutorial video not found")
}

func (h *VideoHandler) listVideos(c *gin.Context) ([]Video, error) {
	if h == nil || h.videos == nil {
		return nil, errors.New("tutorial video provider is not configured")
	}
	videos, err := h.videos(c.Request.Context())
	if err != nil {
		return nil, err
	}
	return videos, nil
}
