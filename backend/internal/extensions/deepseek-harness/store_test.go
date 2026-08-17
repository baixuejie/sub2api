package deepseekharness

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestDeepSeekHarnessRedisStoreConsumesTicketOnce(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store := newRedisInstallStore(client)
	now := time.Now().UTC()
	session := &InstallSession{
		ID:        "session-1",
		UserID:    7,
		APIKeyID:  42,
		Status:    statusAwaitingHelper,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}
	ticket := &ticketRecord{SessionID: session.ID, UserID: 7, APIKeyID: 42, EventToken: "event-token"}

	err := store.Create(context.Background(), session, "ticket-digest", ticket, time.Minute, time.Hour)
	require.NoError(t, err)
	require.True(t, server.Exists(sessionKeyPrefix+session.ID))
	require.True(t, server.Exists(ticketKeyPrefix+"ticket-digest"))

	consumed, err := store.ConsumeTicket(context.Background(), "ticket-digest")
	require.NoError(t, err)
	require.Equal(t, session.ID, consumed.SessionID)
	_, err = store.ConsumeTicket(context.Background(), "ticket-digest")
	require.ErrorIs(t, err, errTicketNotFound)
}

func TestDeepSeekHarnessRedisStoreCASPreventsStaleTerminalOverwrite(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store := newRedisInstallStore(client)
	now := time.Now().UTC()
	session := &InstallSession{
		ID:        "session-cas",
		UserID:    7,
		APIKeyID:  42,
		Status:    statusInstalling,
		Progress:  30,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}
	ticket := &ticketRecord{SessionID: session.ID, UserID: 7, APIKeyID: 42, EventToken: "event-token"}
	require.NoError(t, store.Create(context.Background(), session, "ticket-cas", ticket, time.Minute, time.Hour))

	staleRead := make(chan struct{})
	allowStaleWrite := make(chan struct{})
	var firstAttempt atomic.Bool
	staleResult := make(chan error, 1)
	go func() {
		_, err := store.UpdateSession(context.Background(), session.ID, func(current *InstallSession) error {
			if firstAttempt.CompareAndSwap(false, true) {
				close(staleRead)
				<-allowStaleWrite
			}
			if current.Status == statusCompleted {
				return errInvalidEvent
			}
			current.Status = statusConfiguring
			current.Progress = 60
			return nil
		})
		staleResult <- err
	}()

	<-staleRead
	_, err := store.UpdateSession(context.Background(), session.ID, func(current *InstallSession) error {
		current.Status = statusCompleted
		current.Progress = 100
		current.HarnessURL = "http://127.0.0.1:3080"
		return nil
	})
	require.NoError(t, err)
	close(allowStaleWrite)
	require.ErrorIs(t, <-staleResult, errInvalidEvent)

	stored, err := store.GetSession(context.Background(), session.ID)
	require.NoError(t, err)
	require.Equal(t, statusCompleted, stored.Status)
	require.Equal(t, 100, stored.Progress)
}

func TestDeepSeekHarnessRedisStoreAllowsOnlyOneConcurrentConsumer(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store := newRedisInstallStore(client)
	now := time.Now().UTC()
	session := &InstallSession{ID: "session-2", UserID: 7, APIKeyID: 42, ExpiresAt: now.Add(time.Hour)}
	ticket := &ticketRecord{SessionID: session.ID, UserID: 7, APIKeyID: 42, EventToken: "event-token"}
	require.NoError(t, store.Create(context.Background(), session, "ticket-digest", ticket, time.Minute, time.Hour))

	var successes atomic.Int32
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if _, err := store.ConsumeTicket(context.Background(), "ticket-digest"); err == nil {
				successes.Add(1)
			}
		}()
	}
	wait.Wait()
	require.EqualValues(t, 1, successes.Load())
}
