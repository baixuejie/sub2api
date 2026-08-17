package deepseekharness

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	coreservice "github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	defaultHelperReleaseBaseURL = "https://github.com/baixuejie/sub2api/releases/download/dsh-helper-v0.1.0"
	defaultHelperReleasesPage   = "https://github.com/baixuejie/sub2api/releases"
)

type apiKeyReader interface {
	GetByID(context.Context, int64) (*coreservice.APIKey, error)
}

type settingsReader interface {
	GetPublicSettings(context.Context) (*coreservice.PublicSettings, error)
	IsDeepSeekHarnessEnabled(context.Context) bool
}

type installService struct {
	apiKeys    apiKeyReader
	settings   settingsReader
	store      installStore
	now        func() time.Time
	ticketTTL  time.Duration
	sessionTTL time.Duration
}

func newInstallService(apiKeys apiKeyReader, settings settingsReader, store installStore) *installService {
	return &installService{
		apiKeys:    apiKeys,
		settings:   settings,
		store:      store,
		now:        time.Now,
		ticketTTL:  defaultTicketTTL,
		sessionTTL: defaultSessionTTL,
	}
}

func (s *installService) Profile(ctx context.Context, userID, apiKeyID int64, fallbackOrigin string) (ProfileResponse, error) {
	if err := s.ensureEnabled(ctx); err != nil {
		return ProfileResponse{}, err
	}
	profile, _, err := s.resolveProfile(ctx, userID, apiKeyID, fallbackOrigin, "")
	if err != nil {
		return ProfileResponse{}, err
	}
	return ProfileResponse{
		Profile:         profile,
		HelperDownloads: helperDownloads(),
		RequiredNode:    ">=22.19.0",
		DSHVersion:      pinnedDSHVersion,
	}, nil
}

func (s *installService) CreateSession(
	ctx context.Context,
	userID int64,
	request CreateSessionRequest,
	fallbackOrigin string,
) (SessionResponse, error) {
	if err := s.ensureEnabled(ctx); err != nil {
		return SessionResponse{}, err
	}
	profile, _, err := s.resolveProfile(ctx, userID, request.APIKeyID, fallbackOrigin, request.Model)
	if err != nil {
		return SessionResponse{}, err
	}
	sessionID, err := randomToken(18)
	if err != nil {
		return SessionResponse{}, fmt.Errorf("generate deepseek harness session id: %w", err)
	}
	ticketToken, err := randomToken(32)
	if err != nil {
		return SessionResponse{}, fmt.Errorf("generate deepseek harness ticket: %w", err)
	}
	eventToken, err := randomToken(32)
	if err != nil {
		return SessionResponse{}, fmt.Errorf("generate deepseek harness event token: %w", err)
	}

	now := s.now().UTC()
	session := &InstallSession{
		ID:               sessionID,
		UserID:           userID,
		APIKeyID:         request.APIKeyID,
		Profile:          profile,
		ServerURL:        profile.ServerURL,
		Status:           statusAwaitingHelper,
		Stage:            "waiting_for_helper",
		Message:          "Waiting for the local helper",
		Progress:         0,
		EventTokenDigest: digestToken(eventToken),
		CreatedAt:        now,
		UpdatedAt:        now,
		ExpiresAt:        now.Add(s.sessionTTL),
	}
	ticket := &ticketRecord{
		SessionID:  sessionID,
		UserID:     userID,
		APIKeyID:   request.APIKeyID,
		Model:      profile.SelectedModel,
		EventToken: eventToken,
	}
	if err = s.store.Create(ctx, session, digestToken(ticketToken), ticket, s.ticketTTL, s.sessionTTL+sessionRetentionGrace); err != nil {
		return SessionResponse{}, err
	}

	launchURL := &url.URL{Scheme: "sub2api-harness", Host: "bootstrap"}
	query := launchURL.Query()
	query.Set("server", profile.ServerURL)
	query.Set("ticket", ticketToken)
	query.Set("operation_id", sessionID)
	launchURL.RawQuery = query.Encode()
	ticketExpiresAt := now.Add(s.ticketTTL)
	response := sessionResponse(session)
	response.LaunchURI = launchURL.String()
	response.TicketExpiresAt = &ticketExpiresAt
	return response, nil
}

