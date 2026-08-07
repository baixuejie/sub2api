package imagegeneration

import (
	"context"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	core "github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type fakeGroups struct {
	groups []core.Group
}

func (f *fakeGroups) GetAvailableGroups(context.Context, int64) ([]core.Group, error) {
	return f.groups, nil
}

type fakePlaza struct {
	groups []core.PlazaGroup
}

func (f *fakePlaza) ListPlazaGroups(context.Context) ([]core.PlazaGroup, error) {
	return f.groups, nil
}

type fakeKeys struct {
	listed   []core.APIKey
	hydrated *core.APIKey
}

func (f *fakeKeys) List(context.Context, int64, pagination.PaginationParams, core.APIKeyListFilters) ([]core.APIKey, *pagination.PaginationResult, error) {
	return f.listed, &pagination.PaginationResult{Total: int64(len(f.listed))}, nil
}

func (f *fakeKeys) GetByID(context.Context, int64) (*core.APIKey, error) {
	return f.hydrated, nil
}

func imageGroup(id int64, name string, allow bool) core.Group {
	return core.Group{ID: id, Name: name, Platform: core.PlatformOpenAI, Status: core.StatusActive, AllowImageGeneration: allow, Hydrated: true}
}

func imagePlazaGroup(id int64, models ...string) core.PlazaGroup {
	plazaModels := make([]core.PlazaModel, 0, len(models))
	for _, model := range models {
		plazaModels = append(plazaModels, core.PlazaModel{Name: model, Platform: core.PlatformOpenAI})
	}
	return core.PlazaGroup{ID: id, Name: "group", Platform: core.PlatformOpenAI, Models: plazaModels}
}

func newTestService(keys *fakeKeys) *Service {
	return NewService(
		&fakeGroups{groups: []core.Group{
			imageGroup(1, "enabled", true),
			imageGroup(2, "disabled", false),
		}},
		&fakePlaza{groups: []core.PlazaGroup{
			imagePlazaGroup(1, "gpt-image-2", "gpt-5.4-mini"),
			imagePlazaGroup(2, "gpt-image-1"),
		}},
		keys,
	)
}

func TestImageGenerationServiceGetOptionsFiltersGroupsAndModels(t *testing.T) {
	svc := newTestService(&fakeKeys{})
	options, err := svc.GetOptions(context.Background(), 42)
	require.NoError(t, err)
	require.Len(t, options.Groups, 1)
	require.Equal(t, int64(1), options.Groups[0].ID)
	require.Len(t, options.Groups[0].Models, 1)
	require.Equal(t, "gpt-image-2", options.Groups[0].Models[0].Name)
	require.Equal(t, 9, options.Groups[0].Models[0].MaxN)
	require.Equal(t, []string{"auto", "1024x1024", "1536x1024", "1024x1536", "2048x2048", "3072x2048", "2048x3072"}, options.Groups[0].Models[0].Sizes)
	require.NotNil(t, options.Groups[0].Models[0].CustomSize)
	require.Equal(t, gptImage2MaxEdge, options.Groups[0].Models[0].CustomSize.MaxEdge)
	require.Equal(t, "auto", options.Defaults.Size)
}

