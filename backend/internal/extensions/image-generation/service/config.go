package imagegeneration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	core "github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	imageGenerationUserSettingPrefix = "extension.image_generation.user."
	defaultImageModel                = "gpt-image-2"
	defaultPromptModel               = "gpt-4.1-mini"
	defaultConfigSize                = "1024x1024"
	defaultConfigN                   = 1
)

var (
	ErrConfigUnavailable = infraerrors.InternalServer("IMAGE_GENERATION_CONFIG_UNAVAILABLE", "image generation configuration is unavailable")
	ErrConfigInvalid     = infraerrors.BadRequest("IMAGE_GENERATION_CONFIG_INVALID", "image generation configuration is invalid")
	ErrPromptConfig      = infraerrors.Forbidden("IMAGE_GENERATION_PROMPT_CONFIG_REQUIRED", "prompt optimization is not configured")
)

// UserImageConfig is persisted per user. API keys are represented only by IDs;
// the credential value is always resolved from the server-side API key store.
type UserImageConfig struct {
	Version        int    `json:"version"`
	PromptGroupID  int64  `json:"prompt_group_id"`
	PromptModel    string `json:"prompt_model"`
	PromptAPIKeyID int64  `json:"prompt_api_key_id"`
	ImageGroupID   int64  `json:"image_group_id"`
	ImageModel     string `json:"image_model"`
	ImageAPIKeyID  int64  `json:"image_api_key_id"`
	DefaultSize    string `json:"default_size"`
	DefaultN       int    `json:"default_n"`
}

type ConfigModelOption struct {
	Name string `json:"name"`
}

type ConfigGroupOption struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Platform    string              `json:"platform"`
	Models      []ConfigModelOption `json:"models"`
}

type ConfigAPIKeyOption struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	MaskedKey    string `json:"masked_key"`
	GroupID      int64  `json:"group_id"`
	GroupName    string `json:"group_name"`
	ImageEnabled bool   `json:"image_enabled"`
	Status       string `json:"status"`
}

type ConfigOptions struct {
	Config       UserImageConfig      `json:"config"`
	PromptGroups []ConfigGroupOption  `json:"prompt_groups"`
	ImageGroups  []ConfigGroupOption  `json:"image_groups"`
	APIKeys      []ConfigAPIKeyOption `json:"api_keys"`
}

type PromptOptimizationRequest struct {
	Prompt string `json:"prompt"`
}

type PreparedPrompt struct {
	APIKey  *core.APIKey
	GroupID int64
	Model   string
	Prompt  string
}

type configCatalog struct {
	options      ConfigOptions
	imageModels  map[int64]map[string]struct{}
	promptModels map[int64]map[string]struct{}
	keys         map[int64]*core.APIKey
}

func userSettingKey(userID int64) string {
	return imageGenerationUserSettingPrefix + strconv.FormatInt(userID, 10)
}

// GetConfigOptions returns model and key choices visible to the current user,
// plus a normalized persisted configuration. Secrets are intentionally reduced
// to a short mask before crossing the HTTP boundary.
func (s *Service) GetConfigOptions(ctx context.Context, userID int64) (ConfigOptions, error) {
	catalog, err := s.buildConfigCatalog(ctx, userID)
	if err != nil {
		return ConfigOptions{}, err
	}
	catalog.options.Config = s.normalizeStoredConfig(ctx, userID, catalog)
	return catalog.options, nil
}

// SaveConfig validates every group/model/key relationship against fresh server
// state before persisting the user-scoped JSON setting.
func (s *Service) SaveConfig(ctx context.Context, userID int64, requested UserImageConfig) (ConfigOptions, error) {
	if userID <= 0 || s == nil || s.settings == nil {
		return ConfigOptions{}, ErrConfigUnavailable
	}
	if requested.DefaultN < 1 || requested.DefaultN > maxGenerationCount {
		return ConfigOptions{}, ErrConfigInvalid
	}
	catalog, err := s.buildConfigCatalog(ctx, userID)
	if err != nil {
		return ConfigOptions{}, err
	}
	requested = normalizeConfig(requested)
	if err := validateConfig(requested, catalog); err != nil {
		return ConfigOptions{}, err
	}
	raw, err := json.Marshal(requested)
	if err != nil {
		return ConfigOptions{}, fmt.Errorf("marshal image generation config: %w", err)
	}
	if err := s.settings.Set(ctx, userSettingKey(userID), string(raw)); err != nil {
		return ConfigOptions{}, fmt.Errorf("save image generation config: %w", err)
	}
	catalog.options.Config = requested
	return catalog.options, nil
}

