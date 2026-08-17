package dsh

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/tools/deepseek-harness-helper/internal/config"
)

func TestPinnedDSHLoadsGeneratedConfiguration(t *testing.T) {
	nodePath := os.Getenv("DSH_SMOKE_NODE")
	dshBin := os.Getenv("DSH_SMOKE_BIN")
	if nodePath == "" || dshBin == "" {
		t.Skip("set DSH_SMOKE_NODE and DSH_SMOKE_BIN to run the pinned DSH smoke test")
	}
	if _, err := os.Stat(nodePath); err != nil {
		t.Fatalf("stat DSH_SMOKE_NODE: %v", err)
	}
	if _, err := os.Stat(dshBin); err != nil {
		t.Fatalf("stat DSH_SMOKE_BIN: %v", err)
	}

	paths := config.PathsFor(t.TempDir())
	provider := config.ProviderConfig{
		Route:          "sub2api-openai",
		DisplayName:    "Sub2API OpenAI",
		Protocol:       "openai-responses",
		BaseURL:        "https://api.example.com",
		CredentialName: "SUB2API_API_KEY",
		ModelID:        "gpt-5.6-sol",
		ModelName:      "GPT-5.6 Sol",
		ContextWindow:  1_050_000,
		MaxTokens:      128_000,
	}
	if err := config.Apply(paths, provider, "ci-smoke-placeholder-key"); err != nil {
		t.Fatalf("apply generated configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	environment := Environment{NodePath: nodePath}
	result, err := StartOrReuse(ctx, environment, paths, dshBin, SupportedVersion, "ci-smoke-operation")
	if err != nil {
		t.Fatalf("start pinned DSH with generated configuration: %v", err)
	}
	t.Cleanup(func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer stopCancel()
		if err := StopManaged(stopCtx, environment, paths, dshBin, SupportedVersion); err != nil {
			t.Errorf("stop pinned DSH smoke process: %v", err)
		}
	})
	if result.URL == "" {
		t.Fatal("pinned DSH did not return a loopback URL")
	}
	if err := requireSmokeProvider(ctx, result.URL, provider.Route); err != nil {
		t.Fatal(err)
	}
}

func requireSmokeProvider(ctx context.Context, harnessURL, route string) error {
	const rpcID = "44fe4f75-2bcb-46f4-b8c5-2469f195d1e1"
	payload, err := json.Marshal(map[string]any{
		"type":    "client-request",
		"rpcId":   rpcID,
		"method":  "llm.providers",
		"payload": map[string]any{},
	})
	if err != nil {
		return err
	}
	endpoint := strings.TrimRight(harnessURL, "/") + "/api/llm.providers"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("query pinned DSH providers: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("query pinned DSH providers: unexpected HTTP status %d", response.StatusCode)
	}
	var envelope struct {
		Type   string `json:"type"`
		RPCID  string `json:"rpcId"`
		Result struct {
			OK    bool `json:"ok"`
			Value struct {
				Providers []struct {
					Provider string `json:"provider"`
					Active   bool   `json:"active"`
					Declared bool   `json:"declared"`
				} `json:"providers"`
			} `json:"value"`
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		} `json:"result"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(&envelope); err != nil {
		return fmt.Errorf("decode pinned DSH provider response: %w", err)
	}
	if envelope.Type != "server-response" || envelope.RPCID != rpcID {
		return errors.New("pinned DSH provider response did not preserve the RPC envelope")
	}
	if !envelope.Result.OK {
		return fmt.Errorf("pinned DSH rejected generated provider config: %s: %s", envelope.Result.Error.Code, envelope.Result.Error.Message)
	}
	for _, provider := range envelope.Result.Value.Providers {
		if provider.Provider == route {
			if !provider.Active || !provider.Declared {
				return fmt.Errorf("pinned DSH provider %q is not active and declared", route)
			}
			return nil
		}
	}
	return fmt.Errorf("pinned DSH did not load generated provider %q", route)
}
