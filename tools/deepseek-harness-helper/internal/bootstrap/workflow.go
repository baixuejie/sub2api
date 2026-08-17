package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/consent"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/runner"
)

type Workflow struct {
	Client        *Client
	Runner        runner.Runner
	Paths         config.Paths
	Output        io.Writer
	WarningOutput io.Writer
	ConfirmTrust  consent.Prompt
	Registry      *AdapterRegistry
	HelperVersion string
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
	registry := w.Registry
	if registry == nil {
		registry = DefaultAdapterRegistry()
	}
	adapter, normalizedTask, err := registry.Resolve(task, w.HelperVersion)
	if err != nil {
		code := "unsupported_tool_contract"
		var upgradeRequired *HelperUpgradeRequiredError
		if errors.As(err, &upgradeRequired) {
			code = "helper_update_required"
		}
		event := StatusEvent{
			Status: StatusFailed, Stage: StatusFailed, Message: publicFailure(err), Progress: 100, ErrorCode: code,
		}
		if reportErr := client.Report(ctx, task, event); reportErr != nil {
			return "", fmt.Errorf("%w; additionally failed to report status: %v", err, reportErr)
		}
		return "", err
	}
	task = normalizedTask
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
	run := w.Runner
	if run == nil {
		run = runner.ExecRunner{}
	}
	result, executionErr := adapter.Execute(ctx, AdapterExecution{Task: task, Runner: run, Paths: w.Paths, Report: report})
	if executionErr != nil {
		var reportErr *workflowReportError
		if errors.As(executionErr, &reportErr) {
			return "", reportErr.cause
		}
		var stageErr *workflowStageError
		if errors.As(executionErr, &stageErr) {
			return fail(stageErr.code, stageErr.cause)
		}
		return fail("tool_execution_failed", executionErr)
	}
	if err := client.Report(ctx, task, StatusEvent{Status: StatusCompleted, Stage: StatusCompleted, Message: result.CompletionMessage, Progress: 100, HarnessURL: result.OpenURL}); err != nil {
		reportWarnings = append(reportWarnings, err)
	}
	if w.WarningOutput != nil {
		for _, warning := range reportWarnings {
			_, _ = fmt.Fprintf(w.WarningOutput, "warning: status synchronization did not complete: %v\n", warning)
		}
	}
	if w.Output != nil && result.OpenURL != "" {
		_, _ = fmt.Fprintln(w.Output, result.OpenURL)
	}
	return result.OpenURL, nil
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
