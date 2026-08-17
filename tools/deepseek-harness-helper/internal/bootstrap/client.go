package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/dsh"
)

const maxResponseBytes = 1 << 20

type Client struct {
	HTTP *http.Client
}

type HTTPStatusError struct {
	Status  int
	Message string
}

func (e *HTTPStatusError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("HTTP %d", e.Status)
	}
	return fmt.Sprintf("HTTP %d: %s", e.Status, e.Message)
}

type ReportOutcomeUnknownError struct {
	Cause error
}

func (e *ReportOutcomeUnknownError) Error() string { return e.Cause.Error() }
func (e *ReportOutcomeUnknownError) Unwrap() error { return e.Cause }

func IsReportOutcomeUnknown(err error) bool {
	var target *ReportOutcomeUnknownError
	return errors.As(err, &target)
}

func NewClient() *Client {
	return &Client{HTTP: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}}
}

func (c *Client) Exchange(ctx context.Context, launch LaunchRequest) (Task, error) {
	server, err := ValidateServerURL(launch.Server)
	if err != nil {
		return Task{}, err
	}
	endpoint := server.ResolveReference(&url.URL{Path: "/api/v1/deepseek-harness/exchange"})
	var envelope Envelope[Task]
	if err := c.postJSON(ctx, endpoint.String(), "", ExchangeRequest{Ticket: launch.Ticket}, &envelope); err != nil {
		return Task{}, fmt.Errorf("exchange bootstrap ticket: %w", err)
	}
	task := envelope.Data
	task.ServerOrigin = launch.Server
	if err := validateTask(task, launch); err != nil {
		return Task{}, err
	}
	return task, nil
}

func (c *Client) Report(ctx context.Context, task Task, event StatusEvent) error {
	if _, err := ValidateStatusURL(task.StatusURL, task.ServerOrigin, task.OperationID); err != nil {
		return err
	}
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		var response json.RawMessage
		lastErr = c.postJSON(ctx, task.StatusURL, task.EventToken, event, &response)
		if lastErr == nil {
			return nil
		}
		retry, _ := reportErrorDisposition(lastErr)
		if !retry {
			return fmt.Errorf("report %s: %w", event.Status, lastErr)
		}
		if attempt < 4 {
			if err := waitForRetry(ctx, time.Duration(200*(1<<attempt))*time.Millisecond); err != nil {
				lastErr = errors.Join(lastErr, err)
				break
			}
		}
	}
	_, outcomeUnknown := reportErrorDisposition(lastErr)
	wrapped := fmt.Errorf("report %s: %w", event.Status, lastErr)
	if outcomeUnknown {
		return &ReportOutcomeUnknownError{Cause: wrapped}
	}
	return wrapped
}

func reportErrorDisposition(err error) (retry, outcomeUnknown bool) {
	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) {
		if statusErr.Status == http.StatusTooManyRequests {
			return true, false
		}
		if statusErr.Status >= 400 && statusErr.Status < 500 && statusErr.Status != http.StatusRequestTimeout {
			return false, false
		}
		return true, true
	}
	return true, true
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *Client) postJSON(ctx context.Context, endpoint, bearer string, payload, output any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	client := c.HTTP
	if client == nil {
		client = NewClient().HTTP
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return err
	}
	if len(data) > maxResponseBytes {
		return errors.New("response body is too large")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(string(data))
		if len(message) > 500 {
			message = message[:500]
		}
		return &HTTPStatusError{Status: resp.StatusCode, Message: message}
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return errors.New("empty JSON response")
	}
	if err := json.Unmarshal(data, output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func validateTask(task Task, launch LaunchRequest) error {
	if task.OperationID == "" || task.OperationID != launch.OperationID || task.EventToken == "" || task.APIKey == "" {
		return errors.New("exchange returned an incomplete task")
	}
	if task.DSHVersion != dsh.SupportedVersion {
		return fmt.Errorf("exchange returned unsupported dsh_version: expected %s", dsh.SupportedVersion)
	}
	if len(task.EventToken) > 4096 || len(task.APIKey) > 64<<10 {
		return errors.New("exchange returned an oversized credential")
	}
	if _, err := ValidateStatusURL(task.StatusURL, launch.Server, task.OperationID); err != nil {
		return err
	}
	p := task.Provider
	if p.Route == "" || p.DisplayName == "" || p.Protocol == "" || p.BaseURL == "" || p.CredentialName == "" || p.Model.ID == "" || p.Model.Name == "" || p.Model.ContextWindow <= 0 || p.Model.MaxTokens <= 0 {
		return errors.New("exchange returned an incomplete provider")
	}
	if p.CredentialName != "SUB2API_API_KEY" || !allowedProviderProtocol(p.Route, p.Protocol) {
		return errors.New("exchange returned an unsupported provider contract")
	}
	if len(p.DisplayName) > 128 || len(p.Model.ID) > 256 || len(p.Model.Name) > 256 || p.Model.ContextWindow > 10_000_000 || p.Model.MaxTokens > 1_000_000 {
		return errors.New("exchange returned invalid provider limits")
	}
	base, err := url.Parse(p.BaseURL)
	if err != nil || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" || base.EscapedPath() != "/v1" {
		return errors.New("provider base_url must end at /v1")
	}
	if base.Scheme != "https" && (base.Scheme != "http" || !isLoopbackHost(strings.TrimSuffix(strings.ToLower(base.Hostname()), "."))) {
		return errors.New("provider base_url must use HTTPS or localhost HTTP")
	}
	return nil
}

func allowedProviderProtocol(route, protocol string) bool {
	allowed := map[string]map[string]struct{}{
		"sub2api-openai":      {"openai-responses": {}},
		"sub2api-anthropic":   {"anthropic-messages": {}},
		"sub2api-grok":        {"openai-responses": {}},
		"sub2api-gemini":      {"openai-completions": {}},
		"sub2api-antigravity": {"anthropic-messages": {}, "openai-completions": {}},
		"sub2api-composite":   {"openai-completions": {}},
	}
	protocols, ok := allowed[route]
	if !ok {
		return false
	}
	_, ok = protocols[protocol]
	return ok
}