func (s *installService) GetSession(ctx context.Context, userID int64, sessionID string) (SessionResponse, error) {
	if err := s.ensureEnabled(ctx); err != nil {
		return SessionResponse{}, err
	}
	sessionID = strings.TrimSpace(sessionID)
	if !validOpaqueID(sessionID, 128) {
		return SessionResponse{}, errSessionNotFound
	}
	session, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return SessionResponse{}, err
	}
	if session.UserID != userID {
		return SessionResponse{}, errSessionNotFound
	}
	if !s.now().Before(session.ExpiresAt) && session.Status != statusCompleted && session.Status != statusFailed {
		session, err = s.store.UpdateSession(ctx, sessionID, func(current *InstallSession) error {
			if current.UserID != userID {
				return errSessionNotFound
			}
			if current.Status == statusCompleted || current.Status == statusFailed || current.Status == statusExpired {
				return nil
			}
			current.Status = statusExpired
			current.Stage = statusExpired
			current.Message = "Installation session expired"
			current.UpdatedAt = s.now().UTC()
			return nil
		})
		if err != nil {
			return SessionResponse{}, err
		}
	}
	return sessionResponse(session), nil
}

func (s *installService) Exchange(ctx context.Context, request ExchangeRequest) (BootstrapTask, error) {
	if err := s.ensureEnabled(ctx); err != nil {
		return BootstrapTask{}, err
	}
	ticketToken := strings.TrimSpace(request.Ticket)
	if !validOpaqueID(ticketToken, 256) {
		return BootstrapTask{}, errTicketNotFound
	}
	ticketDigest := digestToken(ticketToken)
	ticket, err := s.store.GetTicket(ctx, ticketDigest)
	if err != nil {
		return BootstrapTask{}, err
	}
	session, err := s.store.GetSession(ctx, ticket.SessionID)
	if err != nil {
		return BootstrapTask{}, err
	}
	if session.UserID != ticket.UserID || session.APIKeyID != ticket.APIKeyID ||
		!secureDigestEqual(session.EventTokenDigest, digestToken(ticket.EventToken)) {
		return BootstrapTask{}, errInvalidSession
	}
	if !s.now().Before(session.ExpiresAt) {
		return BootstrapTask{}, errInvalidSession
	}
	profile, key, err := s.resolveProfile(ctx, ticket.UserID, ticket.APIKeyID, session.ServerURL, ticket.Model)
	if err != nil {
		s.failSession(ctx, session.ID, "key_revalidation_failed", "API key validation failed")
		return BootstrapTask{}, err
	}
	model, err := selectedModelOption(profile)
	if err != nil {
		s.failSession(ctx, session.ID, "model_revalidation_failed", "Model validation failed")
		return BootstrapTask{}, err
	}
	statusURL := strings.TrimRight(profile.ServerURL, "/") + "/api/v1/deepseek-harness/sessions/" + session.ID + "/events"
	task := BootstrapTask{
		OperationID: session.ID,
		EventToken:  ticket.EventToken,
		StatusURL:   statusURL,
		DSHVersion:  pinnedDSHVersion,
		APIKey:      key.Key,
		Provider: ProviderTask{
			Route:          profile.Provider,
			DisplayName:    profile.ProviderName,
			Protocol:       profile.Protocol,
			BaseURL:        profile.BaseURL,
			CredentialName: credentialReferenceName,
			Model:          model,
		},
	}

	_, err = s.store.UpdateSession(ctx, session.ID, func(current *InstallSession) error {
		if current.UserID != ticket.UserID || current.APIKeyID != ticket.APIKeyID ||
			!secureDigestEqual(current.EventTokenDigest, digestToken(ticket.EventToken)) ||
			!s.now().Before(current.ExpiresAt) {
			return errInvalidSession
		}
		if current.Status != statusAwaitingHelper && current.Status != statusCheckingEnv {
			return errInvalidSession
		}
		current.Profile = profile
		current.ServerURL = profile.ServerURL
		current.Status = statusCheckingEnv
		current.Stage = "checking_environment"
		current.Message = "Checking the local environment"
		current.Progress = 5
		current.UpdatedAt = s.now().UTC()
		return nil
	})
	if err != nil {
		return BootstrapTask{}, err
	}
	consumed, err := s.store.ConsumeTicket(ctx, ticketDigest)
	if err != nil {
		return BootstrapTask{}, err
	}
	if consumed.SessionID != ticket.SessionID || !secureDigestEqual(digestToken(consumed.EventToken), digestToken(ticket.EventToken)) {
		return BootstrapTask{}, errInvalidSession
	}
	return task, nil
}

