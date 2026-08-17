package deepseekharness

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"sync"
	"testing"
	"time"

	coreservice "github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type fakeAPIKeyReader struct {
	key *coreservice.APIKey
	err error
}

func (f *fakeAPIKeyReader) GetByID(context.Context, int64) (*coreservice.APIKey, error) {
	return f.key, f.err
}

type fakeSettingsReader struct {
	baseURL string
	enabled bool
	err     error
}

func (f *fakeSettingsReader) GetPublicSettings(context.Context) (*coreservice.PublicSettings, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &coreservice.PublicSettings{APIBaseURL: f.baseURL, DeepSeekHarnessEnabled: true}, nil
}

func (f *fakeSettingsReader) IsDeepSeekHarnessEnabled(context.Context) bool {
	return f.enabled
}

type memoryInstallStore struct {
	mu       sync.Mutex
	sessions map[string]*InstallSession
	tickets  map[string]*ticketRecord
}

func newMemoryInstallStore() *memoryInstallStore {
	return &memoryInstallStore{
		sessions: make(map[string]*InstallSession),
		tickets:  make(map[string]*ticketRecord),
	}
}

func (s *memoryInstallStore) Create(
	_ context.Context,
	session *InstallSession,
	digest string,
	ticket *ticketRecord,
	_ time.Duration,
	_ time.Duration,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = cloneJSON(session)
	s.tickets[digest] = cloneJSON(ticket)
	return nil
}

func (s *memoryInstallStore) GetTicket(_ context.Context, digest string) (*ticketRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket, ok := s.tickets[digest]
	if !ok {
		return nil, errTicketNotFound
	}
	return cloneJSON(ticket), nil
}

func (s *memoryInstallStore) ConsumeTicket(_ context.Context, digest string) (*ticketRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket, ok := s.tickets[digest]
	if !ok {
		return nil, errTicketNotFound
	}
	delete(s.tickets, digest)
	return cloneJSON(ticket), nil
}

func (s *memoryInstallStore) GetSession(_ context.Context, id string) (*InstallSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, errSessionNotFound
	}
	return cloneJSON(session), nil
}

func (s *memoryInstallStore) UpdateSession(
	_ context.Context,
	id string,
	update func(*InstallSession) error,
) (*InstallSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, errSessionNotFound
	}
	candidate := cloneJSON(session)
	if err := update(candidate); err != nil {
		return nil, err
	}
	s.sessions[id] = cloneJSON(candidate)
	return cloneJSON(candidate), nil
}

func cloneJSON[T any](value *T) *T {
	payload, _ := json.Marshal(value)
	var clone T
	_ = json.Unmarshal(payload, &clone)
	return &clone
}

func activeOpenAIKey() *coreservice.APIKey {
	return &coreservice.APIKey{
		ID:     42,
		UserID: 7,
		Key:    "sk-test-secret-value",
		Name:   "Codex key",
		Status: coreservice.StatusAPIKeyActive,
		Group: &coreservice.Group{
			ID:       3,
			Name:     "OpenAI",
			Platform: "openai",
			Status:   coreservice.StatusActive,
		},
	}
}

func newTestInstallService(key *coreservice.APIKey, store installStore) *installService {
	service := newInstallService(
		&fakeAPIKeyReader{key: key},
		&fakeSettingsReader{baseURL: "https://api.example.com", enabled: true},
		store,
	)
	service.now = func() time.Time { return time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC) }
	return service
}

func TestDeepSeekHarnessFeatureFlagFailsClosed(t *testing.T) {
	service := newInstallService(
		&fakeAPIKeyReader{key: activeOpenAIKey()},
		&fakeSettingsReader{baseURL: "https://api.example.com", enabled: false},
		newMemoryInstallStore(),
	)

	_, err := service.Profile(context.Background(), 7, 42, "")
	require.ErrorIs(t, err, errFeatureDisabled)
	_, err = service.CreateSession(context.Background(), 7, CreateSessionRequest{APIKeyID: 42}, "")
	require.ErrorIs(t, err, errFeatureDisabled)
	_, err = service.Exchange(context.Background(), ExchangeRequest{Ticket: "opaque-ticket"})
	require.ErrorIs(t, err, errFeatureDisabled)
}

