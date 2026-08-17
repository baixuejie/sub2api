package config

import (
	"path/filepath"
	"testing"
)

func TestToolDataDirUsesValidatedAdapterNamespace(t *testing.T) {
	paths := PathsFor(t.TempDir())
	directory, err := paths.ToolDataDir("hermes")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(paths.DataDir, "tools", "hermes"); directory != want {
		t.Fatalf("directory = %q, want %q", directory, want)
	}
	for _, invalid := range []string{"", "../escape", "Hermes", "openclaw/shell"} {
		if _, err := paths.ToolDataDir(invalid); err == nil {
			t.Fatalf("ToolDataDir(%q) accepted an invalid ID", invalid)
		}
	}
}
