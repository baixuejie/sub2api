-- Local, opt-in hourly HTML/SVG gallery. No changes to upstream monitor tables.
CREATE TABLE IF NOT EXISTS pelican_config (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    revision BIGINT NOT NULL DEFAULT 1,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    api_key_encrypted TEXT NOT NULL DEFAULT '',
    selected_group_ids JSONB NOT NULL DEFAULT '[]',
    topic_mode TEXT NOT NULL DEFAULT 'rotate' CHECK (topic_mode IN ('rotate', 'fixed')),
    fixed_topic_id TEXT NOT NULL DEFAULT 'pelican-ski',
    rotation_sequence BIGINT NOT NULL DEFAULT 0,
    max_output_tokens INTEGER NOT NULL DEFAULT 16384 CHECK (max_output_tokens BETWEEN 4096 AND 32768),
    timeout_seconds INTEGER NOT NULL DEFAULT 300 CHECK (timeout_seconds BETWEEN 60 AND 600),
    retention_days INTEGER NOT NULL DEFAULT 30 CHECK (retention_days BETWEEN 7 AND 90),
    next_run_at TIMESTAMPTZ,
    updated_by BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
INSERT INTO pelican_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS pelican_runs (
    id BIGSERIAL PRIMARY KEY,
    config_id SMALLINT NOT NULL REFERENCES pelican_config(id),
    config_revision BIGINT NOT NULL,
    scheduled_for TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('running','succeeded','failed','interrupted','invalid_output','preview_blocked','skipped')),
    claim_token TEXT NOT NULL,
    lease_expires_at TIMESTAMPTZ NOT NULL,
    topic_id TEXT NOT NULL,
    prompt_version TEXT NOT NULL,
    prompt_hash TEXT NOT NULL,
    prompt_snapshot TEXT NOT NULL,
    request_snapshot JSONB NOT NULL,
    selected_groups_snapshot JSONB NOT NULL DEFAULT '[]',
    response_model TEXT NOT NULL DEFAULT '',
    input_tokens BIGINT,
    output_tokens BIGINT,
    total_tokens BIGINT,
    usage_details JSONB,
    response_id TEXT NOT NULL DEFAULT '',
    error_code TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    latency_ms BIGINT,
    skipped_hours BIGINT NOT NULL DEFAULT 0,
    UNIQUE (config_id, scheduled_for)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pelican_one_running ON pelican_runs(config_id) WHERE status = 'running';
CREATE INDEX IF NOT EXISTS idx_pelican_success_time ON pelican_runs(scheduled_for DESC, id DESC) WHERE status = 'succeeded';
CREATE INDEX IF NOT EXISTS idx_pelican_topic_time ON pelican_runs(topic_id, scheduled_for DESC, id DESC) WHERE status = 'succeeded';
CREATE INDEX IF NOT EXISTS idx_pelican_group_tags ON pelican_runs USING GIN(selected_groups_snapshot jsonb_path_ops) WHERE status = 'succeeded';

CREATE TABLE IF NOT EXISTS pelican_artifacts (
    run_id BIGINT PRIMARY KEY REFERENCES pelican_runs(id) ON DELETE CASCADE,
    raw_output TEXT NOT NULL CHECK (octet_length(raw_output) <= 1048576),
    preview_html TEXT NOT NULL CHECK (octet_length(preview_html) <= 1048576),
    content_sha256 TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    preview_policy_version INTEGER NOT NULL
);
