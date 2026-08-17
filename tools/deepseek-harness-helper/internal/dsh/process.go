package dsh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
)

var webURLPattern = regexp.MustCompile(`(?m)^dsh web: (http://(?:127\.0\.0\.1|\[::1\]|localhost):[0-9]+)\r?$`)

const maxStartupLogBytes = 2 << 20

func StopManaged(ctx context.Context, env Environment, paths config.Paths, dshBin, version string) error {
	if err := config.EnsurePrivateDir(paths.DataDir); err != nil {
		return err
	}
	lock, err := acquireStartLock(paths.StateFile+".lock", 30*time.Second)
	if err != nil {
		return err
	}
	defer lock.release()
	state, err := readState(paths.StateFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read Helper-managed DSH state: %w", err)
	}
	if state.DSHVersion != version || !samePath(state.NodePath, env.NodePath) || !samePath(state.DSHBin, dshBin) {
		return errors.New("existing Helper-managed DSH runtime does not match the pinned installation")
	}
	if processAlive(state.PID) {
		if !processIdentityMatches(state) {
			return errors.New("tracked DSH process identity does not match; refusing to terminate its PID")
		}
		if err := terminatePID(state.PID); err != nil && processAlive(state.PID) {
			return fmt.Errorf("stop Helper-managed DSH process: %w", err)
		}
		deadline := time.Now().Add(3 * time.Second)
		for processAlive(state.PID) && time.Now().Before(deadline) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
			}
		}
		if processAlive(state.PID) {
			return errors.New("timed out stopping Helper-managed DSH process")
		}
	}
	if err := os.Remove(paths.StateFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

type StartResult struct {
	URL string
}

func StartOrReuse(ctx context.Context, env Environment, paths config.Paths, dshBin, version, configurationID string) (StartResult, error) {
	if strings.TrimSpace(configurationID) == "" {
		return StartResult{}, errors.New("configuration id is required")
	}
	if err := config.EnsurePrivateDir(paths.DataDir); err != nil {
		return StartResult{}, err
	}
	lock, err := acquireStartLock(paths.StateFile+".lock", 30*time.Second)
	if err != nil {
		return StartResult{}, err
	}
	defer lock.release()
	state, stateErr := readState(paths.StateFile)
	if stateErr == nil {
		if state.DSHVersion != version || !samePath(state.NodePath, env.NodePath) || !samePath(state.DSHBin, dshBin) {
			return StartResult{}, errors.New("existing Helper-managed DSH runtime does not match the pinned installation")
		}
		if processAlive(state.PID) {
			if !processIdentityMatches(state) {
				return StartResult{}, errors.New("tracked DSH process identity does not match; refusing to reuse or terminate its PID")
			}
			if healthyLoopback(ctx, state.URL) {
				state.ConfigurationID = configurationID
				if err := writeState(paths.StateFile, state); err != nil {
					return StartResult{}, err
				}
				return StartResult{URL: state.URL}, nil
			}
			if err := terminatePID(state.PID); err != nil {
				return StartResult{}, fmt.Errorf("stop unhealthy Helper-managed DSH process: %w", err)
			}
		}
	} else if !os.IsNotExist(stateErr) {
		return StartResult{}, fmt.Errorf("read Helper-managed DSH state: %w", stateErr)
	}
	if err := os.Remove(paths.StateFile); err != nil && !os.IsNotExist(err) {
		return StartResult{}, err
	}
	startOffset := int64(0)
	if info, statErr := os.Stat(paths.LogFile); statErr == nil {
		startOffset = info.Size()
	}
	logFile, err := os.OpenFile(paths.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return StartResult{}, err
	}
	args := []string{dshBin, "--profile", "web", "--host", "127.0.0.1", "--port", "0"}
	cmd := exec.Command(env.NodePath, args...)
	cmd.Dir = paths.DataDir
	cmd.Env = childEnvironment(paths.DSHHome)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	configureProcess(cmd)
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return StartResult{}, err
	}
	// The child owns duplicated log handles after Start. Closing the parent's handle
	// avoids keeping a pipe whose reader disappears when the Helper exits.
	_ = logFile.Close()
	waitCh := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		removeStateForPID(paths.StateFile, cmd.Process.Pid)
		waitCh <- err
	}()
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			harnessURL, readErr := announcedURLSince(paths.LogFile, startOffset)
			if readErr != nil {
				_ = terminateProcess(cmd)
				return StartResult{}, readErr
			}
			if harnessURL == "" {
				continue
			}
			if !healthyLoopback(ctx, harnessURL) {
				_ = terminateProcess(cmd)
				return StartResult{}, errors.New("DSH announced a URL that is not reachable")
			}
			startedAt, identityErr := processStartTime(cmd.Process.Pid, env.NodePath, dshBin)
			if identityErr != nil {
				_ = terminateProcess(cmd)
				return StartResult{}, fmt.Errorf("verify DSH process identity: %w", identityErr)
			}
			state := ProcessState{PID: cmd.Process.Pid, StartedAt: startedAt, NodePath: env.NodePath, DSHBin: dshBin, DSHVersion: version, ConfigurationID: configurationID, URL: harnessURL}
			if err := writeState(paths.StateFile, state); err != nil {
				_ = terminateProcess(cmd)
				return StartResult{}, err
			}
			return StartResult{URL: harnessURL}, nil
		case err := <-waitCh:
			if err == nil {
				return StartResult{}, errors.New("DSH exited before startup")
			}
			return StartResult{}, fmt.Errorf("DSH exited before startup: %w", err)
		case <-ctx.Done():
			_ = terminateProcess(cmd)
			return StartResult{}, ctx.Err()
		case <-timer.C:
			_ = terminateProcess(cmd)
			return StartResult{}, errors.New("timed out waiting for DSH startup URL")
		}
	}
}

func announcedURLSince(filename string, offset int64) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if info.Size() < offset {
		return "", errors.New("DSH log was truncated during startup")
	}
	size := info.Size() - offset
	if size == 0 {
		return "", nil
	}
	if size > maxStartupLogBytes {
		return "", errors.New("DSH startup log exceeded the size limit")
	}
	content := make([]byte, size)
	if _, err = file.ReadAt(content, offset); err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	matches := webURLPattern.FindSubmatch(content)
	if matches == nil {
		return "", nil
	}
	return string(matches[1]), nil
}

func childEnvironment(dshHome string) []string {
	environment := make([]string, 0, len(os.Environ())+1)
	for _, entry := range os.Environ() {
		name, _, found := strings.Cut(entry, "=")
		if found && (strings.EqualFold(name, "DSH_HOME") || strings.EqualFold(name, "SUB2API_API_KEY")) {
			continue
		}
		environment = append(environment, entry)
	}
	return append(environment, "DSH_HOME="+dshHome)
}

func healthyLoopback(ctx context.Context, raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.User != nil || u.Port() == "" {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if host != "localhost" && host != "127.0.0.1" && host != "::1" {
		return false
	}
	requestCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, raw, nil)
	if err != nil {
		return false
	}
	client := &http.Client{Timeout: 2 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	return err == nil && strings.Contains(string(body), "<title>DeepSeek Harness</title>")
}