func TestImageGenerationServicePrepareRejectsUnauthorizedModelAndInvalidParameters(t *testing.T) {
	key := &core.APIKey{
		ID: 9, UserID: 42, Key: "sk-server-secret", GroupID: int64Ptr(1), Status: core.StatusAPIKeyActive,
		Group: &core.Group{ID: 1, Platform: core.PlatformOpenAI, Status: core.StatusActive, AllowImageGeneration: true, Hydrated: true},
		User:  &core.User{ID: 42, Status: core.StatusActive},
	}
	keys := &fakeKeys{listed: []core.APIKey{{ID: key.ID, UserID: key.UserID, GroupID: key.GroupID, Status: key.Status}}, hydrated: key}
	svc := newTestService(keys)

	_, err := svc.Prepare(context.Background(), 42, GenerationRequest{GroupID: 1, Model: "gpt-5.4-mini", Prompt: "draw"})
	require.ErrorIs(t, err, ErrModelNotAvailable)

	_, err = svc.Prepare(context.Background(), 42, GenerationRequest{GroupID: 1, Model: "gpt-image-2", Prompt: "draw", OutputFormat: "jpeg", OutputCompression: intPtr(101)})
	require.ErrorIs(t, err, ErrInvalidParameter)

	prepared, err := svc.Prepare(context.Background(), 42, GenerationRequest{GroupID: 1, Model: "gpt-image-2", Prompt: "draw", Size: "2048x2048"})
	require.NoError(t, err)
	require.Equal(t, "2048x2048", prepared.Request.Size)

	_, err = svc.Prepare(context.Background(), 42, GenerationRequest{GroupID: 1, Model: "gpt-image-2", Prompt: "draw", Size: "1000x1000"})
	require.ErrorIs(t, err, ErrInvalidParameter)

	_, err = svc.Prepare(context.Background(), 42, GenerationRequest{GroupID: 99, Model: "gpt-image-2", Prompt: "draw"})
	require.ErrorIs(t, err, ErrGroupNotAllowed)

	_, err = svc.Prepare(context.Background(), 42, GenerationRequest{GroupID: 1, Model: "gpt-image-2", Prompt: strings.Repeat("x", maxPromptRunes+1)})
	require.ErrorIs(t, err, ErrPromptTooLong)
}

func TestImageGenerationServicePrepareSelectsHydratedServerKey(t *testing.T) {
	key := &core.APIKey{
		ID: 9, UserID: 42, Key: "sk-server-secret", GroupID: int64Ptr(1), Status: core.StatusAPIKeyActive,
		Group: &core.Group{ID: 1, Platform: core.PlatformOpenAI, Status: core.StatusActive, AllowImageGeneration: true, Hydrated: true},
		User:  &core.User{ID: 42, Status: core.StatusActive},
	}
	keys := &fakeKeys{listed: []core.APIKey{{ID: key.ID, UserID: key.UserID, GroupID: key.GroupID, Status: key.Status}}, hydrated: key}
	svc := newTestService(keys)
	prepared, err := svc.Prepare(context.Background(), 42, GenerationRequest{GroupID: 1, Model: "GPT-IMAGE-2", Prompt: " draw "})
	require.NoError(t, err)
	require.Same(t, key, prepared.APIKey)
	require.Equal(t, "gpt-image-2", prepared.Request.Model)
	require.Equal(t, "draw", prepared.Request.Prompt)
	require.Equal(t, 1, prepared.Request.N)
	require.Equal(t, "png", prepared.Request.OutputFormat)
}

func TestImageGenerationServiceRejectsKeyThatChangesStateAfterList(t *testing.T) {
	key := &core.APIKey{
		ID: 9, UserID: 42, Key: "sk-server-secret", GroupID: int64Ptr(1), Status: core.StatusAPIKeyDisabled,
		Group: &core.Group{ID: 1, Platform: core.PlatformOpenAI, Status: core.StatusActive, AllowImageGeneration: true, Hydrated: true},
		User:  &core.User{ID: 42, Status: core.StatusActive},
	}
	keys := &fakeKeys{listed: []core.APIKey{{ID: key.ID, UserID: key.UserID, GroupID: key.GroupID, Status: core.StatusAPIKeyActive}}, hydrated: key}
	svc := newTestService(keys)
	_, err := svc.Prepare(context.Background(), 42, GenerationRequest{GroupID: 1, Model: "gpt-image-2", Prompt: "draw"})
	require.ErrorIs(t, err, ErrImageAPIKeyMissing)
}

func int64Ptr(v int64) *int64 { return &v }
func intPtr(v int) *int       { return &v }
