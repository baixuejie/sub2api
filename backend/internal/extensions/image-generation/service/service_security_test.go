package imagegeneration

import (
	"context"
	"testing"
	"time"

	core "github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// TestPrepareRejectsHydratedKeyThatChangedAfterTheInitialList verifies that
// the second (hydrated) read cannot re-enable a key that was disabled, expired,
// or quota-exhausted between the list and the generation request.
func TestPrepareRejectsHydratedKeyThatChangedAfterTheInitialList(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*core.APIKey)
	}{
		{
			name: "disabled",
			mutate: func(key *core.APIKey) {
				key.Status = core.StatusAPIKeyDisabled
			},
		},
		{
			name: "expired",
			mutate: func(key *core.APIKey) {
				expired := time.Now().Add(-time.Minute)
				key.ExpiresAt = &expired
			},
		},
		{
			name: "quota exhausted",
			mutate: func(key *core.APIKey) {
				key.Quota = 1
				key.QuotaUsed = 1
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := &core.APIKey{
				ID: 9, UserID: 42, Key: "server-only-secret", GroupID: int64Ptr(1), Status: core.StatusAPIKeyActive,
				Group: &core.Group{ID: 1, Platform: core.PlatformOpenAI, Status: core.StatusActive, AllowImageGeneration: true, Hydrated: true},
				User:  &core.User{ID: 42, Status: core.StatusActive},
			}
			listed := &core.APIKey{ID: key.ID, UserID: key.UserID, GroupID: key.GroupID, Status: core.StatusAPIKeyActive}
			// Simulate the state change after List returned its lightweight row.
			tc.mutate(key)
			svc := newTestService(&fakeKeys{listed: []core.APIKey{*listed}, hydrated: key})

			_, err := svc.Prepare(context.Background(), 42, GenerationRequest{
				GroupID: 1,
				Model:   "gpt-image-2",
				Prompt:  "draw a safe test image",
			})
			require.ErrorIs(t, err, ErrImageAPIKeyMissing)
		})
	}
}
