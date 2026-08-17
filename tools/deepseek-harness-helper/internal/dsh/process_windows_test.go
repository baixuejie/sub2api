//go:build windows

package dsh

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSamePathResolvesDirectorySymlink(t *testing.T) {
	root := t.TempDir()
	targetDir := filepath.Join(root, "node-version")
	if err := os.Mkdir(targetDir, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(targetDir, "node.exe")
	if err := os.WriteFile(target, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	linkDir := filepath.Join(root, "node-current")
	if err := os.Symlink(targetDir, linkDir); err != nil {
		t.Skipf("directory symlinks are unavailable: %v", err)
	}
	if !samePath(filepath.Join(linkDir, "node.exe"), target) {
		t.Fatal("symlinked and resolved executable paths should identify the same file")
	}
}
