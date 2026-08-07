package imagegeneration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	core "github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type configFakeSettings struct {
	values map[string]string
	setKey string
	setRaw string
}

func (f *configFakeSettings) Get(context.Context, string) (*core.Setting, error) {
	return nil, core.ErrSettingNotFound
}

func (f *configFakeSettings) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := f.values[key]; ok {
		return value, nil
	}
	return "", core.ErrSettingNotFound
}

func (f *configFakeSettings) Set(_ context.Context, key, value string) error {
	if f.values == nil {
		f.values = make(map[string]string)
	}
	f.values[key] = value
	f.setKey = key
	f.setRaw = value
	return nil
}

func (f *configFakeSettings) GetMultiple(context.Context, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (f *configFakeSettings) SetMultiple(context.Context, map[string]string) error { return nil }

func (f *configFakeSettings) GetAll(context.Context) (map[string]string, error) {
	return f.values, nil
}

func (f *configFakeSettings) Delete(_ context.Context, key string) error {
	delete(f.values, key)
	return nil
}

type configFakeKeys struct {
	listed   []core.APIKey
	hydrated map[int64]*core.APIKey
}

func (f *configFakeKeys) List(_ context.Context, _ int64, _ pagination.PaginationParams, filters core.APIKeyListFilters) ([]core.APIKey, *pagination.PaginationResult, error) {
	items := make([]core.APIKey, 0, len(f.listed))
	for _, key := range f.listed {
		if filters.GroupID != nil && (key.GroupID == nil || *key.GroupID != *filters.GroupID) {
			continue
		}
		items = append(items, key)
	}
	return items, &pagination.PaginationResult{Total: int64(len(items))}, nil
}

func (f *configFakeKeys) GetByID(_ context.Context, id int64) (*core.APIKey, error) {
	return f.hydrated[id], nil
}

func configTestKey(id, userID, groupID int64, name, secret string, imageEnabled bool) *core.APIKey {
	return &core.APIKey{
		ID: id, UserID: userID, Name: name, Key: secret, GroupID: int64Ptr(groupID), Status: core.StatusAPIKeyActive,
		User:  &core.User{ID: userID, Status: core.StatusActive},
		Group: &core.Group{ID: groupID, Name: "OpenAI", Platform: core.PlatformOpenAI, Status: core.StatusActive, AllowImageGeneration: imageEnabled, Hydrated: true},
	}
}

func newConfigTestService(settings *configFakeSettings) (*Service, *configFakeKeys) {
	first := configTestKey(10, 42, 1, "Primary", "sk-primary-secret-1234", true)
	second := configTestKey(11, 42, 1, "Secondary", "sk-secondary-secret-5678", true)
	keys := &configFakeKeys{
		listed: []core.APIKey{
			{ID: first.ID, UserID: first.UserID, GroupID: first.GroupID, Status: first.Status},
			{ID: second.ID, UserID: second.UserID, GroupID: second.GroupID, Status: second.Status},
		},
		hydrated: map[int64]*core.APIKey{first.ID: first, second.ID: second},
	}
	svc := NewService(
		&fakeGroups{groups: []core.Group{imageGroup(1, "OpenAI", true)}},
		&fakePlaza{groups: []core.PlazaGroup{imagePlazaGroup(1, "gpt-image-2", "gpt-4.1-mini", "gpt-5.4")}},
		keys,
		settings,
	)
	return svc, keys
}

func TestImageGenerationConfigDefaultsAndNeverReturnsPlaintextKeys(t *testing.T) {
	svc, _ := newConfigTestService(&configFakeSettings{})
	options, err := svc.GetConfigOptions(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, "gpt-image-2", options.Config.ImageModel)
	require.Equal(t, "gpt-4.1-mini", options.Config.PromptModel)
	require.Equal(t, 1, options.Config.DefaultN)
	require.Equal(t, "1024x1024", options.Config.DefaultSize)
	require.Len(t, options.APIKeys, 2)
	require.Equal(t, "****1234", options.APIKeys[0].MaskedKey)

	raw, err := json.Marshal(options)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "sk-primary-secret")
	require.NotContains(t, string(raw), "sk-secondary-secret")
}

func TestImageGenerationConfigPersistsOnlyIDsAndUsesPreferredImageKey(t *testing.T) {
	settings := &configFakeSettings{}
	svc, keys := newConfigTestService(settings)
	requested := UserImageConfig{
		PromptGroupID: 1, PromptModel: "gpt-5.4", PromptAPIKeyID: 10,
		ImageGroupID: 1, ImageModel: "gpt-image-2", ImageAPIKeyID: 11,
		DefaultSize: "2048x2048", DefaultN: 6,
	}
	options, err := svc.SaveConfig(context.Background(), 42, requested)
	require.NoError(t, err)
	require.Equal(t, userSettingKey(42), settings.setKey)
	require.NotContains(t, settings.setRaw, keys.hydrated[10].Key)
	require.NotContains(t, settings.setRaw, keys.hydrated[11].Key)
	require.Equal(t, int64(11), options.Config.ImageAPIKeyID)

	prepared, err := svc.Prepare(context.Background(), 42, GenerationRequest{
		GroupID: 1, Model: "gpt-image-2", Prompt: "draw", N: 2, Size: "2048x2048",
	})
	require.NoError(t, err)
	require.Equal(t, int64(11), prepared.APIKey.ID)

	prompt, err := svc.PreparePrompt(context.Background(), 42, "  improve this prompt  ")
	require.NoError(t, err)
	require.Equal(t, int64(10), prompt.APIKey.ID)
	require.Equal(t, "gpt-5.4", prompt.Model)
	require.Equal(t, "improve this prompt", prompt.Prompt)
}

func TestImageGenerationConfigRejectsInvalidCountAndCrossGroupKey(t *testing.T) {
	settings := &configFakeSettings{}
	svc, _ := newConfigTestService(settings)

	_, err := svc.SaveConfig(context.Background(), 42, UserImageConfig{
		ImageGroupID: 1, ImageModel: "gpt-image-2", ImageAPIKeyID: 11,
		DefaultSize: "1024x1024", DefaultN: 10,
	})
	require.ErrorIs(t, err, ErrConfigInvalid)

	_, err = svc.SaveConfig(context.Background(), 42, UserImageConfig{
		ImageGroupID: 2, ImageModel: "gpt-image-2", ImageAPIKeyID: 11,
		DefaultSize: "1024x1024", DefaultN: 1,
	})
	require.ErrorIs(t, err, ErrConfigInvalid)
	require.Empty(t, settings.setRaw)
}

func TestImageGenerationConfigRejectsHydratedKeyWithDifferentID(t *testing.T) {
	settings := &configFakeSettings{}
	svc, keys := newConfigTestService(settings)
	keys.hydrated[10].ID = 99
	keys.hydrated[11].ID = 111
	settings.values = map[string]string{
		userSettingKey(42): `{"version":1,"prompt_group_id":1,"prompt_model":"gpt-4.1-mini","prompt_api_key_id":10}`,
	}

	_, err := svc.PreparePrompt(context.Background(), 42, "draw a mountain")
	require.ErrorIs(t, err, ErrPromptConfig)
}
