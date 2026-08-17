package deepseekharness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	ticketKeyPrefix  = "ext:deepseek-harness:ticket:v1:"
	sessionKeyPrefix = "ext:deepseek-harness:session:v1:"
)

var (
	errTicketNotFound  = errors.New("deepseek harness ticket not found")
	errSessionNotFound = errors.New("deepseek harness session not found")
)

type installStore interface {
	Create(context.Context, *InstallSession, string, *ticketRecord, time.Duration, time.Duration) error
	GetTicket(context.Context, string) (*ticketRecord, error)
	ConsumeTicket(context.Context, string) (*ticketRecord, error)
	GetSession(context.Context, string) (*InstallSession, error)
	UpdateSession(context.Context, string, func(*InstallSession) error) (*InstallSession, error)
}

type redisInstallStore struct {
	client *redis.Client
}

func newRedisInstallStore(client *redis.Client) installStore {
	return &redisInstallStore{client: client}
}

func (s *redisInstallStore) Create(
	ctx context.Context,
	session *InstallSession,
	ticketDigest string,
	ticket *ticketRecord,
	ticketTTL time.Duration,
	sessionTTL time.Duration,
) error {
	if s == nil || s.client == nil || session == nil || ticket == nil || ticketDigest == "" {
		return errors.New("invalid deepseek harness store request")
	}
	sessionPayload, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("encode deepseek harness session: %w", err)
	}
	ticketPayload, err := json.Marshal(ticket)
	if err != nil {
		return fmt.Errorf("encode deepseek harness ticket: %w", err)
	}
	_, err = s.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, sessionKeyPrefix+session.ID, sessionPayload, sessionTTL)
		pipe.Set(ctx, ticketKeyPrefix+ticketDigest, ticketPayload, ticketTTL)
		return nil
	})
	if err != nil {
		return fmt.Errorf("store deepseek harness session: %w", err)
	}
	return nil
}

func (s *redisInstallStore) GetTicket(ctx context.Context, ticketDigest string) (*ticketRecord, error) {
	if s == nil || s.client == nil || ticketDigest == "" {
		return nil, errTicketNotFound
	}
	payload, err := s.client.Get(ctx, ticketKeyPrefix+ticketDigest).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, errTicketNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get deepseek harness ticket: %w", err)
	}
	return decodeTicket(payload)
}

func (s *redisInstallStore) ConsumeTicket(ctx context.Context, ticketDigest string) (*ticketRecord, error) {
	if s == nil || s.client == nil || ticketDigest == "" {
		return nil, errTicketNotFound
	}
	payload, err := s.client.GetDel(ctx, ticketKeyPrefix+ticketDigest).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, errTicketNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("consume deepseek harness ticket: %w", err)
	}
	return decodeTicket(payload)
}

func decodeTicket(payload []byte) (*ticketRecord, error) {
	var ticket ticketRecord
	if err := json.Unmarshal(payload, &ticket); err != nil || ticket.SessionID == "" {
		return nil, errTicketNotFound
	}
	return &ticket, nil
}

func (s *redisInstallStore) GetSession(ctx context.Context, sessionID string) (*InstallSession, error) {
	if s == nil || s.client == nil || sessionID == "" {
		return nil, errSessionNotFound
	}
	payload, err := s.client.Get(ctx, sessionKeyPrefix+sessionID).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, errSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get deepseek harness session: %w", err)
	}
	var session InstallSession
	if err = json.Unmarshal(payload, &session); err != nil || session.ID == "" {
		return nil, errSessionNotFound
	}
	return &session, nil
}

func (s *redisInstallStore) UpdateSession(
	ctx context.Context,
	sessionID string,
	update func(*InstallSession) error,
) (*InstallSession, error) {
	if s == nil || s.client == nil || sessionID == "" || update == nil {
		return nil, errSessionNotFound
	}
	key := sessionKeyPrefix + sessionID
	const maxAttempts = 8
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var updated *InstallSession
		err := s.client.Watch(ctx, func(tx *redis.Tx) error {
			payload, err := tx.Get(ctx, key).Bytes()
			if errors.Is(err, redis.Nil) {
				return errSessionNotFound
			}
			if err != nil {
				return err
			}
			var session InstallSession
			if err = json.Unmarshal(payload, &session); err != nil || session.ID == "" {
				return errSessionNotFound
			}
			if err = update(&session); err != nil {
				return err
			}
			payload, err = json.Marshal(&session)
			if err != nil {
				return fmt.Errorf("encode deepseek harness session: %w", err)
			}
			if _, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, key, payload, redis.KeepTTL)
				return nil
			}); err != nil {
				return err
			}
			updated = &session
			return nil
		}, key)
		if errors.Is(err, redis.TxFailedErr) {
			continue
		}
		if err != nil {
			if errors.Is(err, errSessionNotFound) {
				return nil, errSessionNotFound
			}
			return nil, fmt.Errorf("update deepseek harness session: %w", err)
		}
		return updated, nil
	}
	return nil, errors.New("deepseek harness session update conflicted too many times")
}
