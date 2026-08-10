package tutorials

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	DocumentURL = "https://doc.aiprox.net/doc"

	VideosMaxItems   = 100
	VideoMaxIDLen    = 64
	VideoMaxTitleLen = 200
	VideoMaxURLLen   = 2048
)

// Video is the persisted and public contract for a tutorial video.
type Video struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CoverURL  string `json:"cover_url"`
	VideoURL  string `json:"video_url"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sort_order"`
}

// NormalizeVideosJSON validates and canonicalizes administrator-provided JSON.
func NormalizeVideosJSON(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "[]", nil
	}
	var videos []Video
	if err := json.Unmarshal([]byte(raw), &videos); err != nil {
		return "", fmt.Errorf("tutorial videos must be a valid JSON array: %w", err)
	}
	if videos == nil {
		videos = []Video{}
	}
	if len(videos) > VideosMaxItems {
		return "", fmt.Errorf("too many tutorial videos (max %d)", VideosMaxItems)
	}
	seen := make(map[string]struct{}, len(videos))
	for i := range videos {
		video := &videos[i]
		video.ID = strings.TrimSpace(video.ID)
		video.Title = strings.TrimSpace(video.Title)
		video.CoverURL = strings.TrimSpace(video.CoverURL)
		video.VideoURL = strings.TrimSpace(video.VideoURL)
		if video.ID == "" || utf8.RuneCountInString(video.ID) > VideoMaxIDLen {
			return "", fmt.Errorf("tutorial video %d id is required and must be at most %d characters", i, VideoMaxIDLen)
		}
		if _, ok := seen[video.ID]; ok {
			return "", fmt.Errorf("duplicate tutorial video id: %s", video.ID)
		}
		seen[video.ID] = struct{}{}
		if video.Title == "" || utf8.RuneCountInString(video.Title) > VideoMaxTitleLen {
			return "", fmt.Errorf("tutorial video %s title is required and must be at most %d characters", video.ID, VideoMaxTitleLen)
		}
		if video.VideoURL == "" {
			return "", fmt.Errorf("tutorial video %s video_url is required", video.ID)
		}
		if err := validateURL(video.VideoURL); err != nil {
			return "", fmt.Errorf("tutorial video %s video_url: %w", video.ID, err)
		}
		if video.CoverURL != "" {
			if err := validateCoverURL(video.CoverURL); err != nil {
				return "", fmt.Errorf("tutorial video %s cover_url: %w", video.ID, err)
			}
		}
		if video.SortOrder < 0 || video.SortOrder > 100000 {
			return "", fmt.Errorf("tutorial video %s sort_order must be between 0 and 100000", video.ID)
		}
	}
	sort.SliceStable(videos, func(i, j int) bool { return videos[i].SortOrder < videos[j].SortOrder })
	b, err := json.Marshal(videos)
	if err != nil {
		return "", fmt.Errorf("marshal tutorial videos: %w", err)
	}
	return string(b), nil
}

func validateURL(raw string) error {
	if len(raw) > VideoMaxURLLen {
		return fmt.Errorf("URL is too long (max %d characters)", VideoMaxURLLen)
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("URL must be an absolute http(s) URL")
	}
	return nil
}

func validateCoverURL(raw string) error {
	if strings.HasPrefix(raw, tutorialCoverURLPrefix) {
		filename := strings.TrimPrefix(raw, tutorialCoverURLPrefix)
		if tutorialCoverNamePattern.MatchString(filename) {
			return nil
		}
		return fmt.Errorf("local cover URL contains an invalid filename")
	}
	return validateURL(raw)
}

// ParsePublicVideos returns enabled, valid videos and fails closed on corrupt data.
func ParsePublicVideos(raw string) []Video {
	var videos []Video
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &videos); err != nil || videos == nil {
		return []Video{}
	}
	filtered := make([]Video, 0, len(videos))
	for _, video := range videos {
		if !video.Enabled {
			continue
		}
		normalized, err := NormalizeVideosJSON(mustMarshalVideo(video))
		if err == nil {
			var clean []Video
			if json.Unmarshal([]byte(normalized), &clean) == nil && len(clean) == 1 {
				filtered = append(filtered, clean[0])
			}
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].SortOrder < filtered[j].SortOrder })
	return filtered
}

func mustMarshalVideo(video Video) string {
	b, _ := json.Marshal([]Video{video})
	return string(b)
}
