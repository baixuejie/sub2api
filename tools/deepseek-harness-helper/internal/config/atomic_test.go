package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWithFileLockUsesDSHExclusiveSiblingProtocol(t *testing.T) {
	t.Parallel()
	filename := filepath.Join(t.TempDir(), "settings.yaml")
	lockPath := filename + ".lock"
	if err := os.WriteFile(lockPath, []byte("other-writer\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	released := make(chan struct{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = os.Remove(lockPath)
		close(released)
	}()
	if err := withFileLock(filename, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	<-released
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("writer lock remains after operation: %v", err)
	}
}
