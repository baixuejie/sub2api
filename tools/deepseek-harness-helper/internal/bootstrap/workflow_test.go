package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/consent"
)

func TestWorkflowDeclinedTrustDoesNotContactServer(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()
	launchURI := "sub2api-harness://bootstrap?server=" + url.QueryEscape(server.URL) + "&ticket=ticket&operation_id=operation"
	workflow := Workflow{
		Paths: config.PathsFor(t.TempDir()),
		ConfirmTrust: func(context.Context, consent.TrustRequest) (bool, error) {
			return false, nil
		},
	}
	_, err := workflow.Run(context.Background(), launchURI)
	if !errors.Is(err, consent.ErrTrustDeclined) {
		t.Fatalf("error = %v", err)
	}
	if requests.Load() != 0 {
		t.Fatalf("server received %d requests before trust", requests.Load())
	}
}

func TestWorkflowDispatchesThroughRegisteredAdapter(t *testing.T) {
	t.Parallel()
	var executions atomic.Int32
	var completed atomic.Int32
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	mux.HandleFunc("/api/v1/test-extension/exchange", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(Envelope[Task]{Data: Task{
			OperationID: "operation", EventToken: "event",
			StatusURL:       server.URL + "/api/v1/test-extension/sessions/operation/events",
			ProtocolVersion: CurrentTaskProtocolVersion, ToolID: "test-adapter",
			ToolVersion: "2.3.4", MinimumHelperVersion: "0.1.0",
		}})
	})
	mux.HandleFunc("/api/v1/test-extension/sessions/operation/events", func(w http.ResponseWriter, r *http.Request) {
		var event StatusEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Fatal(err)
		}
		if event.Status == StatusCompleted {
			completed.Add(1)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"status": event.Status}})
	})
	registry, err := NewAdapterRegistry(recordingAdapter{executions: &executions})
	if err != nil {
		t.Fatal(err)
	}
	workflow := Workflow{
		Paths: config.PathsFor(t.TempDir()), Registry: registry, HelperVersion: "0.1.0",
		ConfirmTrust: func(context.Context, consent.TrustRequest) (bool, error) { return true, nil },
	}
	launchURI := "sub2api-harness://bootstrap?server=" + url.QueryEscape(server.URL) + "&ticket=ticket&operation_id=operation&extension_id=test-extension"
	harnessURL, err := workflow.Run(context.Background(), launchURI)
	if err != nil {
		t.Fatal(err)
	}
	if harnessURL != "http://127.0.0.1:39000" || executions.Load() != 1 || completed.Load() != 1 {
		t.Fatalf("url=%q executions=%d completed=%d", harnessURL, executions.Load(), completed.Load())
	}
}

func TestWorkflowReportsHelperUpgradeRequirement(t *testing.T) {
	t.Parallel()
	var reported StatusEvent
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	mux.HandleFunc("/api/v1/test-extension/exchange", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(Envelope[Task]{Data: Task{
			OperationID: "operation", EventToken: "event",
			StatusURL:       server.URL + "/api/v1/test-extension/sessions/operation/events",
			ProtocolVersion: CurrentTaskProtocolVersion, ToolID: "test-adapter",
			ToolVersion: "2.3.4", MinimumHelperVersion: "0.2.0",
		}})
	})
	mux.HandleFunc("/api/v1/test-extension/sessions/operation/events", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&reported); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"status": reported.Status}})
	})
	registry, err := NewAdapterRegistry(recordingAdapter{executions: &atomic.Int32{}})
	if err != nil {
		t.Fatal(err)
	}
	workflow := Workflow{
		Paths: config.PathsFor(t.TempDir()), Registry: registry, HelperVersion: "0.1.0",
		ConfirmTrust: func(context.Context, consent.TrustRequest) (bool, error) { return true, nil },
	}
	launchURI := "sub2api-harness://bootstrap?server=" + url.QueryEscape(server.URL) + "&ticket=ticket&operation_id=operation&extension_id=test-extension"
	_, err = workflow.Run(context.Background(), launchURI)
	var upgradeRequired *HelperUpgradeRequiredError
	if !errors.As(err, &upgradeRequired) {
		t.Fatalf("error = %v", err)
	}
	if reported.Status != StatusFailed || reported.ErrorCode != "helper_update_required" || reported.Progress != 100 {
		t.Fatalf("reported event = %#v", reported)
	}
}

type recordingAdapter struct {
	executions *atomic.Int32
}

func (recordingAdapter) ToolID() string { return "test-adapter" }

func (recordingAdapter) AllowedExtensionIDs() []string { return []string{"test-extension"} }

func (recordingAdapter) Validate(task Task) error {
	if task.ToolVersion != "2.3.4" {
		return errors.New("unexpected tool version")
	}
	return nil
}

func (a recordingAdapter) Execute(_ context.Context, execution AdapterExecution) (AdapterResult, error) {
	if execution.Task.ToolID != a.ToolID() {
		return AdapterResult{}, errors.New("workflow dispatched the wrong task")
	}
	a.executions.Add(1)
	return AdapterResult{OpenURL: "http://127.0.0.1:39000", CompletionMessage: "ready"}, nil
}
