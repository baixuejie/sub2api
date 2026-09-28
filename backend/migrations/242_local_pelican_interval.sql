-- Transactional, idempotent extension migration; existing schedules stay hourly.
ALTER TABLE pelican_config
    ADD COLUMN IF NOT EXISTS interval_minutes INTEGER NOT NULL DEFAULT 60
    CHECK (interval_minutes IN (10, 30, 60));
