package pelican

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const configColumns = `revision, enabled, api_key_encrypted, selected_group_ids, topic_mode,
 fixed_topic_id, rotation_sequence, max_output_tokens, timeout_seconds, retention_days, next_run_at`

type scanner interface{ Scan(...any) error }

func scanConfig(row scanner) (Config, error) {
	c := Config{Model: Model, ReasoningEffort: ReasoningEffort}
	var groups []byte
	err := row.Scan(&c.Revision, &c.Enabled, &c.EncryptedKey, &groups, &c.TopicMode,
		&c.FixedTopicID, &c.Sequence, &c.MaxOutputTokens, &c.TimeoutSeconds, &c.RetentionDays, &c.NextRunAt)
	if err != nil {
		return c, err
	}
	if err = json.Unmarshal(groups, &c.SelectedGroupIDs); err != nil {
		return c, err
	}
	c.KeyConfigured = c.EncryptedKey != ""
	if c.KeyConfigured {
		c.KeyMasked = "••••••••"
	}
	return c, nil
}

func (r *Repository) Config(ctx context.Context) (Config, error) {
	return scanConfig(r.db.QueryRowContext(ctx, `SELECT `+configColumns+` FROM pelican_config WHERE id=1`))
}

func (r *Repository) SaveConfig(ctx context.Context, in SaveConfig, encrypted *string, actorID int64, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	cfg, err := scanConfig(tx.QueryRowContext(ctx, `SELECT `+configColumns+` FROM pelican_config WHERE id=1 FOR UPDATE`))
	if err != nil {
		return err
	}
	if cfg.Revision != in.Revision {
		return ErrConflict
	}
	if encrypted != nil {
		cfg.EncryptedKey = *encrypted
	}
	if in.Enabled && cfg.EncryptedKey == "" {
		return ErrKeyRequired
	}
	groups, _ := json.Marshal(in.SelectedGroupIDs)
	var count int
	err = tx.QueryRowContext(ctx, `SELECT count(*) FROM groups WHERE deleted_at IS NULL AND status='active'
 AND id IN (SELECT value::bigint FROM jsonb_array_elements_text($1::jsonb))`, string(groups)).Scan(&count)
	if err != nil {
		return err
	}
	if count != len(in.SelectedGroupIDs) {
		return ErrGroups
	}
	var next *time.Time
	if in.Enabled {
		next = cfg.NextRunAt
		if !cfg.Enabled || next == nil {
			n := nextHour(now)
			next = &n
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE pelican_config SET revision=revision+1, enabled=$1,
 api_key_encrypted=$2, selected_group_ids=$3::jsonb, topic_mode=$4, fixed_topic_id=$5,
 max_output_tokens=$6, timeout_seconds=$7, retention_days=$8, next_run_at=$9, updated_by=NULLIF($10,0), updated_at=$11 WHERE id=1`,
		in.Enabled, cfg.EncryptedKey, string(groups), in.TopicMode, in.FixedTopicID,
		in.MaxOutputTokens, in.TimeoutSeconds, in.RetentionDays, next, actorID, now)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) ClearKey(ctx context.Context, revision, actorID int64) error {
	res, err := r.db.ExecContext(ctx, `UPDATE pelican_config SET api_key_encrypted='', enabled=FALSE,
 next_run_at=NULL, revision=revision+1, updated_by=NULLIF($2,0), updated_at=NOW() WHERE id=1 AND revision=$1`, revision, actorID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err == nil && n == 0 {
		return ErrConflict
	}
	return err
}

// Claim advances the schedule before dispatch. An ambiguous commit or lost worker
// is deliberately NOT retried: the next scan will see the claimed hour.
func (r *Repository) Claim(ctx context.Context, now time.Time) (*Claim, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	cfg, err := scanConfig(tx.QueryRowContext(ctx, `SELECT `+configColumns+` FROM pelican_config WHERE id=1 FOR UPDATE`))
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE pelican_runs SET status='interrupted', error_code='lease_expired', finished_at=$1
 WHERE config_id=1 AND status='running' AND lease_expires_at <= $1`, now)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled || cfg.NextRunAt == nil || cfg.NextRunAt.After(now) {
		return nil, tx.Commit()
	}
	slot := now.UTC().Truncate(time.Hour)
	missed := int64(slot.Sub(cfg.NextRunAt.UTC().Truncate(time.Hour)) / time.Hour)
	if missed < 0 {
		missed = 0
	}
	topic := selectTopic(cfg, missed)
	prompt := topic.prompt()
	var busy bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pelican_runs WHERE config_id=1 AND status='running')`).Scan(&busy); err != nil {
		return nil, err
	}
	ids, _ := json.Marshal(cfg.SelectedGroupIDs)
	rows, err := tx.QueryContext(ctx, `SELECT id, name FROM groups WHERE deleted_at IS NULL AND status='active'
 AND id IN (SELECT value::bigint FROM jsonb_array_elements_text($1::jsonb)) ORDER BY id`, string(ids))
	if err != nil {
		return nil, err
	}
	tags := []GroupTag{}
	for rows.Next() {
		var g GroupTag
		if err = rows.Scan(&g.ID, &g.Name); err != nil {
			rows.Close()
			return nil, err
		}
		tags = append(tags, g)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	tagsJSON, _ := json.Marshal(tags)
	snapshot, _ := json.Marshal(map[string]any{"model": Model, "reasoning": map[string]string{"effort": ReasoningEffort},
		"max_output_tokens": cfg.MaxOutputTokens, "timeout_seconds": cfg.TimeoutSeconds})
	var random [16]byte
	if _, err = rand.Read(random[:]); err != nil {
		return nil, err
	}
	claim := &Claim{Token: hex.EncodeToString(random[:]), LeaseExpiresAt: now.Add(time.Duration(cfg.TimeoutSeconds+60) * time.Second), Prompt: prompt, Config: cfg}
	status, code := "running", ""
	var finished *time.Time
	if busy {
		status, code, finished = "skipped", "previous_run_active", &now
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO pelican_runs
 (config_id, config_revision, scheduled_for, status, claim_token, lease_expires_at, topic_id, prompt_version,
 prompt_hash, prompt_snapshot, request_snapshot, selected_groups_snapshot, started_at, finished_at, error_code, skipped_hours)
 VALUES (1,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11::jsonb,$12,$13,$14,$15)
 ON CONFLICT (config_id, scheduled_for) DO NOTHING RETURNING id`, cfg.Revision, slot, status, claim.Token,
		claim.LeaseExpiresAt, topic.ID, topic.Version, contentHash(prompt), prompt, string(snapshot), string(tagsJSON), now, finished, code, missed).Scan(&claim.ID)
	duplicate := errors.Is(err, sql.ErrNoRows)
	if err != nil && !duplicate {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE pelican_config SET next_run_at=$1, rotation_sequence=rotation_sequence+$2 WHERE id=1`, nextHour(now), missed+1)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	if busy || duplicate {
		return nil, nil
	}
	return claim, nil
}

func (r *Repository) ClaimActive(ctx context.Context, c *Claim) (bool, error) {
	var active bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pelican_runs
 WHERE id=$1 AND claim_token=$2 AND status='running' AND lease_expires_at > NOW())`, c.ID, c.Token).Scan(&active)
	return active, err
}

