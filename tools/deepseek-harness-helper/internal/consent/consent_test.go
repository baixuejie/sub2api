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
	prompt := func(_ context.Context, origin string) (bool, error) {
		prompts++
		if origin != "https://example.com" {
			t.Fatalf("origin = %q", origin)
		}
		return true, nil
	}
	if err := EnsureTrusted(context.Background(), file, "HTTPS://EXAMPLE.COM:443/", prompt); err != nil {
		t.Fatal(err)
	}
	if err := EnsureTrusted(context.Background(), file, "https://example.com", prompt); err != nil {
		t.Fatal(err)
	}
	if prompts != 1 {
		t.Fatalf("prompts = %d, want 1", prompts)
	}
	trusted, err := IsTrusted(file, "https://example.com/")
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

func TestEnsureTrustedDeclineAndCorruptionFailClosed(t *testing.T) {
	file := filepath.Join(t.TempDir(), "trusted-sites.json")
	if err := EnsureTrusted(context.Background(), file, "https://example.com", func(context.Context, string) (bool, error) { return false, nil }); !errors.Is(err, ErrTrustDeclined) {
		t.Fatalf("decline error = %v", err)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatalf("trust file exists after decline: %v", err)
	}
	if err := os.WriteFile(file, []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsureTrusted(context.Background(), file, "https://example.com", func(context.Context, string) (bool, error) { return true, nil }); err == nil {
		t.Fatal("expected corrupt trust file rejection")
	}
}
