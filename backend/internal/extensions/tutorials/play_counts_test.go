package tutorials

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakePlayCountStore struct {
	counts map[string]int64
}

func (s *fakePlayCountStore) List(_ context.Context, videoIDs []string) (map[string]int64, error) {
	result := make(map[string]int64, len(videoIDs))
	for _, videoID := range videoIDs {
		result[videoID] = s.counts[videoID]
	}
	return result, nil
}

func (s *fakePlayCountStore) Increment(_ context.Context, videoID string) (int64, error) {
	s.counts[videoID]++
	return s.counts[videoID], nil
}

func TestPublicHandlerListsCountsAndIncrementsOnlyKnownVideos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	videos := []Video{
		{ID: "intro", Title: "Intro", VideoURL: "https://example.com/intro", Enabled: true},
		{ID: "intro/video", Title: "Intro video", VideoURL: "https://example.com/intro-video", Enabled: true},
	}
	store := &fakePlayCountStore{counts: map[string]int64{"intro": 4, "intro/video": 2}}
	h := NewVideoHandler(func(context.Context) ([]Video, error) { return videos, nil }, store)
	router := gin.New()
	router.GET("/tutorials/videos", h.List)
	router.POST("/tutorials/videos/*id", h.Play)

	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/tutorials/videos", nil))
	require.Equal(t, http.StatusOK, listRecorder.Code)
	var listPayload struct {
		Data struct {
			Videos []PublicVideo `json:"videos"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(listRecorder.Body.Bytes(), &listPayload))
	require.Len(t, listPayload.Data.Videos, 2)
	require.Equal(t, int64(4), listPayload.Data.Videos[0].PlayCount)

	playRecorder := httptest.NewRecorder()
	router.ServeHTTP(playRecorder, httptest.NewRequest(http.MethodPost, "/tutorials/videos/intro/play", nil))
	require.Equal(t, http.StatusOK, playRecorder.Code)
	require.Equal(t, int64(5), store.counts["intro"])

	slashRecorder := httptest.NewRecorder()
	router.ServeHTTP(slashRecorder, httptest.NewRequest(http.MethodPost, "/tutorials/videos/intro%2Fvideo/play", nil))
	require.Equal(t, http.StatusOK, slashRecorder.Code)
	require.Equal(t, int64(3), store.counts["intro/video"])

	unknownRecorder := httptest.NewRecorder()
	router.ServeHTTP(unknownRecorder, httptest.NewRequest(http.MethodPost, "/tutorials/videos/unknown/play", nil))
	require.Equal(t, http.StatusNotFound, unknownRecorder.Code)
	require.Equal(t, int64(5), store.counts["intro"])
}

func TestSQLPlayCountStoreIncrementUsesAtomicUpsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := NewSQLPlayCountStore(db)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO tutorial_video_play_counts")).
		WithArgs("intro").
		WillReturnRows(sqlmock.NewRows([]string{"play_count"}).AddRow(int64(7)))

	count, err := store.Increment(context.Background(), "intro")
	require.NoError(t, err)
	require.Equal(t, int64(7), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewSQLPlayCountStoreAcceptsNilForConfigurationChecks(t *testing.T) {
	store := NewSQLPlayCountStore((*sql.DB)(nil))
	_, err := store.Increment(context.Background(), "intro")
	require.Error(t, err)
}
