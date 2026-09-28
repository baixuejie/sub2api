// Package pelican implements the opt-in hourly code-generation gallery.
package pelican

import (
	"errors"
	"strings"
	"time"
)

const (
	Model                = "gpt-6-astra"
	ReasoningEffort      = "high"
	PreviewPolicyVersion = 1
	MaxArtifactBytes     = 1 << 20
	MaxResponseBytes     = 4 << 20
)

var (
	ErrConflict    = errors.New("configuration changed; reload and try again")
	ErrGroups      = errors.New("selected groups must be active and available")
	ErrKeyRequired = errors.New("an API key is required before enabling generation")
	ErrNotFound    = errors.New("not found")
)

type Encryptor interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

type Config struct {
	Revision         int64      `json:"revision"`
	Enabled          bool       `json:"enabled"`
	SelectedGroupIDs []int64    `json:"selected_group_ids"`
	TopicMode        string     `json:"topic_mode"`
	FixedTopicID     string     `json:"fixed_topic_id"`
	MaxOutputTokens  int        `json:"max_output_tokens"`
	TimeoutSeconds   int        `json:"timeout_seconds"`
	RetentionDays    int        `json:"retention_days"`
	NextRunAt        *time.Time `json:"next_run_at"`
	KeyConfigured    bool       `json:"key_configured"`
	KeyMasked        string     `json:"key_masked"`
	EncryptionReady  bool       `json:"encryption_ready"`
	KeyUnavailable   bool       `json:"key_unavailable"`
	Model            string     `json:"model"`
	ReasoningEffort  string     `json:"reasoning_effort"`
	EncryptedKey     string     `json:"-"`
	Sequence         int64      `json:"-"`
}

type SaveConfig struct {
	Revision         int64   `json:"revision"`
	Enabled          bool    `json:"enabled"`
	APIKey           *string `json:"api_key,omitempty"`
	SelectedGroupIDs []int64 `json:"selected_group_ids"`
	TopicMode        string  `json:"topic_mode"`
	FixedTopicID     string  `json:"fixed_topic_id"`
	MaxOutputTokens  int     `json:"max_output_tokens"`
	TimeoutSeconds   int     `json:"timeout_seconds"`
	RetentionDays    int     `json:"retention_days"`
}

func (c *SaveConfig) Validate() error {
	if c.Revision < 1 || (c.TopicMode != "rotate" && c.TopicMode != "fixed") || topicByID(c.FixedTopicID) == nil {
		return errors.New("invalid configuration revision or topic")
	}
	if c.MaxOutputTokens < 4096 || c.MaxOutputTokens > 32768 || c.TimeoutSeconds < 60 || c.TimeoutSeconds > 600 || c.RetentionDays < 7 || c.RetentionDays > 90 {
		return errors.New("output tokens, timeout or retention is outside the allowed range")
	}
	if len(c.SelectedGroupIDs) > 200 {
		return errors.New("select at most 200 group labels")
	}
	ids := make([]int64, 0, len(c.SelectedGroupIDs))
	seen := map[int64]bool{}
	for _, id := range c.SelectedGroupIDs {
		if id <= 0 {
			return ErrGroups
		}
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	c.SelectedGroupIDs = ids
	if c.APIKey != nil {
		key := strings.TrimSpace(*c.APIKey)
		if key == "" || len(key) > 1024 || strings.ContainsAny(key, "\r\n\t ") {
			return errors.New("API key must be nonempty and contain no whitespace; omit it to keep the saved key")
		}
		c.APIKey = &key
	}
	return nil
}

func nextHour(now time.Time) time.Time { return now.UTC().Truncate(time.Hour).Add(time.Hour) }

type GroupTag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Run struct {
	ID              int64      `json:"id"`
	ScheduledFor    time.Time  `json:"scheduled_for"`
	Status          string     `json:"status"`
	TopicID         string     `json:"topic_id"`
	Groups          []GroupTag `json:"groups"`
	Model           string     `json:"model"`
	ReasoningEffort string     `json:"reasoning_effort"`
	InputTokens     *int64     `json:"input_tokens"`
	OutputTokens    *int64     `json:"output_tokens"`
	TotalTokens     *int64     `json:"total_tokens"`
	LatencyMS       *int64     `json:"latency_ms"`
	FinishedAt      *time.Time `json:"finished_at"`
	ErrorCode       string     `json:"error_code,omitempty"`
}

type Claim struct {
	ID             int64
	Token          string
	LeaseExpiresAt time.Time
	Prompt         string
	Config         Config
}

type Page struct {
	Items      []Run  `json:"items"`
	NextCursor string `json:"next_cursor"`
}
