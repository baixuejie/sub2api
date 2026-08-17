package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/consent"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/dsh"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/runner"
)

type Workflow struct {
	Client        *Client
	Runner        runner.Runner
	Paths         config.Paths
	Output        io.Writer
	WarningOutput io.Writer
	ConfirmTrust  consent.Prompt
}

func (w *Workflow) Run(ctx context.Context, rawURI string) (string, error) {
	launch, err := ParseLaunchURI(rawURI)
	if err != nil {
		return "", err
	}
	if w.Paths.TrustedSitesFile == "" {
		return "", errors.New("trusted server file path is required")
	}
	if err := consent.EnsureTrusted(ctx, w.Paths.TrustedSitesFile, launch.Server, w.ConfirmTrust); err != nil {
		return "", err
	}
	client := w.Client
	if client == nil {
		client = NewClient()
	}
	task, err := client.Exchange(ctx, launch)
	if err != nil {
		return "", err
	}
	if task.OperationID != launch.OperationID {
		return "", errors.New("exchange operation_id mismatch")
	}
	var reportWarnings []error
	report := func(event StatusEvent) error {
		err := client.Report(ctx, task, event)
		if IsReportOutcomeUnknown(err) {
			reportWarnings = append(reportWarnings, err)
			return nil
		}
		return err
	}
	fail := func(code string, cause error) (string, error) {
		event := StatusEvent{Status: StatusFailed, Stage: StatusFailed, Message: publicFailure(cause), Progress: 100, ErrorCode: code}
		if reportErr := client.Report(ctx, task, event); reportErr != nil {
			return "", fmt.Errorf("%w; additionally failed to report status: %v", cause, reportErr)
		}
		return "", cause
	}
	if err := report(StatusEvent{Status: StatusCheckingEnvironment, Stage: StatusCheckingEnvironment, Message: "Checking Node.js and npm", Progress: 10}); err != nil {
		return "", err
	}
	run := w.Runner
	if run == nil {
		run = runner.ExecRunner{}
	}
	environment, err := dsh.CheckEnvironment(ctx, run)
	if err != nil {
		return fail("environment_check_failed", err)
	}
	if err := report(StatusEvent{Status: StatusInstalling, Stage: StatusInstalling, Message: "Installing the pinned DeepSeek Harness version", Progress: 30}); err != nil {
		return "", err
	}
	var dshBin, harnessURL string
	reportFailed := false
	localErr := config.WithBootstrapLock(w.Paths, func() error {
		dshBin, err = dsh.Install(ctx, run, environment, w.Paths, task.DSHVersion)
		if err != nil {
			return &workflowStageError{code: "dsh_install_failed", cause: err}
		}
		if err := report(StatusEvent{Status: StatusConfiguring, Stage: StatusConfiguring, Message: "Writing provider and credential configuration", Progress: 60}); err != nil {
			reportFailed = true
			return err
		}
		provider := config.ProviderConfig{
			Route: task.Provider.Route, DisplayName: task.Provider.DisplayName, Protocol: task.Provider.Protocol,
			BaseURL: task.Provider.BaseURL, CredentialName: task.Provider.CredentialName,
			ModelID: task.Provider.Model.ID, ModelName: task.Provider.Model.Name,
			ContextWindow: task.Provider.Model.ContextWindow, MaxTokens: task.Provider.Model.MaxTokens,
		}
		if err := report(StatusEvent{Status: StatusStarting, Stage: StatusStarting, Message: "Starting DeepSeek Harness", Progress: 80}); err != nil {
			reportFailed = true
			return err
		}
		if err := dsh.StopManaged(ctx, environment, w.Paths, dshBin, task.DSHVersion); err != nil {
			return &workflowStageError{code: "dsh_stop_failed", cause: err}
		}
		if err := config.Apply(w.Paths, provider, task.APIKey); err != nil {
			return &workflowStageError{code: "configuration_failed", cause: err}
		}
		started, err := dsh.StartOrReuse(ctx, environment, w.Paths, dshBin, task.DSHVersion, task.OperationID)
		if err != nil {
			return &workflowStageError{code: "dsh_start_failed", cause: err}
		}
		harnessURL = started.URL
		if err := client.Report(ctx, task, StatusEvent{Status: StatusCompleted, Stage: StatusCompleted, Message: "DeepSeek Harness is ready", Progress: 100, HarnessURL: harnessURL}); err != nil {
			reportWarnings = append(reportWarnings, err)
		}
		return nil
	})
	if localErr != nil {
		if reportFailed {
			return "", localErr
		}
		var stageErr *workflowStageError
		if errors.As(localErr, &stageErr) {
			return fail(stageErr.code, stageErr.cause)
		}
		return fail("bootstrap_lock_failed", localErr)
	}
	if w.WarningOutput != nil {
		for _, warning := range reportWarnings {
			_, _ = fmt.Fprintf(w.WarningOutput, "warning: status synchronization did not complete: %v\n", warning)
		}
	}
	if w.Output != nil {
		_, _ = fmt.Fprintln(w.Output, harnessURL)
	}
	return harnessURL, nil
}

type workflowStageError struct {
	code  string
	cause error
}

func (e *workflowStageError) Error() string { return e.cause.Error() }
func (e *workflowStageError) Unwrap() error { return e.cause }

func publicFailure(err error) string {
	message := strings.TrimSpace(err.Error())
	if len(message) > 500 {
		message = message[:500]
	}
	return message
}
