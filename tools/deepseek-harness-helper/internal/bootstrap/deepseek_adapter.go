package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/dsh"
)

type DeepSeekHarnessAdapter struct{}

func (DeepSeekHarnessAdapter) ToolID() string { return DeepSeekHarnessToolID }

func (DeepSeekHarnessAdapter) Validate(task Task) error {
	if task.ToolVersion != dsh.SupportedVersion {
		return fmt.Errorf("unsupported tool_version for %s: expected %s", DeepSeekHarnessToolID, dsh.SupportedVersion)
	}
	p := task.Provider
	if p.Route == "" || p.DisplayName == "" || p.Protocol == "" || p.BaseURL == "" || p.CredentialName == "" || p.Model.ID == "" || p.Model.Name == "" || p.Model.ContextWindow <= 0 || p.Model.MaxTokens <= 0 {
		return errors.New("exchange returned an incomplete provider")
	}
	if p.CredentialName != "SUB2API_API_KEY" || !allowedProviderProtocol(p.Route, p.Protocol) {
		return errors.New("exchange returned an unsupported provider contract")
	}
	if len(p.DisplayName) > 128 || len(p.Model.ID) > 256 || len(p.Model.Name) > 256 || p.Model.ContextWindow > 10_000_000 || p.Model.MaxTokens > 1_000_000 {
		return errors.New("exchange returned invalid provider limits")
	}
	base, err := url.Parse(p.BaseURL)
	if err != nil || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" || base.EscapedPath() != "/v1" {
		return errors.New("provider base_url must end at /v1")
	}
	if base.Scheme != "https" && (base.Scheme != "http" || !isLoopbackHost(strings.TrimSuffix(strings.ToLower(base.Hostname()), "."))) {
		return errors.New("provider base_url must use HTTPS or localhost HTTP")
	}
	return nil
}

func (DeepSeekHarnessAdapter) Execute(ctx context.Context, execution AdapterExecution) (AdapterResult, error) {
	if execution.Report == nil {
		return AdapterResult{}, errors.New("status reporter is required")
	}
	if err := execution.Report(StatusEvent{Status: StatusCheckingEnvironment, Stage: StatusCheckingEnvironment, Message: "Checking Node.js and npm", Progress: 10}); err != nil {
		return AdapterResult{}, &workflowReportError{cause: err}
	}
	environment, err := dsh.CheckEnvironment(ctx, execution.Runner)
	if err != nil {
		return AdapterResult{}, &workflowStageError{code: "environment_check_failed", cause: err}
	}
	if err := execution.Report(StatusEvent{Status: StatusInstalling, Stage: StatusInstalling, Message: "Installing the pinned DeepSeek Harness version", Progress: 30}); err != nil {
		return AdapterResult{}, &workflowReportError{cause: err}
	}

	var dshBin, harnessURL string
	localErr := config.WithBootstrapLock(execution.Paths, func() error {
		dshBin, err = dsh.Install(ctx, execution.Runner, environment, execution.Paths, execution.Task.ToolVersion)
		if err != nil {
			return &workflowStageError{code: "dsh_install_failed", cause: err}
		}
		if err := execution.Report(StatusEvent{Status: StatusConfiguring, Stage: StatusConfiguring, Message: "Writing provider and credential configuration", Progress: 60}); err != nil {
			return &workflowReportError{cause: err}
		}
		provider := config.ProviderConfig{
			Route: execution.Task.Provider.Route, DisplayName: execution.Task.Provider.DisplayName, Protocol: execution.Task.Provider.Protocol,
			BaseURL: execution.Task.Provider.BaseURL, CredentialName: execution.Task.Provider.CredentialName,
			ModelID: execution.Task.Provider.Model.ID, ModelName: execution.Task.Provider.Model.Name,
			ContextWindow: execution.Task.Provider.Model.ContextWindow, MaxTokens: execution.Task.Provider.Model.MaxTokens,
		}
		if err := execution.Report(StatusEvent{Status: StatusStarting, Stage: StatusStarting, Message: "Starting DeepSeek Harness", Progress: 80}); err != nil {
			return &workflowReportError{cause: err}
		}
		if err := dsh.StopManaged(ctx, environment, execution.Paths, dshBin, execution.Task.ToolVersion); err != nil {
			return &workflowStageError{code: "dsh_stop_failed", cause: err}
		}
		if err := config.Apply(execution.Paths, provider, execution.Task.APIKey); err != nil {
			return &workflowStageError{code: "configuration_failed", cause: err}
		}
		started, err := dsh.StartOrReuse(ctx, environment, execution.Paths, dshBin, execution.Task.ToolVersion, execution.Task.OperationID)
		if err != nil {
			return &workflowStageError{code: "dsh_start_failed", cause: err}
		}
		harnessURL = started.URL
		return nil
	})
	if localErr != nil {
		var reportErr *workflowReportError
		var stageErr *workflowStageError
		switch {
		case errors.As(localErr, &reportErr), errors.As(localErr, &stageErr):
			return AdapterResult{}, localErr
		default:
			return AdapterResult{}, &workflowStageError{code: "bootstrap_lock_failed", cause: localErr}
		}
	}
	return AdapterResult{OpenURL: harnessURL, CompletionMessage: "DeepSeek Harness is ready"}, nil
}

func allowedProviderProtocol(route, protocol string) bool {
	allowed := map[string]map[string]struct{}{
		"sub2api-openai":      {"openai-responses": {}},
		"sub2api-anthropic":   {"anthropic-messages": {}},
		"sub2api-grok":        {"openai-responses": {}},
		"sub2api-gemini":      {"openai-completions": {}},
		"sub2api-antigravity": {"anthropic-messages": {}, "openai-completions": {}},
		"sub2api-composite":   {"openai-completions": {}},
	}
	protocols, ok := allowed[route]
	if !ok {
		return false
	}
	_, ok = protocols[protocol]
	return ok
}

type workflowReportError struct {
	cause error
}

func (e *workflowReportError) Error() string { return e.cause.Error() }
func (e *workflowReportError) Unwrap() error { return e.cause }
