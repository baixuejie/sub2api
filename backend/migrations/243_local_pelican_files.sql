-- Transactional, idempotent. Legacy text is retained until the application has
-- atomically written and verified the files, then cleared by MigrateArtifacts.
ALTER TABLE pelican_artifacts ADD COLUMN IF NOT EXISTS raw_path TEXT NOT NULL DEFAULT '';
ALTER TABLE pelican_artifacts ADD COLUMN IF NOT EXISTS preview_path TEXT NOT NULL DEFAULT '';
ALTER TABLE pelican_artifacts ADD COLUMN IF NOT EXISTS preview_notes JSONB NOT NULL DEFAULT '[]';
ALTER TABLE pelican_artifacts ALTER COLUMN raw_output DROP NOT NULL;
ALTER TABLE pelican_artifacts ALTER COLUMN preview_html DROP NOT NULL;
