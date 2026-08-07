//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func newGroupPlazaService(groups []Group, pricing *PricingService) *ChannelService {
	return NewChannelService(&mockChannelRepository{
		listAllFn: func(context.Context) ([]Channel, error) {
			return nil, errors.New("channels must not be queried")
		},
	}, &stubGroupRepoForAvailable{activeGroups: groups}, nil, pricing)
}

func TestListPlazaGroups_UsesEnabledGroupModelsWithoutChannels(t *testing.T) {
	groups := []Group{
		{ID: 10, Name: "public", Description: "desc", Platform: "openai", RateMultiplier: 1,
			ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{" gpt-5 ", "gpt-5", "", "gpt-image-2"}}},
		{ID: 20, Name: "disabled", Platform: "anthropic", ModelsListConfig: GroupModelsListConfig{Models: []string{"claude-sonnet"}}},
		{ID: 30, Name: "empty", Platform: "openai", ModelsListConfig: GroupModelsListConfig{Enabled: true}},
	}
	out, err := newGroupPlazaService(groups, nil).ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(10), out[0].ID)
	require.Equal(t, "desc", out[0].Description)
	require.Equal(t, "openai", out[0].Platform)
	require.Equal(t, []string{"gpt-5", "gpt-image-2"}, []string{out[0].Models[0].Name, out[0].Models[1].Name})
	for _, model := range out[0].Models {
		require.Equal(t, "openai", model.Platform)
	}
}

func TestListPlazaGroups_DedupesTrimmedModelsAndSortsGroups(t *testing.T) {
	groups := []Group{
		{ID: 1, Name: "standard", Platform: "openai", RateMultiplier: 1,
			ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{" z-model ", "a-model", "z-model"}}},
		{ID: 2, Name: "cheap", Platform: "anthropic", RateMultiplier: 0.5,
			ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{"claude-sonnet"}}},
	}
	out, err := newGroupPlazaService(groups, nil).ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, "cheap", out[0].Name)
	require.Equal(t, []string{"a-model", "z-model"}, []string{out[1].Models[0].Name, out[1].Models[1].Name})
}

func TestListPlazaGroups_SynthesizesOfficialPricingAndKeepsUnknown(t *testing.T) {
	pricing := newStubPricingServiceFromMap(map[string]*LiteLLMModelPricing{
		"claude-sonnet": {
			Mode: "chat", InputCostPerToken: 3e-6, OutputCostPerToken: 1.5e-5,
			CacheCreationInputTokenCostAbove1hr: 6e-6,
		},
	})
	groups := []Group{{ID: 10, Name: "g", Platform: "anthropic", RateMultiplier: 1,
		ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{"claude-sonnet", "unknown-model"}}}}
	out, err := newGroupPlazaService(groups, pricing).ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Len(t, out[0].Models, 2)
	known := out[0].Models[0]
	require.Equal(t, "claude-sonnet", known.Name)
	require.NotNil(t, known.Pricing)
	require.InDelta(t, 3e-6, *known.Pricing.InputPrice, 1e-12)
	require.NotNil(t, known.OfficialPricing)
	require.InDelta(t, 6e-6, *known.OfficialPricing.CacheWrite1hPrice, 1e-12)
	unknown := out[0].Models[1]
	require.Equal(t, "unknown-model", unknown.Name)
	require.Nil(t, unknown.Pricing)
	require.Nil(t, unknown.OfficialPricing)
}

func TestListPlazaGroups_GroupImagePriceOverridesOfficialPricing(t *testing.T) {
	groupPrice := 0.02
	pricing := newStubPricingServiceFromMap(map[string]*LiteLLMModelPricing{
		"gpt-image-2": {Mode: "image_generation", OutputCostPerImage: 0.04},
	})
	groups := []Group{{ID: 10, Name: "images", Platform: "openai", RateMultiplier: 1,
		ImagePrice1K:     &groupPrice,
		ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{"gpt-image-2"}}}}
	out, err := newGroupPlazaService(groups, pricing).ListPlazaGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
	p := out[0].Models[0].Pricing
	require.NotNil(t, p)
	require.Equal(t, BillingModeImage, p.BillingMode)
	require.Len(t, p.Intervals, 1)
	require.InDelta(t, groupPrice, *p.Intervals[0].PerRequestPrice, 1e-12)
}

func TestListPlazaGroups_GroupRepoErrorPropagates(t *testing.T) {
	sentinel := errors.New("boom")
	svc := NewChannelService(nil, &stubGroupRepoForAvailable{listActiveErr: sentinel}, nil, nil)
	out, err := svc.ListPlazaGroups(context.Background())
	require.Nil(t, out)
	require.ErrorIs(t, err, sentinel)
}
