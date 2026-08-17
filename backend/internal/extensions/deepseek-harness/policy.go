package deepseekharness

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"

	coreservice "github.com/Wei-Shaw/sub2api/internal/service"
)

var (
	errFeatureDisabled   = errors.New("deepseek harness is disabled")
	errAPIKeyNotFound    = errors.New("api key not found")
	errAPIKeyUnavailable = errors.New("api key is not available")
	errUnsupportedGroup  = errors.New("api key group is not supported")
	errInvalidModel      = errors.New("model is not available for this api key group")
	errInvalidBaseURL    = errors.New("site api base url is invalid")
	errInvalidSession    = errors.New("deepseek harness session is invalid")
	errInvalidEvent      = errors.New("deepseek harness event is invalid")
	errInvalidEventToken = errors.New("deepseek harness event token is invalid")
)

type platformPolicy struct {
	provider     string
	displayName  string
	protocol     string
	defaultModel string
}

var platformPolicies = map[string]platformPolicy{
	"openai": {
		provider:     "sub2api-openai",
		displayName:  "Sub2API OpenAI",
		protocol:     "openai-responses",
		defaultModel: "gpt-5.6-sol",
	},
	"anthropic": {
		provider:     "sub2api-anthropic",
		displayName:  "Sub2API Claude",
		protocol:     "anthropic-messages",
		defaultModel: "claude-opus-5",
	},
	"grok": {
		provider:     "sub2api-grok",
		displayName:  "Sub2API Grok",
		protocol:     "openai-responses",
		defaultModel: "grok-4.5",
	},
	"gemini": {
		provider:     "sub2api-gemini",
		displayName:  "Sub2API Gemini",
		protocol:     "openai-completions",
		defaultModel: "gemini-3.1-pro-preview",
	},
	"antigravity": {
		provider:     "sub2api-antigravity",
		displayName:  "Sub2API Antigravity",
		protocol:     "openai-completions",
		defaultModel: "claude-opus-5",
	},
	"composite": {
		provider:     "sub2api-composite",
		displayName:  "Sub2API Composite",
		protocol:     "openai-completions",
		defaultModel: "gpt-5.6-sol",
	},
}

func buildInstallProfile(key *coreservice.APIKey, rawBaseURL, fallbackOrigin, selectedModel string) (InstallProfile, error) {
	if key == nil || key.Group == nil {
		return InstallProfile{}, errUnsupportedGroup
	}
	platform := strings.ToLower(strings.TrimSpace(key.Group.Platform))
	policy, ok := platformPolicies[platform]
	if !ok {
		return InstallProfile{}, errUnsupportedGroup
	}
	baseURL, serverURL, err := normalizeSiteURLs(rawBaseURL, fallbackOrigin)
	if err != nil {
		return InstallProfile{}, err
	}

	if platform == "composite" && !hasConfiguredModels(key.Group) {
		return InstallProfile{}, errInvalidModel
	}
	models := resolveModels(key.Group, policy.defaultModel)
	defaultModel := policy.defaultModel
	if !containsModel(models, defaultModel) {
		defaultModel = models[0].ID
	}
	selectedModel = strings.TrimSpace(selectedModel)
	if selectedModel == "" {
		selectedModel = defaultModel
	}
	if !containsModel(models, selectedModel) {
		return InstallProfile{}, errInvalidModel
	}
	protocol := policy.protocol
	if platform == "antigravity" {
		protocol = antigravityProtocol(selectedModel)
	}

	return InstallProfile{
		APIKeyID:        key.ID,
		APIKeyName:      key.Name,
		KeyHint:         maskAPIKey(key.Key),
		GroupName:       key.Group.Name,
		Platform:        platform,
		Provider:        policy.provider,
		ProviderName:    policy.displayName,
		Protocol:        protocol,
		BaseURL:         baseURL,
		DefaultModel:    defaultModel,
		SelectedModel:   selectedModel,
		AvailableModels: models,
		ServerURL:       serverURL,
	}, nil
}

func hasConfiguredModels(group *coreservice.Group) bool {
	if group == nil || !group.ModelsListConfig.Enabled {
		return false
	}
	for _, model := range group.ModelsListConfig.Models {
		if strings.TrimSpace(model) != "" {
			return true
		}
	}
	return false
}

func resolveModels(group *coreservice.Group, fallback string) []ModelOption {
	ids := make([]string, 0, len(group.ModelsListConfig.Models)+1)
	if group.ModelsListConfig.Enabled {
		ids = append(ids, group.ModelsListConfig.Models...)
	}
	if len(ids) == 0 {
		ids = append(ids, fallback)
	}
	seen := make(map[string]struct{}, len(ids))
	models := make([]ModelOption, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		models = append(models, modelMetadata(id))
	}
	if len(models) == 0 {
		models = append(models, modelMetadata(fallback))
	}
	return models
}

