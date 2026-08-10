package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type tutorialSettingsRepoStub struct {
	values map[string]string
}

func (s *tutorialSettingsRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *tutorialSettingsRepoStub) GetValue(context.Context, string) (string, error) {
	panic("unexpected GetValue call")
}

func (s *tutorialSettingsRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (s *tutorialSettingsRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

func (s *tutorialSettingsRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *tutorialSettingsRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *tutorialSettingsRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func TestGetFrameSrcOriginsIncludesTutorialDocumentation(t *testing.T) {
	service := NewSettingService(&tutorialSettingsRepoStub{values: map[string]string{}}, &config.Config{})

	origins, err := service.GetFrameSrcOrigins(context.Background())

	require.NoError(t, err)
	require.Contains(t, origins, "https://doc.aiprox.net")
}
