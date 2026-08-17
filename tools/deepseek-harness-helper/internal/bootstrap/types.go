package bootstrap

import "encoding/json"

const (
	StatusCheckingEnvironment = "checking_environment"
	StatusInstalling          = "installing"
	StatusConfiguring         = "configuring"
	StatusStarting            = "starting"
	StatusCompleted           = "completed"
	StatusFailed              = "failed"
)

type LaunchRequest struct {
	Server      string
	Ticket      string
	OperationID string
	ExtensionID string
}

type ExchangeRequest struct {
	Ticket string `json:"ticket"`
}

type Envelope[T any] struct {
	Data T `json:"data"`
}

type Model struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ContextWindow int    `json:"context_window"`
	MaxTokens     int    `json:"max_tokens"`
}

type Provider struct {
	Route          string `json:"route"`
	DisplayName    string `json:"display_name"`
	Protocol       string `json:"protocol"`
	BaseURL        string `json:"base_url"`
	CredentialName string `json:"credential_name"`
	Model          Model  `json:"model"`
}

type Task struct {
	ServerOrigin         string          `json:"-"`
	ExtensionID          string          `json:"-"`
	OperationID          string          `json:"operation_id"`
	EventToken           string          `json:"event_token"`
	StatusURL            string          `json:"status_url"`
	ProtocolVersion      string          `json:"protocol_version"`
	ToolID               string          `json:"tool_id"`
	ToolVersion          string          `json:"tool_version"`
	MinimumHelperVersion string          `json:"minimum_helper_version"`
	Payload              json.RawMessage `json:"payload,omitempty"`
	// Deprecated DSH compatibility fields. New adapters decode Payload instead.
	DSHVersion string   `json:"dsh_version,omitempty"`
	APIKey     string   `json:"api_key"`
	Provider   Provider `json:"provider"`
}

type DeepSeekHarnessPayload struct {
	APIKey   string   `json:"api_key"`
	Provider Provider `json:"provider"`
}

const (
	CurrentTaskProtocolVersion = "1"
	DeepSeekHarnessToolID      = "deepseek-harness"
	DevelopmentHelperVersion   = "0.1.4"
)

type StatusEvent struct {
	Status     string `json:"status"`
	Stage      string `json:"stage"`
	Message    string `json:"message"`
	Progress   int    `json:"progress"`
	HarnessURL string `json:"harness_url,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
}