func (s *installService) UpdateSession(
	ctx context.Context,
	sessionID, eventToken string,
	event InstallEvent,
) (SessionResponse, error) {
	if err := s.ensureEnabled(ctx); err != nil {
		return SessionResponse{}, err
	}
	sessionID = strings.TrimSpace(sessionID)
	eventToken = strings.TrimSpace(eventToken)
	if !validOpaqueID(sessionID, 128) || !validOpaqueID(eventToken, 256) {
		return SessionResponse{}, errInvalidEventToken
	}
	session, err := s.store.UpdateSession(ctx, sessionID, func(current *InstallSession) error {
		if !secureDigestEqual(current.EventTokenDigest, digestToken(eventToken)) {
			return errInvalidEventToken
		}
		if !s.now().Before(current.ExpiresAt) {
			return errInvalidSession
		}
		if err := validateEvent(current, &event); err != nil {
			return err
		}
		current.Status = event.Status
		current.Stage = strings.TrimSpace(event.Stage)
		current.Message = strings.TrimSpace(event.Message)
		current.Progress = event.Progress
		current.HarnessURL = strings.TrimSpace(event.HarnessURL)
		current.ErrorCode = strings.TrimSpace(event.ErrorCode)
		current.UpdatedAt = s.now().UTC()
		return nil
	})
	if err != nil {
		return SessionResponse{}, err
	}
	return sessionResponse(session), nil
}

func (s *installService) ensureEnabled(ctx context.Context) error {
	if s == nil || s.settings == nil || !s.settings.IsDeepSeekHarnessEnabled(ctx) {
		return errFeatureDisabled
	}
	return nil
}

func (s *installService) resolveProfile(
	ctx context.Context,
	userID, apiKeyID int64,
	fallbackOrigin, selectedModel string,
) (InstallProfile, *coreservice.APIKey, error) {
	if s == nil || s.apiKeys == nil || s.settings == nil || s.store == nil || userID <= 0 || apiKeyID <= 0 {
		return InstallProfile{}, nil, errAPIKeyNotFound
	}
	key, err := s.apiKeys.GetByID(ctx, apiKeyID)
	if err != nil || key == nil || key.UserID != userID {
		return InstallProfile{}, nil, errAPIKeyNotFound
	}
	if !key.IsActive() || key.IsExpired() || key.IsQuotaExhausted() || strings.TrimSpace(key.Key) == "" {
		return InstallProfile{}, nil, errAPIKeyUnavailable
	}
	if key.Group == nil || !key.Group.IsActive() {
		return InstallProfile{}, nil, errUnsupportedGroup
	}
	settings, err := s.settings.GetPublicSettings(ctx)
	if err != nil {
		return InstallProfile{}, nil, fmt.Errorf("load public settings: %w", err)
	}
	if settings == nil {
		return InstallProfile{}, nil, errors.New("public settings are unavailable")
	}
	profile, err := buildInstallProfile(key, settings.APIBaseURL, fallbackOrigin, selectedModel)
	if err != nil {
		return InstallProfile{}, nil, err
	}
	return profile, key, nil
}

func (s *installService) failSession(ctx context.Context, sessionID, code, message string) {
	if sessionID == "" {
		return
	}
	_, _ = s.store.UpdateSession(ctx, sessionID, func(current *InstallSession) error {
		if current.Status == statusCompleted || current.Status == statusFailed || current.Status == statusExpired {
			return nil
		}
		current.Status = statusFailed
		current.Stage = statusFailed
		current.ErrorCode = code
		current.Message = message
		current.UpdatedAt = s.now().UTC()
		return nil
	})
}

