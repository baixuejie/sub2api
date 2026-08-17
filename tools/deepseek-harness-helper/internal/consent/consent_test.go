package consent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEnsureTrustedPromptsOnceAndCanonicalizesOrigin(t *testing.T) {
	file := filepath.Join(t.TempDir(), "trusted-sites.json")
	prompts := 0
	prompt := func(_ context.Context, request TrustRequest) (bool, error) {
		prompts++
		if request.Origin != "https://example.com" || request.ExtensionID != "deepseek-harness" {
			t.Fatalf("request = %#v", request)
		}
		return true, nil
	}
	if err := EnsureTrusted(context.Background(), file, "HTTPS://EXAMPLE.COM:443/", "deepseek-harness", prompt); err != nil {
		t.Fatal(err)
	}
	if err := EnsureTrusted(context.Background(), file, "https://example.com", "deepseek-harness", prompt); err != nil {
		t.Fatal(err)
	}
	if prompts != 1 {
		t.Fatalf("prompts = %d, want 1", prompts)
	}
	trusted, err := IsTrusted(file, "https://example.com/", "deepseek-harness")
	if err != nil || !trusted {
		t.Fatalf("trusted = %v, err = %v", trusted, err)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(file)
		if err != nil || info.Mode().Perm()&0o077 != 0 {
			t.Fatalf("trust file permissions = %v, err = %v", info.Mode(), err)
		}
	}
}

func TestEnsureTrustedScopesApprovalByExtension(t *testing.T) {
	file := filepath.Join(t.TempDir(), "trusted-sites.json")
	prompts := 0
	prompt := func(_ context.Context, _ TrustRequest) (bool, error) {
		prompts++
		return true, nil
	}
	if err := EnsureTrusted(context.Background(), file, "https://example.com", "deepseek-harness", prompt); err != nil {
		t.Fatal(err)
	}
	if err := EnsureTrusted(context.Background(), file, "https://example.com", "hermes", prompt); err != nil {
		t.Fatal(err)
	}
	if prompts != 2 {
		t.Fatalf("prompts = %d, want 2", prompts)
	}
}

func TestLegacyOriginTrustRequiresScopedApproval(t *testing.T) {
	file := filepath.Join(t.TempDir(), "trusted-sites.json")
	if err := os.WriteFile(file, []byte("{\"version\":1,\"origins\":[\"https://example.com\"]}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	prompts := 0
	if err := EnsureTrusted(context.Background(), file, "https://example.com", "deepseek-harness", func(context.Context, TrustRequest) (bool, error) {
		prompts++
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}
	if prompts != 1 {
		t.Fatalf("prompts = %d, want 1", prompts)
	}
}

func TestEnsureTrustedDeclineAndCorruptionFailClosed(t *testing.T) {
	file := filepath.Join(t.TempDir(), "trusted-sites.json")
	if err := EnsureTrusted(context.Background(), file, "https://example.com", "deepseek-harness", func(context.Context, TrustRequest) (bool, error) { return false, nil }); !errors.Is(err, ErrTrustDeclined) {
		t.Fatalf("decline error = %v", err)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatalf("trust file exists after decline: %v", err)
	}
	if err := os.WriteFile(file, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsureTrusted(context.Background(), file, "https://example.com", "deepseek-harness", func(context.Context, TrustRequest) (bool, error) { return true, nil }); err == nil {
		t.Fatal("expected corrupt trust file rejection")
	}
}
