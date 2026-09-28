package pelican

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPelicanSupportedIntervals(t *testing.T) {
	now := time.Date(2026, 9, 28, 22, 23, 0, 0, time.FixedZone("CST", 8*3600))
	for _, tc := range []struct{ minutes, hour, minute int }{{60, 23, 0}, {30, 22, 30}, {10, 22, 30}} {
		in := SaveConfig{Revision: 1, TopicMode: "rotate", FixedTopicID: "pelican-ski", MaxOutputTokens: 16384, TimeoutSeconds: 300, RetentionDays: 30, IntervalMinutes: tc.minutes}
		if err := in.Validate(); err != nil {
			t.Fatal(err)
		}
		next := nextScheduledRun(now, tc.minutes).In(now.Location())
		if next.Hour() != tc.hour || next.Minute() != tc.minute || !next.After(now) {
			t.Fatalf("interval %d: %v", tc.minutes, next)
		}
		if got := nextScheduledRun(next, tc.minutes); !got.Equal(next.Add(time.Duration(tc.minutes) * time.Minute)) {
			t.Fatalf("boundary repeated for %d", tc.minutes)
		}
		in.IntervalMinutes = 5
		if in.Validate() == nil {
			t.Fatal("unsupported frequency accepted")
		}
	}
}

func TestPelicanManualTriggerStopsWithoutClaim(t *testing.T) {
	runner := NewRunner(&fakeRunStore{}, nil, testEncryptor{}, true)
	runner.Stop()
	if id, err := runner.Trigger(context.Background()); id != 0 || !errors.Is(err, ErrRunnerUnavailable) {
		t.Fatalf("stopped runner accepted: %d %v", id, err)
	}
}

func TestPelicanManualTriggerUsesExistingWorker(t *testing.T) {
	store := &fakeRunStore{}
	done := make(chan struct{})
	runner := NewRunner(store, generatorFunc(func(context.Context, string, string, int) (*Generation, error) {
		close(done)
		return &Generation{Text: safeExample}, nil
	}), testEncryptor{}, true)
	runner.Start()
	t.Cleanup(runner.Stop)
	id, err := runner.Trigger(context.Background())
	if err != nil || id != 7 {
		t.Fatalf("trigger: %d %v", id, err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker was not dispatched")
	}
	runner.Stop()
	if store.finishes != 1 || store.status != "succeeded" || store.preview == "" {
		t.Fatalf("manual work did not use preview/finish: %+v", store)
	}
}