func validateEvent(session *InstallSession, event *InstallEvent) error {
	if session == nil || event == nil {
		return errInvalidEvent
	}
	event.Status = strings.TrimSpace(event.Status)
	event.Stage = strings.TrimSpace(event.Stage)
	event.Message = strings.TrimSpace(event.Message)
	event.ErrorCode = strings.TrimSpace(event.ErrorCode)
	if len(event.Stage) > 80 || len(event.Message) > 500 || len(event.ErrorCode) > 80 {
		return errInvalidEvent
	}
	if event.Progress < 0 || event.Progress > 100 || event.Progress < session.Progress {
		return errInvalidEvent
	}
	if session.Status == statusCompleted || session.Status == statusFailed || session.Status == statusExpired {
		if session.Status == event.Status &&
			strings.TrimSpace(session.Stage) == strings.TrimSpace(event.Stage) &&
			strings.TrimSpace(session.Message) == strings.TrimSpace(event.Message) &&
			session.Progress == event.Progress &&
			strings.TrimSpace(session.HarnessURL) == strings.TrimSpace(event.HarnessURL) &&
			strings.TrimSpace(session.ErrorCode) == strings.TrimSpace(event.ErrorCode) {
			return nil
		}
		return errInvalidEvent
	}
	if statusRank(event.Status) < statusRank(session.Status) || statusRank(event.Status) < 0 {
		return errInvalidEvent
	}
	if event.Status == statusCompleted {
		if event.Progress != 100 || !validHarnessURL(event.HarnessURL) {
			return errInvalidEvent
		}
	}
	if event.Status != statusCompleted && strings.TrimSpace(event.HarnessURL) != "" {
		return errInvalidEvent
	}
	return nil
}

func statusRank(status string) int {
	switch status {
	case statusAwaitingHelper:
		return 0
	case statusCheckingEnv:
		return 1
	case statusInstalling:
		return 2
	case statusConfiguring:
		return 3
	case statusStarting:
		return 4
	case statusCompleted, statusFailed:
		return 5
	default:
		return -1
	}
}

func validHarnessURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "http" || parsed.User != nil || parsed.Port() == "" {
		return false
	}
	host := strings.Trim(strings.ToLower(parsed.Hostname()), "[]")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func randomToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func digestToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func secureDigestEqual(left, right string) bool {
	leftBytes, leftErr := hex.DecodeString(left)
	rightBytes, rightErr := hex.DecodeString(right)
	if leftErr != nil || rightErr != nil || len(leftBytes) != len(rightBytes) {
		return false
	}
	return subtle.ConstantTimeCompare(leftBytes, rightBytes) == 1
}

func validOpaqueID(value string, maxLength int) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxLength {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_' {
			continue
		}
		return false
	}
	return true
}

func helperReleaseBaseURL() string {
	if value := strings.TrimSpace(os.Getenv("DEEPSEEK_HARNESS_HELPER_RELEASE_BASE_URL")); value != "" {
		return value
	}
	return defaultHelperReleaseBaseURL
}

func helperReleasesPageURL() string {
	if value := strings.TrimSpace(os.Getenv("DEEPSEEK_HARNESS_HELPER_RELEASES_PAGE")); value != "" {
		return value
	}
	return defaultHelperReleasesPage
}

func isPublicError(err error) bool {
	return errors.Is(err, errFeatureDisabled) || errors.Is(err, errAPIKeyNotFound) || errors.Is(err, errAPIKeyUnavailable) ||
		errors.Is(err, errUnsupportedGroup) || errors.Is(err, errInvalidModel) ||
		errors.Is(err, errInvalidBaseURL) || errors.Is(err, errTicketNotFound) ||
		errors.Is(err, errSessionNotFound) || errors.Is(err, errInvalidSession) ||
		errors.Is(err, errInvalidEvent) || errors.Is(err, errInvalidEventToken)
}