func TestDeepSeekHarnessProfileUsesCurrentKeyGroupPolicy(t *testing.T) {
	service := newTestInstallService(activeOpenAIKey(), newMemoryInstallStore())

	result, err := service.Profile(context.Background(), 7, 42, "")
	require.NoError(t, err)
	require.Equal(t, "https://api.example.com/v1", result.Profile.BaseURL)
	require.Equal(t, "openai-responses", result.Profile.Protocol)
	require.Equal(t, "gpt-5.6-sol", result.Profile.DefaultModel)
	require.Equal(t, "gpt-5.6-sol", result.Profile.SelectedModel)
	require.Equal(t, "****alue", result.Profile.KeyHint)
	require.Equal(t, ">=22.19.0", result.RequiredNode)
	require.Equal(t, taskProtocolVersion, result.ProtocolVersion)
	require.Equal(t, deepSeekHarnessToolID, result.ToolID)
	require.Equal(t, pinnedDSHVersion, result.ToolVersion)
	require.Equal(t, minimumHelperVersion, result.MinimumHelperVersion)
	require.Equal(t, pinnedDSHVersion, result.DSHVersion)
}

func TestDeepSeekHarnessCreateAndConsumeTicketOnce(t *testing.T) {
	store := newMemoryInstallStore()
	service := newTestInstallService(activeOpenAIKey(), store)

	created, err := service.CreateSession(context.Background(), 7, CreateSessionRequest{APIKeyID: 42}, "")
	require.NoError(t, err)
	require.Equal(t, statusAwaitingHelper, created.Status)
	require.NotEmpty(t, created.LaunchURI)
	require.NotContains(t, created.LaunchURI, "sk-test-secret-value")

	launchURI, err := url.Parse(created.LaunchURI)
	require.NoError(t, err)
	ticket := launchURI.Query().Get("ticket")
	require.NotEmpty(t, ticket)

	task, err := service.Exchange(context.Background(), ExchangeRequest{Ticket: ticket})
	require.NoError(t, err)
	require.Equal(t, "sk-test-secret-value", task.APIKey)
	require.Equal(t, taskProtocolVersion, task.ProtocolVersion)
	require.Equal(t, deepSeekHarnessToolID, task.ToolID)
	require.Equal(t, pinnedDSHVersion, task.ToolVersion)
	require.Equal(t, minimumHelperVersion, task.MinimumHelperVersion)
	require.Equal(t, pinnedDSHVersion, task.DSHVersion)
	require.Equal(t, "SUB2API_API_KEY", task.Provider.CredentialName)
	require.Equal(t, "gpt-5.6-sol", task.Provider.Model.ID)
	require.Contains(t, task.StatusURL, created.ID)

	_, err = service.Exchange(context.Background(), ExchangeRequest{Ticket: ticket})
	require.ErrorIs(t, err, errTicketNotFound)

	stored, err := store.GetSession(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, statusCheckingEnv, stored.Status)
	require.NotEmpty(t, stored.ServerURL)
	payload, err := json.Marshal(stored)
	require.NoError(t, err)
	require.NotContains(t, string(payload), "sk-test-secret-value")
	require.NotContains(t, string(payload), task.EventToken)
}

func TestDeepSeekHarnessExchangeDoesNotConsumeTicketBeforeValidation(t *testing.T) {
	store := newMemoryInstallStore()
	service := newTestInstallService(activeOpenAIKey(), store)
	created, err := service.CreateSession(context.Background(), 7, CreateSessionRequest{APIKeyID: 42}, "")
	require.NoError(t, err)
	launchURI, err := url.Parse(created.LaunchURI)
	require.NoError(t, err)
	ticketToken := launchURI.Query().Get("ticket")
	ticketDigest := digestToken(ticketToken)

	store.mu.Lock()
	delete(store.sessions, created.ID)
	store.mu.Unlock()
	_, err = service.Exchange(context.Background(), ExchangeRequest{Ticket: ticketToken})
	require.ErrorIs(t, err, errSessionNotFound)

	_, err = store.GetTicket(context.Background(), ticketDigest)
	require.NoError(t, err)
}

