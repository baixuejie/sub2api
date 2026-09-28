package pelican

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Generation struct {
	Text         string
	Model        string
	ResponseID   string
	InputTokens  *int64
	OutputTokens *int64
	TotalTokens  *int64
	Usage        json.RawMessage
}

type generationError struct{ status, code string }

func (e *generationError) Error() string { return e.code }
func failure(status, code string) error  { return &generationError{status: status, code: code} }

type Generator struct {
	endpoint string
	client   *http.Client
}

func LocalEndpoint(host string, port int) string {
	host = strings.Trim(host, "[]")
	switch host {
	case "", "0.0.0.0", "*":
		host = "127.0.0.1"
	case "::":
		host = "::1"
	}
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port)) + "/v1/responses"
}

func NewGenerator(endpoint string) *Generator {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.MaxIdleConnsPerHost = 2
	transport.ResponseHeaderTimeout = 60 * time.Second
	return &Generator{endpoint: endpoint, client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (g *Generator) Close() { g.client.CloseIdleConnections() }

func (g *Generator) Generate(ctx context.Context, key, prompt string, maxTokens int) (*Generation, error) {
	body, err := json.Marshal(map[string]any{
		"model": Model, "reasoning": map[string]string{"effort": ReasoningEffort}, "input": prompt,
		"max_output_tokens": maxTokens, "stream": true, "store": false,
	})
	if err != nil {
		return nil, failure("failed", "request_encoding")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, failure("failed", "request_configuration")
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, failure("failed", "request_failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, failure("failed", "http_"+strconv.Itoa(resp.StatusCode))
	}
	var result *Generation
	limited := &io.LimitedReader{R: resp.Body, N: MaxResponseBytes + 1}
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/event-stream") {
		result, err = parseStream(limited)
	} else {
		data, readErr := io.ReadAll(limited)
		if readErr != nil {
			return nil, failure("failed", "response_read")
		}
		result, err = parseCompleted(data)
	}
	if limited.N <= 0 {
		return nil, failure("invalid_output", "response_too_large")
	}
	if err != nil {
		return nil, err
	}
	// A misbehaving endpoint can echo Authorization in its output. Never publish it.
	if key != "" && (strings.Contains(result.Text, key) || strings.Contains(result.Model, key) || strings.Contains(result.ResponseID, key) || bytes.Contains(result.Usage, []byte(key))) {
		return nil, failure("invalid_output", "credential_echo")
	}
	return result, nil
}