// PreparePrompt resolves the configured prompt model/key for the optimization
// endpoint. It never returns a credential to the caller-facing handler.
func (s *Service) PreparePrompt(ctx context.Context, userID int64, prompt string) (*PreparedPrompt, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, ErrPromptRequired
	}
	if len([]rune(prompt)) > maxPromptRunes {
		return nil, ErrPromptTooLong
	}
	catalog, err := s.buildConfigCatalog(ctx, userID)
	if err != nil {
		return nil, err
	}
	cfg := s.normalizeStoredConfig(ctx, userID, catalog)
	if cfg.PromptGroupID <= 0 || cfg.PromptAPIKeyID <= 0 || strings.TrimSpace(cfg.PromptModel) == "" {
		return nil, ErrPromptConfig
	}
	key := catalog.keys[cfg.PromptAPIKeyID]
	if key == nil || key.GroupID == nil || *key.GroupID != cfg.PromptGroupID {
		return nil, ErrPromptConfig
	}
	hydrated, err := s.apiKeys.GetByID(ctx, key.ID)
	if err != nil || hydrated == nil || hydrated.ID != key.ID || !usableKey(hydrated, userID, cfg.PromptGroupID, false) {
		return nil, ErrPromptConfig
	}
	return &PreparedPrompt{APIKey: hydrated, GroupID: cfg.PromptGroupID, Model: cfg.PromptModel, Prompt: prompt}, nil
}

func (s *Service) buildConfigCatalog(ctx context.Context, userID int64) (*configCatalog, error) {
	if userID <= 0 || s == nil || s.groups == nil || s.plaza == nil {
		return nil, ErrGroupNotAllowed
	}
	available, err := s.groups.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list available groups: %w", err)
	}
	plazaGroups, err := s.plaza.ListPlazaGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("list image models: %w", err)
	}
	plazaByID := make(map[int64]core.PlazaGroup, len(plazaGroups))
	for _, group := range plazaGroups {
		plazaByID[group.ID] = group
	}
	catalog := &configCatalog{
		options: ConfigOptions{
			PromptGroups: make([]ConfigGroupOption, 0),
			ImageGroups:  make([]ConfigGroupOption, 0),
			APIKeys:      make([]ConfigAPIKeyOption, 0),
		},
		imageModels:  make(map[int64]map[string]struct{}),
		promptModels: make(map[int64]map[string]struct{}),
		keys:         make(map[int64]*core.APIKey),
	}
	for i := range available {
		group := &available[i]
		if group == nil || !group.IsActive() || group.Platform != core.PlatformOpenAI {
			continue
		}
		imageOption := ConfigGroupOption{ID: group.ID, Name: group.Name, Description: group.Description, Platform: group.Platform}
		promptOption := imageOption
		imageSeen := make(map[string]struct{})
		promptSeen := make(map[string]struct{})
		plazaGroup, plazaOK := plazaByID[group.ID]
		if plazaOK {
			for _, model := range plazaGroup.Models {
				if model.Platform != core.PlatformOpenAI {
					continue
				}
				name := strings.TrimSpace(model.Name)
				if name == "" {
					continue
				}
				if core.IsGPTImageGenerationModel(name) {
					if !group.AllowImageGeneration {
						continue
					}
					key := strings.ToLower(name)
					if _, seen := imageSeen[key]; !seen {
						imageSeen[key] = struct{}{}
						imageOption.Models = append(imageOption.Models, ConfigModelOption{Name: name})
					}
					continue
				}
				key := strings.ToLower(name)
				if _, seen := promptSeen[key]; !seen {
					promptSeen[key] = struct{}{}
					promptOption.Models = append(promptOption.Models, ConfigModelOption{Name: name})
				}
			}
		}
		// A text-only OpenAI group may have no channel-backed plaza models. Its
		// saved group model definition is still usable for prompt optimization.
		if len(promptSeen) == 0 && !group.AllowImageGeneration {
			for _, name := range group.ModelsListConfig.Models {
				name = strings.TrimSpace(name)
				if name == "" || core.IsGPTImageGenerationModel(name) {
					continue
				}
				key := strings.ToLower(name)
				if _, seen := promptSeen[key]; seen {
					continue
				}
				promptSeen[key] = struct{}{}
				promptOption.Models = append(promptOption.Models, ConfigModelOption{Name: name})
			}
		}
		if len(imageOption.Models) > 0 {
			catalog.options.ImageGroups = append(catalog.options.ImageGroups, imageOption)
			catalog.imageModels[group.ID] = imageSeen
		}
		if len(promptOption.Models) > 0 {
			catalog.options.PromptGroups = append(catalog.options.PromptGroups, promptOption)
			catalog.promptModels[group.ID] = promptSeen
		}
	}
	if s.apiKeys != nil {
		keys, _, listErr := s.apiKeys.List(ctx, userID, pagination.PaginationParams{
			Page: 1, PageSize: apiKeyListPageSize, SortBy: "id", SortOrder: pagination.SortOrderAsc,
		}, core.APIKeyListFilters{Status: core.StatusAPIKeyActive})
		if listErr != nil {
			return nil, fmt.Errorf("list image generation api keys: %w", listErr)
		}
		for i := range keys {
			candidate := &keys[i]
			if candidate.UserID != userID || candidate.GroupID == nil || candidate.Status != core.StatusAPIKeyActive || candidate.IsExpired() || candidate.IsQuotaExhausted() {
				continue
			}
			hydrated, getErr := s.apiKeys.GetByID(ctx, candidate.ID)
			if getErr != nil || hydrated == nil || hydrated.ID != candidate.ID || !usableKey(hydrated, userID, *candidate.GroupID, false) || hydrated.Group == nil {
				continue
			}
			groupID := *hydrated.GroupID
			_, promptOK := catalog.promptModels[groupID]
			_, imageOK := catalog.imageModels[groupID]
			if !promptOK && !imageOK {
				continue
			}
			catalog.keys[hydrated.ID] = hydrated
			catalog.options.APIKeys = append(catalog.options.APIKeys, ConfigAPIKeyOption{
				ID: hydrated.ID, Name: displayKeyName(hydrated), MaskedKey: maskAPIKey(hydrated.Key),
				GroupID: groupID, GroupName: hydrated.Group.Name, ImageEnabled: imageOK && hydrated.Group.AllowImageGeneration,
				Status: hydrated.Status,
			})
		}
	}
	return catalog, nil
}

