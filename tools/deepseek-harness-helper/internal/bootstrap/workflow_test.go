package bootstrap

import (
	"context"
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
		ConfirmTrust: func(context.Context, string) (bool, error) {
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
