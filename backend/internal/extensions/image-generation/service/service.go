package imagegeneration

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	core "github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	maxPromptRunes      = 10000
	maxGenerationCount  = 9
	apiKeyListPageSize  = 1000
	defaultImageSize    = "auto"
	defaultImageQuality = "auto"
	defaultOutputFormat = "png"
	defaultBackground   = "auto"
	defaultModeration   = "auto"

	gptImage2MinPixels    = 655360
	gptImage2MaxPixels    = 8294400
	gptImage2MaxEdge      = 3840
	gptImage2EdgeMultiple = 16
	gptImage2MaxAspect    = 3.0
)

var (
	ErrInvalidGroup       = infraerrors.BadRequest("IMAGE_GENERATION_INVALID_GROUP", "invalid image generation group")
	ErrGroupNotAllowed    = infraerrors.Forbidden("IMAGE_GENERATION_GROUP_NOT_ALLOWED", "you are not allowed to use this group")
	ErrGroupDisabled      = infraerrors.Forbidden("IMAGE_GENERATION_GROUP_DISABLED", "image generation is disabled for this group")
	ErrModelNotAvailable  = infraerrors.BadRequest("IMAGE_GENERATION_MODEL_NOT_AVAILABLE", "the selected image model is not available in this group")
	ErrPromptRequired     = infraerrors.BadRequest("IMAGE_GENERATION_PROMPT_REQUIRED", "prompt is required")
	ErrPromptTooLong      = infraerrors.BadRequest("IMAGE_GENERATION_PROMPT_TOO_LONG", "prompt is too long")
	ErrInvalidParameter   = infraerrors.BadRequest("IMAGE_GENERATION_INVALID_PARAMETER", "one or more image parameters are invalid")
	ErrImageAPIKeyMissing = infraerrors.Forbidden("IMAGE_GENERATION_API_KEY_REQUIRED", "no active API key is available for this group")
)

var (
	legacyImageSizes        = []string{"auto", "1024x1024", "1536x1024", "1024x1536"}
	gptImage2PresetSizes    = []string{"auto", "1024x1024", "1536x1024", "1024x1536", "2048x2048", "3072x2048", "2048x3072"}
	allowedImageQualities   = []string{"auto", "low", "medium", "high"}
	allowedOutputFormats    = []string{"png", "jpeg", "webp"}
	allowedBackgrounds      = []string{"auto", "opaque", "transparent"}
	allowedModerationLevels = []string{"auto", "low"}
)

// GroupProvider is the stable read-only surface used to apply user group permissions.
type GroupProvider interface {
	GetAvailableGroups(ctx context.Context, userID int64) ([]core.Group, error)
}

// PlazaProvider supplies the effective channel/model definitions.
type PlazaProvider interface {
	ListPlazaGroups(ctx context.Context) ([]core.PlazaGroup, error)
}

// APIKeyProvider is deliberately limited to the operations needed by this extension.
type APIKeyProvider interface {
	List(ctx context.Context, userID int64, params pagination.PaginationParams, filters core.APIKeyListFilters) ([]core.APIKey, *pagination.PaginationResult, error)
	GetByID(ctx context.Context, id int64) (*core.APIKey, error)
}

// Service contains all image-generation-specific policy and selection logic.
type Service struct {
	groups   GroupProvider
	plaza    PlazaProvider
	apiKeys  APIKeyProvider
	settings core.SettingRepository
}

func NewService(groups GroupProvider, plaza PlazaProvider, apiKeys APIKeyProvider, settings ...core.SettingRepository) *Service {
	var settingRepo core.SettingRepository
	if len(settings) > 0 {
		settingRepo = settings[0]
	}
	return &Service{
		groups:   groups,
		plaza:    plaza,
		apiKeys:  apiKeys,
		settings: settingRepo,
	}
}

// GenerationRequest is the browser-facing request after JSON decoding.
type GenerationRequest struct {
	GroupID           int64  `json:"group_id"`
	Model             string `json:"model"`
	Prompt            string `json:"prompt"`
	N                 int    `json:"n"`
	Size              string `json:"size"`
	Quality           string `json:"quality"`
	OutputFormat      string `json:"output_format"`
	OutputCompression *int   `json:"output_compression"`
	Background        string `json:"background"`
	Moderation        string `json:"moderation"`
}

// PreparedGeneration is an internal result. APIKey never crosses the HTTP response boundary.
type PreparedGeneration struct {
	APIKey  *core.APIKey
	Request GenerationRequest
}

type ModelOption struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Sizes               []string               `json:"sizes"`
	Qualities           []string               `json:"qualities"`
	OutputFormats       []string               `json:"output_formats"`
	Backgrounds         []string               `json:"backgrounds"`
	Moderations         []string               `json:"moderations"`
	MaxN                int                    `json:"max_n"`
	SupportsCompression bool                   `json:"supports_compression"`
	CustomSize          *CustomSizeConstraints `json:"custom_size,omitempty"`
}