func (s *Service) normalizeStoredConfig(ctx context.Context, userID int64, catalog *configCatalog) UserImageConfig {
	cfg := UserImageConfig{Version: 1, DefaultSize: defaultConfigSize, DefaultN: defaultConfigN}
	if stored, ok := s.loadStoredConfig(ctx, userID); ok {
		cfg = stored
	}
	if catalog == nil {
		return cfg
	}
	if !validModel(catalog.imageModels[cfg.ImageGroupID], cfg.ImageModel) {
		cfg.ImageGroupID, cfg.ImageModel = preferredImageSelection(catalog)
	}
	if !validKeyForGroup(catalog, cfg.ImageAPIKeyID, cfg.ImageGroupID, true) {
		cfg.ImageAPIKeyID = firstKeyForGroup(catalog, cfg.ImageGroupID, true)
	}
	if !validModel(catalog.promptModels[cfg.PromptGroupID], cfg.PromptModel) {
		cfg.PromptGroupID, cfg.PromptModel = preferredPromptSelection(catalog)
	}
	if !validKeyForGroup(catalog, cfg.PromptAPIKeyID, cfg.PromptGroupID, false) {
		cfg.PromptAPIKeyID = firstKeyForGroup(catalog, cfg.PromptGroupID, false)
	}
	return cfg
}

func (s *Service) loadStoredConfig(ctx context.Context, userID int64) (UserImageConfig, bool) {
	if s == nil || s.settings == nil || userID <= 0 {
		return UserImageConfig{}, false
	}
	raw, err := s.settings.GetValue(ctx, userSettingKey(userID))
	if err != nil {
		if errors.Is(err, core.ErrSettingNotFound) {
			return UserImageConfig{}, false
		}
		return UserImageConfig{}, false
	}
	var stored UserImageConfig
	if json.Unmarshal([]byte(raw), &stored) != nil {
		return UserImageConfig{}, false
	}
	return normalizeConfig(stored), true
}

func normalizeConfig(cfg UserImageConfig) UserImageConfig {
	cfg.Version = 1
	cfg.PromptModel = strings.TrimSpace(cfg.PromptModel)
	cfg.ImageModel = strings.TrimSpace(cfg.ImageModel)
	cfg.DefaultSize = strings.ToLower(strings.TrimSpace(cfg.DefaultSize))
	if cfg.DefaultSize == "" {
		cfg.DefaultSize = defaultConfigSize
	}
	if cfg.DefaultN < 1 {
		cfg.DefaultN = defaultConfigN
	}
	if cfg.DefaultN > maxGenerationCount {
		cfg.DefaultN = maxGenerationCount
	}
	return cfg
}