// Finish never logs or persists the upstream error body, which can echo a key.
func (r *Repository) Finish(ctx context.Context, claim *Claim, result *Generation, preview, status, code string, latency time.Duration) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if result == nil {
		result = &Generation{}
	}
	res, err := tx.ExecContext(ctx, `UPDATE pelican_runs SET status=$3, error_code=$4, finished_at=NOW(), latency_ms=$5,
 response_model=$6, input_tokens=$7, output_tokens=$8, total_tokens=$9, usage_details=$10::jsonb, response_id=$11
 WHERE id=$1 AND claim_token=$2 AND status='running' AND lease_expires_at > NOW()`, claim.ID, claim.Token, status, code,
		latency.Milliseconds(), result.Model, result.InputTokens, result.OutputTokens, result.TotalTokens, nullableJSON(result.Usage), result.ResponseID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	if status == "succeeded" {
		_, err = tx.ExecContext(ctx, `INSERT INTO pelican_artifacts (run_id,raw_output,preview_html,content_sha256,size_bytes,preview_policy_version)
 VALUES ($1,$2,$3,$4,$5,$6)`, claim.ID, result.Text, preview, contentHash(result.Text), len(preview), PreviewPolicyVersion)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func nullableJSON(v json.RawMessage) any {
	if len(v) == 0 {
		return nil
	}
	return string(v)
}

type ListOptions struct {
	Admin   bool
	Limit   int
	Cursor  string
	Topic   string
	GroupID int64
}

type cursor struct {
	At time.Time `json:"at"`
	ID int64     `json:"id"`
}

func parseCursor(s string) (cursor, error) {
	var c cursor
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil || len(b) > 256 {
		return c, errors.New("invalid cursor")
	}
	if err = json.Unmarshal(b, &c); err != nil || c.ID <= 0 || c.At.IsZero() {
		return c, errors.New("invalid cursor")
	}
	return c, nil
}

const runColumns = `r.id, r.scheduled_for, r.status, r.topic_id, r.selected_groups_snapshot,
 r.input_tokens, r.output_tokens, r.total_tokens, r.latency_ms, r.finished_at, r.error_code`

const visibleRun = `r.status='succeeded' AND EXISTS (SELECT 1 FROM pelican_artifacts a WHERE a.run_id=r.id AND a.preview_policy_version=`

func visibleSQL() string { return visibleRun + strconv.Itoa(PreviewPolicyVersion) + `)` }

func scanRun(row scanner) (Run, error) {
	r := Run{Model: Model, ReasoningEffort: ReasoningEffort}
	var groups []byte
	err := row.Scan(&r.ID, &r.ScheduledFor, &r.Status, &r.TopicID, &groups,
		&r.InputTokens, &r.OutputTokens, &r.TotalTokens, &r.LatencyMS, &r.FinishedAt, &r.ErrorCode)
	if err == nil {
		err = json.Unmarshal(groups, &r.Groups)
	}
	return r, err
}

func (r *Repository) List(ctx context.Context, opt ListOptions) (Page, error) {
	page := Page{Items: []Run{}}
	if opt.Limit <= 0 || opt.Limit > 100 {
		opt.Limit = 30
	}
	where := []string{"TRUE"}
	args := []any{}
	add := func(v any) string { args = append(args, v); return "$" + strconv.Itoa(len(args)) }
	if !opt.Admin {
		where = append(where, visibleSQL())
	}
	if opt.Topic != "" {
		where = append(where, "r.topic_id="+add(opt.Topic))
	}
	if opt.GroupID > 0 {
		tag, _ := json.Marshal([]map[string]int64{{"id": opt.GroupID}})
		where = append(where, "r.selected_groups_snapshot @> "+add(string(tag))+"::jsonb")
	}
	if opt.Cursor != "" {
		c, err := parseCursor(opt.Cursor)
		if err != nil {
			return page, err
		}
		where = append(where, "(r.scheduled_for,r.id) < ("+add(c.At)+","+add(c.ID)+")")
	}
	limit := add(opt.Limit + 1)
	rows, err := r.db.QueryContext(ctx, `SELECT `+runColumns+` FROM pelican_runs r WHERE `+strings.Join(where, " AND ")+` ORDER BY r.scheduled_for DESC,r.id DESC LIMIT `+limit, args...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		item, err := scanRun(rows)
		if err != nil {
			return page, err
		}
		if !opt.Admin {
			item.ErrorCode = ""
		}
		page.Items = append(page.Items, item)
	}
	if err = rows.Err(); err != nil {
		return page, err
	}
	if len(page.Items) > opt.Limit {
		page.Items = page.Items[:opt.Limit]
		last := page.Items[len(page.Items)-1]
		b, _ := json.Marshal(cursor{At: last.ScheduledFor, ID: last.ID})
		page.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return page, nil
}

func (r *Repository) Artifact(ctx context.Context, id int64, source bool) (string, error) {
	column, filter := "a.preview_html", " AND r.status='succeeded' AND a.preview_policy_version="+strconv.Itoa(PreviewPolicyVersion)
	if source {
		column, filter = "a.raw_output", ""
	}
	var content string
	err := r.db.QueryRowContext(ctx, `SELECT `+column+` FROM pelican_artifacts a JOIN pelican_runs r ON r.id=a.run_id WHERE r.id=$1`+filter, id).Scan(&content)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return content, err
}

func (r *Repository) Groups(ctx context.Context) ([]GroupTag, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT DISTINCT (tag->>'id')::bigint, tag->>'name' FROM pelican_runs r,
 LATERAL jsonb_array_elements(r.selected_groups_snapshot) tag WHERE `+visibleSQL()+` ORDER BY 2,1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := []GroupTag{}
	for rows.Next() {
		var g GroupTag
		if err = rows.Scan(&g.ID, &g.Name); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *Repository) Status(ctx context.Context) (map[string]any, error) {
	var enabled bool
	var next, last *time.Time
	err := r.db.QueryRowContext(ctx, `SELECT enabled,next_run_at,(SELECT MAX(finished_at) FROM pelican_runs r WHERE `+visibleSQL()+`) FROM pelican_config WHERE id=1`).Scan(&enabled, &next, &last)
	return map[string]any{"enabled": enabled, "next_run_at": next, "last_success_at": last, "interval_seconds": 3600}, err
}

func (r *Repository) Detail(ctx context.Context, id int64) (map[string]any, error) {
	item, err := scanRun(r.db.QueryRowContext(ctx, `SELECT `+runColumns+` FROM pelican_runs r WHERE r.id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var prompt string
	var request, usage json.RawMessage
	var skipped int64
	err = r.db.QueryRowContext(ctx, `SELECT prompt_snapshot,request_snapshot,usage_details,skipped_hours FROM pelican_runs WHERE id=$1`, id).Scan(&prompt, &request, &usage, &skipped)
	return map[string]any{"run": item, "prompt": prompt, "request": request, "usage": usage, "skipped_hours": skipped}, err
}

func (r *Repository) Cleanup(ctx context.Context) error {
	// A bounded batch is enough at 24 runs/day and avoids long-held locks.
	_, err := r.db.ExecContext(ctx, `DELETE FROM pelican_runs WHERE id IN (SELECT r.id FROM pelican_runs r
 JOIN pelican_config c ON c.id=r.config_id WHERE r.status<>'running'
 AND r.scheduled_for < NOW() - make_interval(days => c.retention_days) ORDER BY r.id LIMIT 500)`)
	if err != nil {
		return fmt.Errorf("pelican retention: %w", err)
	}
	return nil
}
