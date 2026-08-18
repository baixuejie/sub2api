package tutorials

import (
	"context"
	"database/sql"
	"errors"
)

// SQLPlayCountStore stores counters in the extension-owned table. All writes
// use one PostgreSQL upsert so concurrent clicks cannot lose increments.
type SQLPlayCountStore struct {
	db *sql.DB
}

func NewSQLPlayCountStore(db *sql.DB) *SQLPlayCountStore {
	return &SQLPlayCountStore{db: db}
}

func (s *SQLPlayCountStore) List(ctx context.Context, videoIDs []string) (map[string]int64, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("tutorial play count store is not configured")
	}
	result := make(map[string]int64, len(videoIDs))
	for _, videoID := range videoIDs {
		var count int64
		if err := s.db.QueryRowContext(ctx,
			`SELECT play_count FROM tutorial_video_play_counts WHERE video_id = $1`,
			videoID,
		).Scan(&count); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, err
		}
		result[videoID] = count
	}
	return result, nil
}

func (s *SQLPlayCountStore) Increment(ctx context.Context, videoID string) (int64, error) {
	if s == nil || s.db == nil {
		return 0, errors.New("tutorial play count store is not configured")
	}
	var count int64
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO tutorial_video_play_counts (video_id, play_count, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (video_id) DO UPDATE
		SET play_count = tutorial_video_play_counts.play_count + 1,
		    updated_at = NOW()
		RETURNING play_count
	`, videoID).Scan(&count)
	return count, err
}
