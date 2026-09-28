package pelican

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

func parseStream(reader io.Reader) (*Generation, error) {
	scan := bufio.NewScanner(reader)
	scan.Buffer(make([]byte, 64<<10), MaxResponseBytes)
	var data bytes.Buffer
	consume := func() (*Generation, error) {
		if data.Len() == 0 {
			return nil, nil
		}
		payload := append([]byte(nil), data.Bytes()...)
		data.Reset()
		if strings.TrimSpace(string(payload)) == "[DONE]" {
			return nil, failure("invalid_output", "missing_completed_event")
		}
		var event struct {
			Type     string          `json:"type"`
			Response json.RawMessage `json:"response"`
		}
		if json.Unmarshal(payload, &event) != nil {
			return nil, failure("invalid_output", "malformed_event")
		}
		switch event.Type {
		case "response.completed":
			return parseCompleted(event.Response)
		case "response.failed", "error":
			return nil, failure("failed", "response_failed")
		case "response.incomplete":
			return nil, failure("invalid_output", "incomplete_response")
		}
		return nil, nil
	}
	for scan.Scan() {
		line := scan.Text()
		if line == "" {
			if r, err := consume(); r != nil || err != nil {
				return r, err
			}
			continue
		}
		if strings.HasPrefix(line, "data:") {
			part := strings.TrimPrefix(line, "data:")
			part = strings.TrimPrefix(part, " ")
			if data.Len()+len(part)+1 > MaxResponseBytes {
				return nil, failure("invalid_output", "response_too_large")
			}
			data.WriteString(part)
			data.WriteByte('\n')
		}
	}
	if scan.Err() != nil {
		return nil, failure("failed", "stream_interrupted")
	}
	if r, err := consume(); r != nil || err != nil {
		return r, err
	}
	return nil, failure("invalid_output", "missing_completed_event")
}

func parseCompleted(data []byte) (*Generation, error) {
	var response struct {
		ID     string          `json:"id"`
		Model  string          `json:"model"`
		Status string          `json:"status"`
		Usage  json.RawMessage `json:"usage"`
		Output []struct {
			Type    string `json:"type"`
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if json.Unmarshal(data, &response) != nil {
		return nil, failure("invalid_output", "malformed_response")
	}
	if response.Status != "completed" {
		return nil, failure("invalid_output", "incomplete_response")
	}
	var text strings.Builder
	for _, item := range response.Output {
		if item.Type != "message" || item.Role != "assistant" {
			continue
		}
		for _, part := range item.Content {
			if part.Type == "refusal" {
				return nil, failure("invalid_output", "refused")
			}
			if part.Type != "output_text" {
				continue
			}
			if text.Len()+len(part.Text) > MaxArtifactBytes {
				return nil, failure("invalid_output", "output_too_large")
			}
			text.WriteString(part.Text)
		}
	}
	if strings.TrimSpace(text.String()) == "" {
		return nil, failure("invalid_output", "empty_output")
	}
	var usage struct {
		Input  *int64 `json:"input_tokens"`
		Output *int64 `json:"output_tokens"`
		Total  *int64 `json:"total_tokens"`
	}
	if len(response.Usage) > 0 && json.Unmarshal(response.Usage, &usage) != nil {
		return nil, failure("invalid_output", "malformed_usage")
	}
	// Persist only known usage fields; never an arbitrary upstream object.
	var details struct {
		Input        *int64 `json:"input_tokens,omitempty"`
		Output       *int64 `json:"output_tokens,omitempty"`
		Total        *int64 `json:"total_tokens,omitempty"`
		InputDetails *struct {
			Cached *int64 `json:"cached_tokens,omitempty"`
		} `json:"input_tokens_details,omitempty"`
		OutputDetails *struct {
			Reasoning *int64 `json:"reasoning_tokens,omitempty"`
		} `json:"output_tokens_details,omitempty"`
	}
	_ = json.Unmarshal(response.Usage, &details)
	cleanUsage, _ := json.Marshal(details)
	if len(response.ID) > 256 || len(response.Model) > 256 {
		return nil, failure("invalid_output", "invalid_response_metadata")
	}
	return &Generation{Text: text.String(), Model: response.Model, ResponseID: response.ID,
		InputTokens: usage.Input, OutputTokens: usage.Output, TotalTokens: usage.Total, Usage: cleanUsage}, nil
}
