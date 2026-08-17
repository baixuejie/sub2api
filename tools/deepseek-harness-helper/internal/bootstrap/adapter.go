package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/runner"
	"golang.org/x/mod/semver"
)

type AdapterExecution struct {
	Task   Task
	Runner runner.Runner
	Paths  config.Paths
	Report func(StatusEvent) error
}

type AdapterResult struct {
	OpenURL           string
	CompletionMessage string
}

type HelperUpgradeRequiredError struct {
	Current string
	Minimum string
}

func (e *HelperUpgradeRequiredError) Error() string {
	return fmt.Sprintf("Helper version %s does not satisfy minimum_helper_version %s", e.Current, e.Minimum)
}

// Adapter is an in-process, explicitly registered tool implementation. Tasks
// select an adapter by ID; they never supply executables, arguments, or scripts.
type Adapter interface {
	ToolID() string
	Validate(Task) error
	Execute(context.Context, AdapterExecution) (AdapterResult, error)
}

type AdapterRegistry struct {
	adapters map[string]Adapter
}

func NewAdapterRegistry(adapters ...Adapter) (*AdapterRegistry, error) {
	registry := &AdapterRegistry{adapters: make(map[string]Adapter, len(adapters))}
	for _, adapter := range adapters {
		if adapter == nil {
			return nil, errors.New("register nil tool adapter")
		}
		id := adapter.ToolID()
		if id == "" || strings.TrimSpace(id) != id {
			return nil, errors.New("register tool adapter with invalid ID")
		}
		if _, exists := registry.adapters[id]; exists {
			return nil, fmt.Errorf("register duplicate tool adapter %q", id)
		}
		registry.adapters[id] = adapter
	}
	return registry, nil
}

func DefaultAdapterRegistry() *AdapterRegistry {
	registry, err := NewAdapterRegistry(DeepSeekHarnessAdapter{})
	if err != nil {
		panic(err)
	}
	return registry
}

func (r *AdapterRegistry) Resolve(task Task, helperVersion string) (Adapter, Task, error) {
	normalized, legacy, err := normalizeTaskContract(task)
	if err != nil {
		return nil, Task{}, err
	}
	if normalized.ProtocolVersion != CurrentTaskProtocolVersion {
		return nil, Task{}, fmt.Errorf("unsupported task protocol_version %q", normalized.ProtocolVersion)
	}
	if !legacy {
		if err := requireHelperVersion(helperVersion, normalized.MinimumHelperVersion); err != nil {
			return nil, Task{}, err
		}
	}
	if r == nil {
		return nil, Task{}, errors.New("tool adapter registry is required")
	}
	adapter, ok := r.adapters[normalized.ToolID]
	if !ok {
		return nil, Task{}, fmt.Errorf("unsupported tool_id %q", normalized.ToolID)
	}
	if err := adapter.Validate(normalized); err != nil {
		return nil, Task{}, err
	}
	return adapter, normalized, nil
}

func normalizeTaskContract(task Task) (Task, bool, error) {
	genericFields := []string{task.ProtocolVersion, task.ToolID, task.ToolVersion, task.MinimumHelperVersion}
	present := 0
	for _, field := range genericFields {
		if field != "" {
			present++
		}
	}
	if present == 0 {
		if task.DSHVersion == "" {
			return Task{}, false, errors.New("exchange returned a task without a tool contract")
		}
		task.ProtocolVersion = CurrentTaskProtocolVersion
		task.ToolID = DeepSeekHarnessToolID
		task.ToolVersion = task.DSHVersion
		return task, true, nil
	}
	if present != len(genericFields) {
		return Task{}, false, errors.New("exchange returned an incomplete tool contract")
	}
	if task.DSHVersion != "" && task.DSHVersion != task.ToolVersion {
		return Task{}, false, errors.New("exchange returned conflicting tool_version and dsh_version")
	}
	if len(task.ProtocolVersion) > 32 || len(task.ToolID) > 128 || len(task.ToolVersion) > 128 || len(task.MinimumHelperVersion) > 128 {
		return Task{}, false, errors.New("exchange returned an oversized tool contract")
	}
	return task, false, nil
}

func requireHelperVersion(current, minimum string) error {
	currentVersion, err := canonicalHelperVersion(current, true)
	if err != nil {
		return fmt.Errorf("invalid current Helper version: %w", err)
	}
	minimumVersion, err := canonicalHelperVersion(minimum, false)
	if err != nil {
		return fmt.Errorf("invalid minimum_helper_version: %w", err)
	}
	if semver.Compare(currentVersion, minimumVersion) < 0 {
		return &HelperUpgradeRequiredError{
			Current: strings.TrimPrefix(currentVersion, "v"),
			Minimum: strings.TrimPrefix(minimumVersion, "v"),
		}
	}
	return nil
}

func canonicalHelperVersion(value string, allowDevelopment bool) (string, error) {
	value = strings.TrimSpace(value)
	if allowDevelopment && (value == "" || value == "dev") {
		value = DevelopmentHelperVersion
	}
	if !strings.HasPrefix(value, "v") {
		value = "v" + value
	}
	if !semver.IsValid(value) {
		return "", fmt.Errorf("%q is not valid semantic version", strings.TrimPrefix(value, "v"))
	}
	return semver.Canonical(value), nil
}