type CustomSizeConstraints struct {
	MinPixels      int     `json:"min_pixels"`
	MaxPixels      int     `json:"max_pixels"`
	MaxEdge        int     `json:"max_edge"`
	EdgeMultiple   int     `json:"edge_multiple"`
	MaxAspectRatio float64 `json:"max_aspect_ratio"`
}

type GroupOption struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Platform    string        `json:"platform"`
	Models      []ModelOption `json:"models"`
}

type Defaults struct {
	Size         string `json:"size"`
	Quality      string `json:"quality"`
	OutputFormat string `json:"output_format"`
	Background   string `json:"background"`
	Moderation   string `json:"moderation"`
	N            int    `json:"n"`
}

type Options struct {
	Groups   []GroupOption `json:"groups"`
	Defaults Defaults      `json:"defaults"`
}

// GetOptions returns only active, authorized groups with configured GPT image models.
func (s *Service) GetOptions(ctx context.Context, userID int64) (Options, error) {
	if userID <= 0 || s == nil || s.groups == nil || s.plaza == nil {
		return Options{}, ErrGroupNotAllowed
	}
	available, err := s.groups.GetAvailableGroups(ctx, userID)
	if err != nil {
		return Options{}, fmt.Errorf("list available groups: %w", err)
	}
	plazaGroups, err := s.plaza.ListPlazaGroups(ctx)
	if err != nil {
		return Options{}, fmt.Errorf("list image models: %w", err)
	}

	plazaByID := make(map[int64]core.PlazaGroup, len(plazaGroups))
	for _, group := range plazaGroups {
		plazaByID[group.ID] = group
	}

	out := Options{
		Groups: make([]GroupOption, 0),
		Defaults: Defaults{
			Size:         defaultImageSize,
			Quality:      defaultImageQuality,
			OutputFormat: defaultOutputFormat,
			Background:   defaultBackground,
			Moderation:   defaultModeration,
			N:            1,
		},
	}
	for i := range available {
		group := &available[i]
		if group == nil || !group.IsActive() || !group.AllowImageGeneration || group.Platform != core.PlatformOpenAI {
			continue
		}
		plazaGroup, ok := plazaByID[group.ID]
		if !ok {
			continue
		}
		models := make([]ModelOption, 0, len(plazaGroup.Models))
		for _, model := range plazaGroup.Models {
			if model.Platform != core.PlatformOpenAI || !core.IsGPTImageGenerationModel(model.Name) {
				continue
			}
			name := strings.TrimSpace(model.Name)
			if name == "" {
				continue
			}
			modelOption := ModelOption{
				ID:                  name,
				Name:                name,
				Sizes:               cloneStrings(legacyImageSizes),
				Qualities:           cloneStrings(allowedImageQualities),
				OutputFormats:       cloneStrings(allowedOutputFormats),
				Backgrounds:         cloneStrings(allowedBackgrounds),
				Moderations:         cloneStrings(allowedModerationLevels),
				MaxN:                maxGenerationCount,
				SupportsCompression: true,
			}
			if isGPTImage2(name) {
				modelOption.Sizes = cloneStrings(gptImage2PresetSizes)
				modelOption.CustomSize = &CustomSizeConstraints{
					MinPixels:      gptImage2MinPixels,
					MaxPixels:      gptImage2MaxPixels,
					MaxEdge:        gptImage2MaxEdge,
					EdgeMultiple:   gptImage2EdgeMultiple,
					MaxAspectRatio: gptImage2MaxAspect,
				}
			}
			models = append(models, modelOption)
		}
		if len(models) == 0 {
			continue
		}
		out.Groups = append(out.Groups, GroupOption{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			Platform:    group.Platform,
			Models:      models,
		})
	}
	return out, nil
}