func TestDeepSeekHarnessSessionOwnershipAndEvents(t *testing.T) {
	store := newMemoryInstallStore()
	service := newTestInstallService(activeOpenAIKey(), store)
	created, err := service.CreateSession(context.Background(), 7, CreateSessionRequest{APIKeyID: 42}, "")
	require.NoError(t, err)
	launchURI, _ := url.Parse(created.LaunchURI)
	task, err := service.Exchange(context.Background(), ExchangeRequest{Ticket: launchURI.Query().Get("ticket")})
	require.NoError(t, err)

	_, err = service.GetSession(context.Background(), 8, created.ID)
	require.ErrorIs(t, err, errSessionNotFound)

	_, err = service.UpdateSession(context.Background(), created.ID, "wrong-token", InstallEvent{
		Status: statusInstalling, Stage: "installing", Message: "Installing", Progress: 30,
	})
	require.ErrorIs(t, err, errInvalidEventToken)

	updated, err := service.UpdateSession(context.Background(), created.ID, task.EventToken, InstallEvent{
		Status: statusInstalling, Stage: "installing", Message: "Installing DSH", Progress: 30,
	})
	require.NoError(t, err)
	require.Equal(t, 30, updated.Progress)

	_, err = service.UpdateSession(context.Background(), created.ID, task.EventToken, InstallEvent{
		Status: statusCompleted, Stage: "completed", Message: "Ready", Progress: 100,
		HarnessURL: "https://example.com:3080",
	})
	require.ErrorIs(t, err, errInvalidEvent)

	completed, err := service.UpdateSession(context.Background(), created.ID, task.EventToken, InstallEvent{
		Status: statusCompleted, Stage: "completed", Message: "Ready", Progress: 100,
		HarnessURL: "http://127.0.0.1:3080",
	})
	require.NoError(t, err)
	require.Equal(t, statusCompleted, completed.Status)

	duplicate, err := service.UpdateSession(context.Background(), created.ID, task.EventToken, InstallEvent{
		Status: statusCompleted, Stage: "completed", Message: "Ready", Progress: 100,
		HarnessURL: "http://127.0.0.1:3080",
	})
	require.NoError(t, err)
	require.Equal(t, statusCompleted, duplicate.Status)

	_, err = service.UpdateSession(context.Background(), created.ID, task.EventToken, InstallEvent{
		Status: statusCompleted, Stage: "completed", Message: "Changed", Progress: 100,
		HarnessURL: "http://127.0.0.1:3080",
	})
	require.ErrorIs(t, err, errInvalidEvent)
}

func TestDeepSeekHarnessRejectsUnavailableOrForeignKeys(t *testing.T) {
	tests := []struct {
		name   string
		key    *coreservice.APIKey
		userID int64
		want   error
	}{
		{name: "foreign", key: activeOpenAIKey(), userID: 8, want: errAPIKeyNotFound},
		{name: "disabled", key: func() *coreservice.APIKey {
			key := activeOpenAIKey()
			key.Status = coreservice.StatusAPIKeyDisabled
			return key
		}(), userID: 7, want: errAPIKeyUnavailable},
		{name: "missing group", key: func() *coreservice.APIKey {
			key := activeOpenAIKey()
			key.Group = nil
			return key
		}(), userID: 7, want: errUnsupportedGroup},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newTestInstallService(test.key, newMemoryInstallStore())
			_, err := service.Profile(context.Background(), test.userID, 42, "")
			require.ErrorIs(t, err, test.want)
		})
	}
}

func TestDeepSeekHarnessProfileRejectsUnavailableModel(t *testing.T) {
	key := activeOpenAIKey()
	key.Group.ModelsListConfig = coreservice.GroupModelsListConfig{
		Enabled: true,
		Models:  []string{"gpt-5.5"},
	}
	service := newTestInstallService(key, newMemoryInstallStore())

	_, err := service.CreateSession(context.Background(), 7, CreateSessionRequest{
		APIKeyID: 42,
		Model:    "gpt-5.6-sol",
	}, "")
	require.ErrorIs(t, err, errInvalidModel)
}

func TestDeepSeekHarnessSettingsFailureDoesNotExposeInternalDetails(t *testing.T) {
	service := newInstallService(
		&fakeAPIKeyReader{key: activeOpenAIKey()},
		&fakeSettingsReader{enabled: true, err: errors.New("database password leaked")},
		newMemoryInstallStore(),
	)
	_, err := service.Profile(context.Background(), 7, 42, "")
	require.Error(t, err)
	require.NotErrorIs(t, err, errAPIKeyNotFound)
}
