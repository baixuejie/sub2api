package config

import (
	"os"
	"strings"
	"testing"
)

func TestApplyWritesBothDocumentsWithoutPersistentLocks(t *testing.T) {
	t.Parallel()
	paths := PathsFor(t.TempDir())
	if err := EnsurePrivateDir(paths.DSHHome); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.SettingsFile, []byte("locale:\n  preference: zh\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.CredentialsFile, []byte("EXISTING_KEY: existing\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provider := ProviderConfig{
		Route: "sub2api-openai", DisplayName: "Sub2API", Protocol: "openai-responses",
		BaseURL: "https://example.com/v1", CredentialName: "SUB2API_API_KEY", ModelID: "gpt-5.6-sol",
		ModelName: "gpt-5.6-sol", ContextWindow: 128000, MaxTokens: 8192,
	}
	if err := Apply(paths, provider, "secret-value"); err != nil {
		t.Fatal(err)
	}
	settings, err := os.ReadFile(paths.SettingsFile)
	if err != nil || !strings.Contains(string(settings), "sub2api-openai:") || !strings.Contains(string(settings), "preference: zh") {
		t.Fatalf("unexpected settings: %v, %s", err, settings)
	}
	credentials, err := os.ReadFile(paths.CredentialsFile)
	if err != nil || !strings.Contains(string(credentials), "SUB2API_API_KEY: secret-value") || !strings.Contains(string(credentials), "EXISTING_KEY: existing") {
		t.Fatalf("unexpected credentials: %v, %s", err, credentials)
	}
	for _, lockPath := range []string{paths.SettingsFile + ".lock", paths.CredentialsFile + ".lock"} {
		if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
			t.Fatalf("persistent DSH writer lock remains at %s: %v", lockPath, err)
		}
	}
}

func TestApplyRejectsInvalidInputBeforeChangingDocuments(t *testing.T) {
	t.Parallel()
	paths := PathsFor(t.TempDir())
	if err := EnsurePrivateDir(paths.DSHHome); err != nil {
		t.Fatal(err)
	}
	original := "locale:\n  preference: zh\n"
	if err := os.WriteFile(paths.SettingsFile, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	provider := ProviderConfig{
		Route: "sub2api-openai", DisplayName: "Sub2API", Protocol: "openai-responses",
		BaseURL: "https://example.com/v1", CredentialName: "SUB2API_API_KEY", ModelID: "gpt-5.6-sol",
		ModelName: "gpt-5.6-sol", ContextWindow: 128000, MaxTokens: 8192,
	}
	if err := Apply(paths, provider, ""); err == nil {
		t.Fatal("expected empty credential rejection")
	}
	data, err := os.ReadFile(paths.SettingsFile)
	if err != nil || string(data) != original {
		t.Fatalf("settings changed after input rejection: %v", err)
	}
}
