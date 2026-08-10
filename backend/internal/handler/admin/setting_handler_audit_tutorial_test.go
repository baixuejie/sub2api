package admin

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDiffSettingsIncludesTutorialVideos(t *testing.T) {
	before := &service.SystemSettings{TutorialVideos: `[{"id":"intro","title":"Intro","video_url":"https://example.com/old","enabled":true,"sort_order":0}]`}
	after := &service.SystemSettings{TutorialVideos: `[{"id":"intro","title":"Intro","video_url":"https://example.com/new","enabled":true,"sort_order":0}]`}

	changed := diffSettings(before, after, nil, nil, UpdateSettingsRequest{})

	require.Contains(t, changed, "tutorial_videos")
}
