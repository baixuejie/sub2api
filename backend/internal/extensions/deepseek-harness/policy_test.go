package deepseekharness

import (
	"testing"

	coreservice "github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDeepSeekHarnessNormalizeSiteURLsAppendsV1ExactlyOnce(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		fallback string
		wantBase string
		wantSite string
	}{
		{name: "append", base: "https://example.com/", wantBase: "https://example.com/v1", wantSite: "https://example.com"},
		{name: "existing", base: "https://example.com/v1/", wantBase: "https://example.com/v1", wantSite: "https://example.com"},
		{name: "fallback", fallback: "http://127.0.0.1:8080", wantBase: "http://127.0.0.1:8080/v1", wantSite: "http://127.0.0.1:8080"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base, site, err := normalizeSiteURLs(test.base, test.fallback)
			require.NoError(t, err)
			require.Equal(t, test.wantBase, base)
			require.Equal(t, test.wantSite, site)
		})
	}
}

func TestDeepSeekHarnessGroupModelListControlsDefault(t *testing.T) {
	key := activeOpenAIKey()
	key.Group.ModelsListConfig = coreservice.GroupModelsListConfig{
		Enabled: true,
		Models:  []string{" custom-model ", "custom-model", "gpt-5.5"},
	}
	profile, err := buildInstallProfile(key, "https://example.com", "", "")
	require.NoError(t, err)
	require.Equal(t, "custom-model", profile.DefaultModel)
	require.Equal(t, []string{"custom-model", "gpt-5.5"}, sortedModelIDs(profile.AvailableModels))
}

func TestDeepSeekHarnessCompositeRequiresConfiguredModels(t *testing.T) {
	key := activeOpenAIKey()
	key.Group.Platform = "composite"
	_, err := buildInstallProfile(key, "https://example.com", "", "")
	require.ErrorIs(t, err, errInvalidModel)

	key.Group.ModelsListConfig = coreservice.GroupModelsListConfig{Enabled: true, Models: []string{"public-model"}}
	profile, err := buildInstallProfile(key, "https://example.com", "", "")
	require.NoError(t, err)
	require.Equal(t, "public-model", profile.SelectedModel)
}

func TestDeepSeekHarnessAntigravityProtocolFollowsSelectedModel(t *testing.T) {
	key := activeOpenAIKey()
	key.Group.Platform = "antigravity"
	key.Group.ModelsListConfig = coreservice.GroupModelsListConfig{
		Enabled: true,
		Models:  []string{"claude-opus-5", "gemini-3.1-pro-preview"},
	}
	claude, err := buildInstallProfile(key, "https://example.com", "", "claude-opus-5")
	require.NoError(t, err)
	require.Equal(t, "anthropic-messages", claude.Protocol)

	gemini, err := buildInstallProfile(key, "https://example.com", "", "gemini-3.1-pro-preview")
	require.NoError(t, err)
	require.Equal(t, "openai-completions", gemini.Protocol)
}

func TestDeepSeekHarnessRejectsUnsafeBaseURL(t *testing.T) {
	for _, raw := range []string{"file:///tmp/api", "https://user:pass@example.com", "http://example.com", "https://example.com/api/v1", "https://example.com/v1?token=x", "://bad"} {
		_, _, err := normalizeSiteURLs(raw, "")
		require.ErrorIs(t, err, errInvalidBaseURL)
	}
}
