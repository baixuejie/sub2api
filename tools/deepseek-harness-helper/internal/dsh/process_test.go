package dsh

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
)

func TestStopManagedRejectsMismatchedRuntimeWithoutKillingProcess(t *testing.T) {
	t.Parallel()
	paths := config.PathsFor(t.TempDir())
	if err := config.EnsurePrivateDir(paths.DataDir); err != nil {
		t.Fatal(err)
	}
	if err := writeState(paths.StateFile, ProcessState{
		PID: os.Getpid(), StartedAt: time.Now().UTC(), NodePath: "different-node", DSHBin: "dsh",
		DSHVersion: "0.1.0-rc.6", ConfigurationID: "old", URL: "http://127.0.0.1:3080",
	}); err != nil {
		t.Fatal(err)
	}
	if err := StopManaged(context.Background(), Environment{NodePath: "node"}, paths, "dsh", "0.1.0-rc.6"); err == nil {
		t.Fatal("expected mismatched runtime rejection")
	}
	if _, err := os.Stat(paths.StateFile); err != nil {
		t.Fatalf("state was removed after mismatch: %v", err)
	}
}

func TestStopManagedRemovesDeadTrackedState(t *testing.T) {
	t.Parallel()
	paths := config.PathsFor(t.TempDir())
	if err := config.EnsurePrivateDir(paths.DataDir); err != nil {
		t.Fatal(err)
	}
	if err := writeState(paths.StateFile, ProcessState{
		PID: 2147483647, StartedAt: time.Now().UTC(), NodePath: "node", DSHBin: "dsh",
		DSHVersion: "0.1.0-rc.6", ConfigurationID: "old", URL: "http://127.0.0.1:3080",
	}); err != nil {
		t.Fatal(err)
	}
	if err := StopManaged(context.Background(), Environment{NodePath: "node"}, paths, "dsh", "0.1.0-rc.6"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(paths.StateFile); !os.IsNotExist(err) {
		t.Fatalf("dead state remains: %v", err)
	}
}

func TestProcessIdentityRejectsReusedPIDTimestamp(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	startedAt, err := processStartTime(os.Getpid(), executable, executable)
	if err != nil {
		t.Fatal(err)
	}
	state := ProcessState{PID: os.Getpid(), StartedAt: startedAt.Add(-time.Minute), NodePath: executable, DSHBin: executable}
	if processIdentityMatches(state) {
		t.Fatal("stale process start time matched current PID")
	}
}

func TestStartOrReuseKeepsHealthyProcessAcrossConfigurationChanges(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><title>DeepSeek Harness</title></html>"))
	}))
	defer server.Close()
	paths := config.PathsFor(t.TempDir())
	if err := config.EnsurePrivateDir(paths.DataDir); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	startedAt, err := processStartTime(os.Getpid(), executable, executable)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeState(paths.StateFile, ProcessState{
		PID: os.Getpid(), StartedAt: startedAt, NodePath: executable, DSHBin: executable,
		DSHVersion: "0.1.0-rc.6", ConfigurationID: "old", URL: server.URL,
	}); err != nil {
		t.Fatal(err)
	}
	result, err := StartOrReuse(context.Background(), Environment{NodePath: executable}, paths, executable, "0.1.0-rc.6", "new")
	if err != nil {
		t.Fatal(err)
	}
	if result.URL != server.URL {
		t.Fatalf("unexpected result: %#v", result)
	}
	state, err := readState(paths.StateFile)
	if err != nil {
		t.Fatal(err)
	}
	if state.ConfigurationID != "new" {
		t.Fatalf("configuration id = %q, want new", state.ConfigurationID)
	}
}

func TestAnnouncedURLSinceReadsOnlyCurrentLaunchOutput(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "dsh.log")
	previous := "dsh web: http://127.0.0.1:3000\n"
	current := "starting\ndsh web: http://127.0.0.1:3080\n"
	if err := os.WriteFile(file, []byte(previous+current), 0o600); err != nil {
		t.Fatal(err)
	}
	url, err := announcedURLSince(file, int64(len(previous)))
	if err != nil {
		t.Fatal(err)
	}
	if url != "http://127.0.0.1:3080" {
		t.Fatalf("url = %q", url)
	}
}

func TestAnnouncedURLSinceRejectsOversizedStartupOutput(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "dsh.log")
	if err := os.WriteFile(file, []byte(strings.Repeat("x", maxStartupLogBytes+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := announcedURLSince(file, 0); err == nil {
		t.Fatal("expected oversized startup log rejection")
	}
}
