package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration224CreatesTutorialVideoPlayCounts(t *testing.T) {
	content, err := FS.ReadFile("224_tutorial_video_play_counts.sql")
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS tutorial_video_play_counts")
	require.Contains(t, sql, "video_id TEXT PRIMARY KEY")
	require.Contains(t, sql, "play_count BIGINT NOT NULL DEFAULT 0")
	require.Contains(t, sql, "CHECK (play_count >= 0)")
	require.Contains(t, sql, "updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()")
}
