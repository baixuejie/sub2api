package pelican

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// Restartable: clear database text only after atomic, verified file writes.
func (r *Repository) MigrateArtifacts(ctx context.Context) error {
	if err := r.CheckStorage(); err != nil {
		return err
	}
	root, err := r.files.open()
	if err != nil {
		return err
	}
	root.Close()
	for {
		var id int64
		err := r.db.QueryRowContext(ctx, `SELECT run_id FROM pelican_artifacts
   WHERE raw_output IS NOT NULL OR preview_html IS NOT NULL OR preview_policy_version<>$1 ORDER BY run_id LIMIT 1`, PreviewPolicyVersion).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if err = r.RebuildPreview(ctx, id); err != nil {
			return fmt.Errorf("migrate pelican run %d: %w", id, err)
		}
	}
}

func (r *Repository) CheckStorage() error { return r.files.checkWritable() }

// RebuildPreview reuses saved source; it never calls the model.
func (r *Repository) RebuildPreview(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	err = tx.QueryRowContext(ctx, `SELECT status FROM pelican_runs WHERE id=$1 FOR UPDATE`, id).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "succeeded" && status != "preview_blocked" {
		return ErrNotFound
	}
	var rawPath string
	var legacy sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT raw_path,raw_output FROM pelican_artifacts WHERE run_id=$1 FOR UPDATE`, id).Scan(&rawPath, &legacy)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	raw := legacy.String
	if rawPath != "" {
		raw, err = r.files.read(rawPath)
		if err != nil {
			return err
		}
	}
	preview, notes, previewErr := sanitizePreview(raw)
	status = "succeeded"
	code := ""
	if previewErr != nil {
		status, code = "preview_blocked", "preview_unavailable"
		notes = []string{previewErr.Error()}
	}
	rawPath, err = r.files.write(id, "source", raw)
	if err != nil {
		return err
	}
	previewPath := ""
	if preview != "" {
		previewPath, err = r.files.write(id, "preview", preview)
		if err != nil {
			return err
		}
	}
	if notes == nil {
		notes = []string{}
	}
	encoded, _ := json.Marshal(notes)
	_, err = tx.ExecContext(ctx, `UPDATE pelican_artifacts SET raw_path=$2,preview_path=$3,raw_output=NULL,preview_html=NULL,
   content_sha256=$4,size_bytes=$5,preview_policy_version=$6,preview_notes=$7::jsonb WHERE run_id=$1`, id, rawPath, previewPath, contentHash(raw), len(preview), PreviewPolicyVersion, string(encoded))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE pelican_runs SET status=$2,error_code=$3 WHERE id=$1`, id, status, code)
	if err != nil {
		return err
	}
	return tx.Commit()
}
