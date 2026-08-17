package consent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	trustedSitesVersion  = 1
	maxTrustedSitesBytes = 64 << 10
)

var ErrTrustDeclined = errors.New("server trust was not approved")

type Prompt func(context.Context, string) (bool, error)

type trustedSites struct {
	Version int      `json:"version"`
	Origins []string `json:"origins"`
}

func EnsureTrusted(ctx context.Context, filename, rawOrigin string, prompt Prompt) error {
	origin, err := canonicalOrigin(rawOrigin)
	if err != nil {
		return err
	}
	trusted, err := readTrustedSites(filename)
	if err != nil {
		return err
	}
	if containsOrigin(trusted.Origins, origin) {
		return nil
	}
	if prompt == nil {
		prompt = ConfirmServer
	}
	approved, err := prompt(ctx, origin)
	if err != nil {
		return fmt.Errorf("confirm server trust: %w", err)
	}
	if !approved {
		return ErrTrustDeclined
	}
	lock, err := acquireLock(ctx, filename+".lock")
	if err != nil {
		return err
	}
	defer lock.release()
	trusted, err = readTrustedSites(filename)
	if err != nil {
		return err
	}
	if containsOrigin(trusted.Origins, origin) {
		return nil
	}
	trusted.Version = trustedSitesVersion
	trusted.Origins = append(trusted.Origins, origin)
	sort.Strings(trusted.Origins)
	return writeTrustedSites(filename, trusted)
}

func IsTrusted(filename, rawOrigin string) (bool, error) {
	origin, err := canonicalOrigin(rawOrigin)
	if err != nil {
		return false, err
	}
	trusted, err := readTrustedSites(filename)
	if err != nil {
		return false, err
	}
	return containsOrigin(trusted.Origins, origin), nil
}

func canonicalOrigin(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", errors.New("invalid trusted server origin")
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "" || !isASCII(host) {
		return "", errors.New("trusted server host must use its ASCII domain form")
	}
	if parsed.Scheme != "https" && (parsed.Scheme != "http" || !isLoopbackHost(host)) {
		return "", errors.New("trusted server must use HTTPS or localhost HTTP")
	}
	port := parsed.Port()
	if (parsed.Scheme == "https" && port == "443") || (parsed.Scheme == "http" && port == "80") {
		port = ""
	}
	canonicalHost := host
	if port != "" {
		canonicalHost = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		canonicalHost = "[" + host + "]"
	}
	return (&url.URL{Scheme: parsed.Scheme, Host: canonicalHost}).String(), nil
}

func isASCII(value string) bool {
	for _, char := range value {
		if char > 0x7f {
			return false
		}
	}
	return true
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func readTrustedSites(filename string) (trustedSites, error) {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		return trustedSites{Version: trustedSitesVersion}, nil
	}
	if err != nil {
		return trustedSites{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxTrustedSitesBytes+1))
	if err != nil {
		return trustedSites{}, err
	}
	if len(data) > maxTrustedSitesBytes {
		return trustedSites{}, errors.New("trusted server file is too large")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var trusted trustedSites
	if err := decoder.Decode(&trusted); err != nil {
		return trustedSites{}, fmt.Errorf("parse trusted server file: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return trustedSites{}, err
	}
	if trusted.Version != trustedSitesVersion || len(trusted.Origins) > 128 {
		return trustedSites{}, errors.New("trusted server file has an unsupported format")
	}
	seen := make(map[string]struct{}, len(trusted.Origins))
	for _, origin := range trusted.Origins {
		canonical, err := canonicalOrigin(origin)
		if err != nil || canonical != origin {
			return trustedSites{}, errors.New("trusted server file contains an invalid origin")
		}
		if _, exists := seen[origin]; exists {
			return trustedSites{}, errors.New("trusted server file contains a duplicate origin")
		}
		seen[origin] = struct{}{}
	}
	return trusted, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trusted server file contains trailing JSON")
		}
		return fmt.Errorf("parse trusted server file: %w", err)
	}
	return nil
}

func containsOrigin(origins []string, target string) bool {
	for _, origin := range origins {
		if origin == target {
			return true
		}
	}
	return false
}

func writeTrustedSites(filename string, trusted trustedSites) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(trusted, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temp, err := os.CreateTemp(filepath.Dir(filename), ".trusted-sites-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	committed := false
	defer func() {
		_ = temp.Close()
		if !committed {
			_ = os.Remove(tempName)
		}
	}()
	if err := temp.Chmod(0o600); err != nil {
		return err
	}
	if _, err := temp.Write(data); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := replaceFile(tempName, filename); err != nil {
		return err
	}
	committed = true
	return os.Chmod(filename, 0o600)
}

type fileLock struct {
	filename string
}

func acquireLock(ctx context.Context, filename string) (*fileLock, error) {
	if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(15 * time.Second)
	for {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_ = file.Close()
			return &fileLock{filename: filename}, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}
		if info, statErr := os.Stat(filename); statErr == nil && time.Since(info.ModTime()) > 30*time.Second {
			_ = os.Remove(filename)
			continue
		}
		if time.Now().After(deadline) {
			return nil, errors.New("timed out waiting for trusted server lock")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (l *fileLock) release() {
	if l != nil {
		_ = os.Remove(l.filename)
	}
}
