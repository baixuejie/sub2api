package tutorials

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeVideosJSON(t *testing.T) {
	normalized, err := NormalizeVideosJSON(`[
		{"id":"second","title":" Second ","cover_url":"https://cdn.example.com/cover.png","video_url":"https://video.example.com/second","enabled":true,"sort_order":20},
		{"id":"first","title":"First","video_url":"https://video.example.com/first","enabled":false,"sort_order":10}
	]`)
	require.NoError(t, err)

	var videos []Video
	require.NoError(t, json.Unmarshal([]byte(normalized), &videos))
	require.Len(t, videos, 2)
	require.Equal(t, "first", videos[0].ID)
	require.Equal(t, "second", videos[1].ID)
	require.Equal(t, "Second", videos[1].Title)
}

func TestNormalizeVideosJSONRejectsInvalidItems(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "malformed JSON", raw: `[{`},
		{name: "missing video URL", raw: `[{
			"id":"one","title":"One","enabled":true,"sort_order":0
		}]`},
		{name: "non HTTP video URL", raw: `[{
			"id":"one","title":"One","video_url":"javascript:alert(1)","enabled":true,"sort_order":0
		}]`},
		{name: "duplicate ID", raw: `[
			{"id":"one","title":"One","video_url":"https://example.com/1","enabled":true,"sort_order":0},
			{"id":"one","title":"Two","video_url":"https://example.com/2","enabled":true,"sort_order":1}
		]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeVideosJSON(tt.raw)
			require.Error(t, err)
		})
	}
}

func TestNormalizeVideosJSONCountsUnicodeCharacters(t *testing.T) {
	longTitle := strings.Repeat("教", VideoMaxTitleLen)
	normalized, err := NormalizeVideosJSON(`[{"id":"one","title":"` + longTitle + `","video_url":"https://example.com/video","enabled":true,"sort_order":0}]`)
	require.NoError(t, err)
	require.Contains(t, normalized, longTitle)
}

func TestParsePublicVideosFiltersDisabledAndMalformedValues(t *testing.T) {
	videos := ParsePublicVideos(`[
		{"id":"hidden","title":"Hidden","video_url":"https://example.com/hidden","enabled":false,"sort_order":0},
		{"id":"shown","title":"Shown","cover_url":"https://example.com/cover.png","video_url":"https://example.com/shown","enabled":true,"sort_order":1}
	]`)
	require.Len(t, videos, 1)
	require.Equal(t, "shown", videos[0].ID)
	require.Equal(t, "https://example.com/shown", videos[0].VideoURL)

	require.Empty(t, ParsePublicVideos(`[{"id":"bad"}]`))
	require.Empty(t, ParsePublicVideos(strings.Repeat("x", VideoMaxURLLen+1)))
}
