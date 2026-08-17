package bootstrap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientReportRetriesAmbiguousSuccessResponse(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			_, _ = w.Write([]byte("not-json"))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"status": StatusCompleted}})
	}))
	defer server.Close()
	task := Task{
		ServerOrigin: server.URL,
		OperationID:  "operation",
		EventToken:   "event-token",
		StatusURL:    server.URL + "/api/v1/deepseek-harness/sessions/operation/events",
	}
	err := NewClient().Report(context.Background(), task, StatusEvent{Status: StatusCompleted, Stage: StatusCompleted, Message: "Ready", Progress: 100, HarnessURL: "http://127.0.0.1:3080"})
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestClientReportDoesNotRetryDefiniteRejection(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "disabled", http.StatusNotFound)
	}))
	defer server.Close()
	task := Task{
		ServerOrigin: server.URL,
		OperationID:  "operation",
		EventToken:   "event-token",
		StatusURL:    server.URL + "/api/v1/deepseek-harness/sessions/operation/events",
	}
	err := NewClient().Report(context.Background(), task, StatusEvent{Status: StatusStarting, Stage: StatusStarting, Message: "Starting", Progress: 80})
	if err == nil || IsReportOutcomeUnknown(err) {
		t.Fatalf("error = %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}

func TestClientExchangeAndReport(t *testing.T) {
	t.Parallel()
	var statusEvent StatusEvent
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	mux.HandleFunc("/api/v1/test-tools/exchange", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.Header.Get("Content-Type"))
		}
		var request ExchangeRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Ticket != "ticket" {
			t.Fatalf("bad request: %#v, %v", request, err)
		}
		_ = json.NewEncoder(w).Encode(Envelope[Task]{Data: Task{
			OperationID: "op", EventToken: "event", StatusURL: server.URL + "/api/v1/test-tools/sessions/op/events",
			DSHVersion: "0.1.0-rc.6", APIKey: "secret",
			Provider: Provider{Route: "sub2api-openai", DisplayName: "Sub2API", Protocol: "openai-responses", BaseURL: "https://api.example", CredentialName: "SUB2API_API_KEY", Model: Model{ID: "m", Name: "M", ContextWindow: 128000, MaxTokens: 8192}},
		}})
	})
	mux.HandleFunc("/api/v1/test-tools/sessions/op/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer event" {
			t.Fatalf("missing bearer token: %q", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&statusEvent); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"status": statusEvent.Status}})
	})
	launch := LaunchRequest{Server: server.URL, Ticket: "ticket", OperationID: "op", ExtensionID: "test-tools"}
	client := NewClient()
	task, err := client.Exchange(context.Background(), launch)
	if err != nil {
		t.Fatal(err)
	}
	if task.ServerOrigin != server.URL || task.ExtensionID != "test-tools" || task.APIKey != "secret" {
		t.Fatalf("unexpected task: %#v", task)
	}
	if err := client.Report(context.Background(), task, StatusEvent{Status: StatusInstalling, Stage: StatusInstalling, Message: "Installing", Progress: 30}); err != nil {
		t.Fatal(err)
	}
	if statusEvent.Status != StatusInstalling || statusEvent.Progress != 30 {
		t.Fatalf("unexpected event: %#v", statusEvent)
	}
}
