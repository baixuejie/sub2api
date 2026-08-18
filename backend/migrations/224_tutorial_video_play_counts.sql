-- Keep play counts separate from the editable tutorial_videos settings JSON.
CREATE TABLE IF NOT EXISTS tutorial_video_play_counts (
    video_id TEXT PRIMARY KEY,
    play_count BIGINT NOT NULL DEFAULT 0 CHECK (play_count >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE tutorial_video_play_counts IS 'Persistent tutorial video play counters';
COMMENT ON COLUMN tutorial_video_play_counts.video_id IS 'Video ID from tutorial_videos settings';
COMMENT ON COLUMN tutorial_video_play_counts.play_count IS 'Cumulative number of video opens';