func validateConfig(cfg UserImageConfig, catalog *configCatalog) error {
	if catalog == nil {
		return ErrConfigInvalid
	}
	if cfg.ImageGroupID != 0 || cfg.ImageModel != "" || cfg.ImageAPIKeyID != 0 {
		if !validModel(catalog.imageModels[cfg.ImageGroupID], cfg.ImageModel) || !validKeyForGroup(catalog, cfg.ImageAPIKeyID, cfg.ImageGroupID, true) {
			return ErrConfigInvalid
		}
		if !validImageSize(cfg.ImageModel, cfg.DefaultSize) {
			return ErrConfigInvalid
		}
	}
	if cfg.PromptGroupID != 0 || cfg.PromptModel != "" || cfg.PromptAPIKeyID != 0 {
		if !validModel(catalog.promptModels[cfg.PromptGroupID], cfg.PromptModel) || !validKeyForGroup(catalog, cfg.PromptAPIKeyID, cfg.PromptGroupID, false) {
			return ErrConfigInvalid
		}
	}
	if cfg.DefaultN < 1 || cfg.DefaultN > maxGenerationCount || !validImageSize(cfg.ImageModel, cfg.DefaultSize) && cfg.ImageModel != "" {
		return ErrConfigInvalid
	}
	return nil
}

func preferredImageSelection(catalog *configCatalog) (int64, string) {
	for _, group := range catalog.options.ImageGroups {
		for _, model := range group.Models {
			if strings.EqualFold(model.Name, defaultImageModel) {
				return group.ID, model.Name
			}
		}
	}
	if len(catalog.options.ImageGroups) > 0 && len(catalog.options.ImageGroups[0].Models) > 0 {
		return catalog.options.ImageGroups[0].ID, catalog.options.ImageGroups[0].Models[0].Name
	}
	return 0, ""
}

func preferredPromptSelection(catalog *configCatalog) (int64, string) {
	for _, group := range catalog.options.PromptGroups {
		for _, model := range group.Models {
			if strings.EqualFold(model.Name, defaultPromptModel) {
				return group.ID, model.Name
			}
		}
	}
	if len(catalog.options.PromptGroups) > 0 && len(catalog.options.PromptGroups[0].Models) > 0 {
		return catalog.options.PromptGroups[0].ID, catalog.options.PromptGroups[0].Models[0].Name
	}
	return 0, ""
}

func validModel(models map[string]struct{}, model string) bool {
	if len(models) == 0 || strings.TrimSpace(model) == "" {
		return false
	}
	_, ok := models[strings.ToLower(strings.TrimSpace(model))]
	return ok
}

func validKeyForGroup(catalog *configCatalog, keyID, groupID int64, imageOnly bool) bool {
	if keyID <= 0 || groupID <= 0 || catalog == nil {
		return false
	}
	key := catalog.keys[keyID]
	if key == nil || key.GroupID == nil || *key.GroupID != groupID || key.Group == nil {
		return false
	}
	if imageOnly && !key.Group.AllowImageGeneration {
		return false
	}
	return true
}

func firstKeyForGroup(catalog *configCatalog, groupID int64, imageOnly bool) int64 {
	if catalog == nil || groupID <= 0 {
		return 0
	}
	for _, option := range catalog.options.APIKeys {
		if option.GroupID == groupID && (!imageOnly || option.ImageEnabled) {
			return option.ID
		}
	}
	return 0
}

func displayKeyName(key *core.APIKey) string {
	if key == nil {
		return "API Key"
	}
	name := strings.TrimSpace(key.Name)
	if name != "" {
		return name
	}
	return fmt.Sprintf("API Key #%d", key.ID)
}

func maskAPIKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return "****" + value[len(value)-4:]
}

func usableKey(key *core.APIKey, userID, groupID int64, requireImage bool) bool {
	if key == nil || key.ID <= 0 || key.UserID != userID || key.User == nil || key.GroupID == nil || *key.GroupID != groupID || key.Key == "" || !key.IsActive() || key.IsExpired() || key.IsQuotaExhausted() || key.Group == nil || !key.Group.IsActive() {
		return false
	}
	if requireImage && !key.Group.AllowImageGeneration {
		return false
	}
	return true
}
