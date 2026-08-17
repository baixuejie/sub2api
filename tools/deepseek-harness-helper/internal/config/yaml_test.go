package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMergeSettingsPreservesOtherNamespaces(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "settings.yaml")
	original := `locale:
  preference: zh
llm-pi-ai:
  providers:
    existing:
      displayName: Existing
    sub2api:
      displayName: Legacy managed
    sub2api-private:
      displayName: User managed
`
	if err := os.WriteFile(file, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	provider := ProviderConfig{Route: "sub2api-openai", DisplayName: "Sub2API", Protocol: "openai-responses", BaseURL: "https://example.com/v1", CredentialName: "SUB2API_API_KEY", ModelID: "model", ModelName: "Model", ContextWindow: 128000, MaxTokens: 8192}
	if err := MergeSettings(file, provider); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	backup, err := os.ReadFile(file + ".bak")
	if err != nil || string(backup) != original {
		t.Fatalf("settings backup mismatch: %v", err)
	}
	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if _, ok := parsed["locale"]; !ok {
		t.Fatal("locale namespace was lost")
	}
	llm := parsed["llm-pi-ai"].(map[string]any)
	providers := llm["providers"].(map[string]any)
	if _, ok := providers["existing"]; !ok {
		t.Fatal("existing provider was lost")
	}
	if _, ok := providers["sub2api"]; ok {
		t.Fatal("legacy Sub2API-managed provider was retained")
	}
	if _, ok := providers["sub2api-private"]; !ok {
		t.Fatal("user provider sharing the prefix was removed")
	}
	added := providers["sub2api-openai"].(map[string]any)
	if added["apiKeyEnv"] != "SUB2API_API_KEY" || added["baseURL"] != "https://example.com/v1" {
		t.Fatalf("unexpected provider: %#v", added)
	}
	selection := parsed["agent-default-model"].(map[string]any)
	if selection["provider"] != "sub2api-openai" || selection["model"] != "model" {
		t.Fatalf("unexpected selection: %#v", selection)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(file)
		if err != nil || info.Mode().Perm()&0o077 != 0 {
			t.Fatalf("settings permissions are not private: %v, %v", info.Mode(), err)
		}
	}
}

func TestMergeCredential(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), ".credentials.yaml")
	if err := MergeCredential(file, "SUB2API_API_KEY", "secret-value"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "SUB2API_API_KEY: secret-value") {
		t.Fatalf("unexpected credential file: %s", data)
	}
	if err := MergeCredential(file, "bad-name", "secret"); err == nil {
		t.Fatal("expected invalid credential name")
	}
	if err := MergeCredential(file, "GOOD_NAME", ""); err == nil {
		t.Fatal("expected empty credential rejection")
	}
}

func TestMergeSettingsRejectsReservedCredentialConflict(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "settings.yaml")
	original := "llm-pi-ai:\n  providers:\n    custom-provider:\n      apiKeyEnv: SUB2API_API_KEY\n"
	if err := os.WriteFile(file, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	provider := ProviderConfig{Route: "sub2api-openai", DisplayName: "Sub2API", Protocol: "openai-responses", BaseURL: "https://example.com/v1", CredentialName: "SUB2API_API_KEY", ModelID: "model", ModelName: "Model", ContextWindow: 128000, MaxTokens: 8192}
	if err := MergeSettings(file, provider); err == nil {
		t.Fatal("expected reserved credential conflict")
	}
}

func TestMergeSettingsRejectsNonMappingTargetNamespaceWithoutChangingFile(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "settings.yaml")
	original := "llm-pi-ai:\n  - unexpected\n"
	if err := os.WriteFile(file, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	provider := ProviderConfig{Route: "sub2api-openai", DisplayName: "Sub2API", Protocol: "openai-responses", BaseURL: "https://example.com/v1", CredentialName: "SUB2API_API_KEY", ModelID: "model", ModelName: "Model", ContextWindow: 128000, MaxTokens: 8192}
	if err := MergeSettings(file, provider); err == nil {
		t.Fatal("expected target namespace rejection")
	}
	data, err := os.ReadFile(file)
	if err != nil || string(data) != original {
		t.Fatalf("target namespace changed after rejection: %v", err)
	}
}

func TestMergeSettingsRejectsMalformedRoot(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "settings.yaml")
	if err := os.WriteFile(file, []byte("- not-a-map\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provider := ProviderConfig{Route: "r", DisplayName: "R", Protocol: "openai-completions", BaseURL: "https://example.com", CredentialName: "KEY", ModelID: "m", ModelName: "M", ContextWindow: 1, MaxTokens: 1}
	if err := MergeSettings(file, provider); err == nil {
		t.Fatal("expected malformed root rejection")
	}
}