func modelMetadata(id string) ModelOption {
	option := ModelOption{ID: id, Name: id, ContextWindow: 200000, MaxTokens: 16384}
	switch {
	case id == "gpt-5.6-sol":
		option.Name = "GPT-5.6 Sol"
		option.ContextWindow = 1050000
		option.MaxTokens = 128000
	case id == "claude-opus-5":
		option.Name = "Claude Opus 5"
		option.ContextWindow = 1000000
		option.MaxTokens = 128000
	case id == "grok-4.5":
		option.Name = "Grok 4.5"
		option.ContextWindow = 256000
		option.MaxTokens = 65536
	case strings.HasPrefix(id, "gpt-5"):
		option.ContextWindow = 400000
		option.MaxTokens = 128000
	case strings.HasPrefix(id, "claude-"):
		option.ContextWindow = 200000
		option.MaxTokens = 64000
	case strings.HasPrefix(id, "gemini-"):
		option.ContextWindow = 1048576
		option.MaxTokens = 65536
	case strings.HasPrefix(id, "grok-"):
		option.ContextWindow = 256000
		option.MaxTokens = 65536
	}
	return option
}

func antigravityProtocol(model string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(model)), "claude-") {
		return "anthropic-messages"
	}
	return "openai-completions"
}

func normalizeSiteURLs(rawBaseURL, fallbackOrigin string) (string, string, error) {
	candidate := strings.TrimSpace(rawBaseURL)
	if candidate == "" {
		candidate = strings.TrimSpace(fallbackOrigin)
	}
	if candidate == "" {
		return "", "", errInvalidBaseURL
	}
	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", "", errInvalidBaseURL
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", "", errInvalidBaseURL
	}
	if parsed.Scheme == "http" && !isLoopbackHost(parsed.Hostname()) {
		return "", "", errInvalidBaseURL
	}
	normalizedPath := strings.TrimRight(parsed.EscapedPath(), "/")
	if normalizedPath != "" && normalizedPath != "/v1" {
		return "", "", errInvalidBaseURL
	}
	parsed.Path = ""
	parsed.RawPath = ""
	baseURL := strings.TrimRight(parsed.String(), "/")
	serverURL := (&url.URL{Scheme: parsed.Scheme, Host: parsed.Host}).String()
	return baseURL, serverURL, nil
}

func isLoopbackHost(host string) bool {
	host = strings.Trim(strings.ToLower(strings.TrimSpace(host)), "[]")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func containsModel(models []ModelOption, model string) bool {
	for i := range models {
		if models[i].ID == model {
			return true
		}
	}
	return false
}

func selectedModelOption(profile InstallProfile) (ModelOption, error) {
	for i := range profile.AvailableModels {
		if profile.AvailableModels[i].ID == profile.SelectedModel {
			return profile.AvailableModels[i], nil
		}
	}
	return ModelOption{}, errInvalidModel
}

func maskAPIKey(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 8 {
		return "********"
	}
	return "****" + value[len(value)-4:]
}

func helperDownloads() HelperDownloads {
	base := strings.TrimRight(strings.TrimSpace(helperReleaseBaseURL()), "/")
	page := helperReleasesPageURL()
	if base == "" {
		return HelperDownloads{ReleasesPage: page}
	}
	return HelperDownloads{
		WindowsAMD64: fmt.Sprintf("%s/deepseek-harness-helper-windows-amd64.exe", base),
		WindowsARM64: fmt.Sprintf("%s/deepseek-harness-helper-windows-arm64.exe", base),
		DarwinAMD64:  fmt.Sprintf("%s/deepseek-harness-helper-darwin-amd64.tar.gz", base),
		DarwinARM64:  fmt.Sprintf("%s/deepseek-harness-helper-darwin-arm64.tar.gz", base),
		LinuxAMD64:   fmt.Sprintf("%s/deepseek-harness-helper-linux-amd64.tar.gz", base),
		LinuxARM64:   fmt.Sprintf("%s/deepseek-harness-helper-linux-arm64.tar.gz", base),
		ReleasesPage: page,
	}
}

func sortedModelIDs(models []ModelOption) []string {
	ids := make([]string, 0, len(models))
	for i := range models {
		ids = append(ids, models[i].ID)
	}
	sort.Strings(ids)
	return ids
}
