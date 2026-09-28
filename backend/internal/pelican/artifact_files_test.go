package pelican

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPelicanCompatiblePreview(t *testing.T) {
	raw := `<!DOCTYPE html><html><head><meta charset="utf-8"><style>.bird{fill:url("#sky")}</style></head><body><custom-wrapper><svg version="1.1" data-scene="bird" xml:space="preserve" viewBox="0 0 960 720"><defs><linearGradient id="sky"><stop offset="0" stop-color="blue"/></linearGradient><marker id="arrow" markerWidth="4" markerHeight="4"><path d="M0 0L4 4"/></marker></defs><foreignObject width="100" height="40"><div>Caption</div></foreignObject><circle class="bird" r="4" unknown-decoration="x"/></svg></custom-wrapper></body></html>`
	preview, notes, err := sanitizePreview(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"foreignObject", "Caption", "linearGradient", "version=", "data-scene=", "marker", `.bird{fill:url("#sky")}`} {
		if !strings.Contains(preview, want) {
			t.Errorf("lost %s: %s", want, preview)
		}
	}
	if strings.Contains(preview, "unknown-decoration") || len(notes) == 0 {
		t.Fatal("unknown attribute not reported/removed")
	}
	if _, err := preparePreview(`<html><body><div style="width:40px;height:40px;background:red"></div></body></html>`); err != nil {
		t.Fatal("CSS drawing rejected", err)
	}
	if _, _, err := sanitizePreview("plain unsupported output"); err == nil || !strings.Contains(err.Error(), "not_html_or_svg") {
		t.Fatal("missing actionable rejection reason")
	}
}

func TestPelicanFilesIntegrityTraversalAndDeletion(t *testing.T) {
	files := &artifactFiles{dir: t.TempDir()}
	path, err := files.write(7, "source", safeExample)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := files.read(path); err != nil || got != safeExample {
		t.Fatalf("roundtrip %v", err)
	}
	info, err := os.Stat(filepath.Join(files.dir, path))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("unsafe file permissions", err)
	}
	for _, path := range []string{"../secret", "/etc/passwd", "7/../../secret"} {
		if _, err = files.read(path); err == nil {
			t.Fatal("allowed traversal")
		}
	}
	if err = os.WriteFile(filepath.Join(files.dir, path), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = files.read(path); err == nil {
		t.Fatal("ignored checksum mismatch")
	}
	if err = files.removeRun(7); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(files.dir, "7")); !os.IsNotExist(err) {
		t.Fatal("retention left files")
	}
	outside := t.TempDir()
	if err = os.Symlink(outside, filepath.Join(files.dir, "8")); err != nil {
		t.Fatal(err)
	}
	if _, err = files.write(8, "source", safeExample); err == nil {
		t.Fatal("followed escaping directory symlink")
	}
}