// Prepare validates all client-controlled values and selects a server-side key.
func (s *Service) Prepare(ctx context.Context, userID int64, req GenerationRequest) (*PreparedGeneration, error) {
	if userID <= 0 {
		return nil, ErrGroupNotAllowed
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return nil, ErrPromptRequired
	}
	if utf8.RuneCountInString(req.Prompt) > maxPromptRunes {
		return nil, ErrPromptTooLong
	}

	options, err := s.GetOptions(ctx, userID)
	if err != nil {
		return nil, err
	}
	group, modelName, groupFound, modelFound := findModelOption(options.Groups, req.GroupID, req.Model)
	if !groupFound {
		if req.GroupID <= 0 {
			return nil, ErrInvalidGroup
		}
		return nil, ErrGroupNotAllowed
	}
	if !modelFound {
		return nil, ErrModelNotAvailable
	}

	req.Prompt = strings.TrimSpace(req.Prompt)
	req.Model = modelName
	if req.N == 0 {
		req.N = 1
	}
	if req.N < 1 || req.N > maxGenerationCount {
		return nil, ErrInvalidParameter
	}
	req.Size = normalizeChoice(req.Size, defaultImageSize)
	req.Quality = normalizeChoice(req.Quality, defaultImageQuality)
	req.OutputFormat = normalizeChoice(req.OutputFormat, defaultOutputFormat)
	req.Background = normalizeChoice(req.Background, defaultBackground)
	req.Moderation = normalizeChoice(req.Moderation, defaultModeration)
	if !validImageSize(req.Model, req.Size) ||
		!containsFold(allowedImageQualities, req.Quality) ||
		!containsFold(allowedOutputFormats, req.OutputFormat) ||
		!containsFold(allowedBackgrounds, req.Background) ||
		!containsFold(allowedModerationLevels, req.Moderation) {
		return nil, ErrInvalidParameter
	}
	if req.OutputCompression != nil {
		if *req.OutputCompression < 0 || *req.OutputCompression > 100 ||
			(req.OutputFormat != "jpeg" && req.OutputFormat != "webp") {
			return nil, ErrInvalidParameter
		}
	}

	groupID := group.ID
	preferredKeyID := int64(0)
	if stored, ok := s.loadStoredConfig(ctx, userID); ok &&
		stored.ImageGroupID == groupID && strings.EqualFold(stored.ImageModel, req.Model) {
		preferredKeyID = stored.ImageAPIKeyID
	}
	selected, err := s.selectAPIKey(ctx, userID, groupID, preferredKeyID, true)
	if err != nil {
		return nil, err
	}
	if selected == nil {
		return nil, ErrImageAPIKeyMissing
	}

	return &PreparedGeneration{APIKey: selected, Request: req}, nil
}

func (s *Service) selectAPIKey(ctx context.Context, userID, groupID, preferredID int64, requireImage bool) (*core.APIKey, error) {
	if s == nil || s.apiKeys == nil {
		return nil, ErrImageAPIKeyMissing
	}
	keys, _, err := s.apiKeys.List(ctx, userID, pagination.PaginationParams{
		Page: 1, PageSize: apiKeyListPageSize, SortBy: "id", SortOrder: pagination.SortOrderAsc,
	}, core.APIKeyListFilters{Status: core.StatusAPIKeyActive, GroupID: &groupID})
	if err != nil {
		return nil, fmt.Errorf("list image generation api keys: %w", err)
	}
	ordered := make([]core.APIKey, 0, len(keys))
	if preferredID > 0 {
		for i := range keys {
			if keys[i].ID == preferredID {
				ordered = append(ordered, keys[i])
				break
			}
		}
	}
	for i := range keys {
		if keys[i].ID != preferredID {
			ordered = append(ordered, keys[i])
		}
	}
	for i := range ordered {
		candidate := &ordered[i]
		if candidate.UserID != userID || candidate.GroupID == nil || *candidate.GroupID != groupID || candidate.Status != core.StatusAPIKeyActive || candidate.IsExpired() || candidate.IsQuotaExhausted() {
			continue
		}
		hydrated, getErr := s.apiKeys.GetByID(ctx, candidate.ID)
		if getErr != nil || hydrated == nil || hydrated.User == nil || !usableKey(hydrated, userID, groupID, requireImage) {
			continue
		}
		return hydrated, nil
	}
	return nil, nil
}

func findModelOption(groups []GroupOption, groupID int64, model string) (GroupOption, string, bool, bool) {
	model = strings.TrimSpace(model)
	if groupID <= 0 || model == "" {
		return GroupOption{}, "", false, false
	}
	for _, group := range groups {
		if group.ID != groupID {
			continue
		}
		for _, candidate := range group.Models {
			if strings.EqualFold(candidate.Name, model) {
				return group, candidate.Name, true, true
			}
		}
		return group, "", true, false
	}
	return GroupOption{}, "", false, false
}

func normalizeChoice(value, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	return value
}

func containsFold(values []string, value string) bool {
	for _, item := range values {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

func isGPTImage2(model string) bool {
	return strings.EqualFold(strings.TrimSpace(model), "gpt-image-2")
}

func validImageSize(model, size string) bool {
	if !isGPTImage2(model) {
		return containsFold(legacyImageSizes, size)
	}
	if strings.EqualFold(size, "auto") {
		return true
	}

	parts := strings.Split(strings.ToLower(strings.TrimSpace(size)), "x")
	if len(parts) != 2 {
		return false
	}
	width, widthErr := strconv.Atoi(parts[0])
	height, heightErr := strconv.Atoi(parts[1])
	if widthErr != nil || heightErr != nil || width <= 0 || height <= 0 {
		return false
	}
	if width%gptImage2EdgeMultiple != 0 || height%gptImage2EdgeMultiple != 0 || width > gptImage2MaxEdge || height > gptImage2MaxEdge {
		return false
	}
	pixels := int64(width) * int64(height)
	if pixels < gptImage2MinPixels || pixels > gptImage2MaxPixels {
		return false
	}
	longEdge, shortEdge := width, height
	if longEdge < shortEdge {
		longEdge, shortEdge = shortEdge, longEdge
	}
	return float64(longEdge)/float64(shortEdge) <= gptImage2MaxAspect
}

func cloneStrings(values []string) []string {
	return append([]string(nil), values...)
}
