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
	w.writeProgress(0, "Parsing the Sub2API local-tool request")
	launch, err := ParseLaunchURI(rawURI)
	if err != nil {
		return "", err
	}
	if w.Paths.TrustedSitesFile == "" {
		return "", errors.New("trusted server file path is required")
	}
	w.writeProgress(5, fmt.Sprintf("Confirming access to %s (%s)", launch.Server, launch.ExtensionID))
	if err := consent.EnsureTrusted(ctx, w.Paths.TrustedSitesFile, launch.Server, launch.ExtensionID, w.ConfirmTrust); err != nil {
		return "", err
	}
	client := w.Client
	if client == nil {
		client = NewClient()
	}
	w.writeProgress(8, "Exchanging the one-time installation ticket")
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
		w.writeStatusEvent(event)
		if reportErr := client.Report(ctx, task, event); reportErr != nil {
			return "", fmt.Errorf("%w; additionally failed to report status: %v", err, reportErr)
		}
		return "", err
	}
	task = normalizedTask
	var reportWarnings []error
	report := func(event StatusEvent) error {
		w.writeStatusEvent(event)
		err := client.Report(ctx, task, event)
		if IsReportOutcomeUnknown(err) {
			reportWarnings = append(reportWarnings, err)
			return nil
		}
		return err
	}
	fail := func(code string, cause error) (string, error) {
		event := StatusEvent{Status: StatusFailed, Stage: StatusFailed, Message: publicFailure(cause), Progress: 100, ErrorCode: code}
		w.writeStatusEvent(event)
		if reportErr := client.Report(ctx, task, event); reportErr != nil {
			return "", fmt.Errorf("%w; additionally failed to report status: %v", cause, reportErr)
		}
		return "", cause
	}
	run := w.Runner
	if run == nil {
		run = runner.ExecRunner{}
	}
	toolDataDir, err := w.Paths.ToolDataDir(task.ToolID)
	if err != nil {
		return fail("tool_data_path_failed", err)
	}
	result, executionErr := adapter.Execute(ctx, AdapterExecution{
		Task: task, Runner: run, Paths: w.Paths, ToolDataDir: toolDataDir, Report: report,
		Output: w.Output, ErrorOutput: w.WarningOutput,
	})
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
	completedEvent := StatusEvent{Status: StatusCompleted, Stage: StatusCompleted, Message: result.CompletionMessage, Progress: 100, HarnessURL: result.OpenURL}
	w.writeStatusEvent(completedEvent)
	if err := client.Report(ctx, task, completedEvent); err != nil {
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

func (w *Workflow) writeStatusEvent(event StatusEvent) {
	w.writeProgress(event.Progress, event.Message)
}

func (w *Workflow) writeProgress(progress int, message string) {
	if w.Output == nil {
		return
	}
	if progress < 0 {
		progress = 0
	} else if progress > 100 {
		progress = 100
	}
	message = strings.TrimSpace(message)
	if message == "" {
		message = "Working"
	}
	_, _ = fmt.Fprintf(w.Output, "[%3d%%] %s\n", progress, message)
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
