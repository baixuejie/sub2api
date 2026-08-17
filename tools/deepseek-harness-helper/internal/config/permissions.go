package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func EnsurePrivateDir(path string) error {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(path, 0o700); err != nil {
		return fmt.Errorf("tighten directory permissions: %w", err)
	}
	return nil
}

func TightenPrivateTree(paths ...string) error {
	for _, path := range paths {
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		mode := os.FileMode(0o600)
		if info.IsDir() {
			mode = 0o700
		}
		if err := os.Chmod(path, mode); err != nil {
			return fmt.Errorf("tighten permissions on %s: %w", filepath.Base(path), err)
		}
	}
	return tightenPlatformPermissions(paths...)
}
