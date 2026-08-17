package deepseekharness

import "time"

const (
	defaultTicketTTL      = 2 * time.Minute
	defaultSessionTTL     = time.Hour
	sessionRetentionGrace = 5 * time.Minute
	pinnedDSHVersion      = "0.1.0-rc.6"
	taskProtocolVersion   = "1"
	deepSeekHarnessToolID = "deepseek-harness"
	minimumHelperVersion  = "0.1.6"

	statusAwaitingHelper    = "awaiting_helper"
	statusCheckingEnv       = "checking_environment"
	statusInstalling        = "installing"
	statusConfiguring       = "configuring"
	statusStarting          = "starting"
	statusCompleted         = "completed"
	statusFailed            = "failed"
	statusExpired           = "expired"
	credentialReferenceName = "SUB2API_API_KEY"
)

type ModelOption struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ContextWindow int    `json:"context_window"`
	MaxTokens     int    `json:"max_tokens"`
}

type InstallProfile struct {
	APIKeyID        int64         `json:"api_key_id"`
	APIKeyName      string        `json:"api_key_name"`
	KeyHint         string        `json:"key_hint"`
	GroupName       string        `json:"group_name"`
	Platform        string        `json:"platform"`
	Provider        string        `json:"provider"`
	ProviderName    string        `json:"provider_name"`
	Protocol        string        `json:"protocol"`
	BaseURL         string        `json:"base_url"`
	DefaultModel    string        `json:"default_model"`
	SelectedModel   string        `json:"selected_model"`
	AvailableModels []ModelOption `json:"available_models"`
	ServerURL       string        `json:"-"`
}

type HelperDownloads struct {
	WindowsAMD64 string `json:"windows_amd64"`
	WindowsARM64 string `json:"windows_arm64"`
	DarwinAMD64  string `json:"darwin_amd64"`
	DarwinARM64  string `json:"darwin_arm64"`
	LinuxAMD64   string `json:"linux_amd64"`
	LinuxARM64   string `json:"linux_arm64"`
	ReleasesPage string `json:"releases_page"`
}

type ProfileResponse struct {
	Profile              InstallProfile  `json:"profile"`
	HelperDownloads      HelperDownloads `json:"helper_downloads"`
	RequiredNode         string          `json:"required_node"`
	ProtocolVersion      string          `json:"protocol_version"`
	ToolID               string          `json:"tool_id"`
	ToolVersion          string          `json:"tool_version"`
	MinimumHelperVersion string          `json:"minimum_helper_version"`
	DSHVersion           string          `json:"dsh_version"`
}

type InstallSession struct {
	ID               string         `json:"id"`
	UserID           int64          `json:"user_id"`
	APIKeyID         int64          `json:"api_key_id"`
	Profile          InstallProfile `json:"profile"`
	ServerURL        string         `json:"server_url"`
	Status           string         `json:"status"`
	Stage            string         `json:"stage"`
	Message          string         `json:"message"`
	Progress         int            `json:"progress"`
	HarnessURL       string         `json:"harness_url,omitempty"`
	ErrorCode        string         `json:"error_code,omitempty"`
	EventTokenDigest string         `json:"event_token_digest"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	ExpiresAt        time.Time      `json:"expires_at"`
}

type SessionResponse struct {
	ID              string         `json:"id"`
	Profile         InstallProfile `json:"profile"`
	Status          string         `json:"status"`
	Stage           string         `json:"stage"`
	Message         string         `json:"message"`
	Progress        int            `json:"progress"`
	HarnessURL      string         `json:"harness_url,omitempty"`
	ErrorCode       string         `json:"error_code,omitempty"`
	LaunchURI       string         `json:"launch_uri,omitempty"`
	TicketExpiresAt *time.Time     `json:"ticket_expires_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	ExpiresAt       time.Time      `json:"expires_at"`
}

func sessionResponse(session *InstallSession) SessionResponse {
	if session == nil {
		return SessionResponse{}
	}
	profile := session.Profile
	profile.ServerURL = ""
	return SessionResponse{
		ID:         session.ID,
		Profile:    profile,
		Status:     session.Status,
		Stage:      session.Stage,
		Message:    session.Message,
		Progress:   session.Progress,
		HarnessURL: session.HarnessURL,
		ErrorCode:  session.ErrorCode,
		CreatedAt:  session.CreatedAt,
		UpdatedAt:  session.UpdatedAt,
		ExpiresAt:  session.ExpiresAt,
	}
}

type ticketRecord struct {
	SessionID  string `json:"session_id"`
	UserID     int64  `json:"user_id"`
	APIKeyID   int64  `json:"api_key_id"`
	Model      string `json:"model"`
	EventToken string `json:"event_token"`
}

type CreateSessionRequest struct {
	APIKeyID int64  `json:"api_key_id"`
	Model    string `json:"model,omitempty"`
}

type ExchangeRequest struct {
	Ticket string `json:"ticket"`
}

type InstallEvent struct {
	Status     string `json:"status"`
	Stage      string `json:"stage"`
	Message    string `json:"message"`
	Progress   int    `json:"progress"`
	HarnessURL string `json:"harness_url,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
}

type ProviderTask struct {
	Route          string      `json:"route"`
	DisplayName    string      `json:"display_name"`
	Protocol       string      `json:"protocol"`
	BaseURL        string      `json:"base_url"`
	CredentialName string      `json:"credential_name"`
	Model          ModelOption `json:"model"`
}

type DeepSeekHarnessPayload struct {
	APIKey   string       `json:"api_key"`
	Provider ProviderTask `json:"provider"`
}

type BootstrapTask struct {
	OperationID          string                 `json:"operation_id"`
	EventToken           string                 `json:"event_token"`
	StatusURL            string                 `json:"status_url"`
	ProtocolVersion      string                 `json:"protocol_version"`
	ToolID               string                 `json:"tool_id"`
	ToolVersion          string                 `json:"tool_version"`
	MinimumHelperVersion string                 `json:"minimum_helper_version"`
	Payload              DeepSeekHarnessPayload `json:"payload"`
	// Deprecated flat fields are retained until pre-payload Helpers age out.
	DSHVersion string       `json:"dsh_version"`
	APIKey     string       `json:"api_key"`
	Provider   ProviderTask `json:"provider"`
}
