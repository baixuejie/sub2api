//go:build !windows

package config

import (
	"os"
)

func replaceFile(temp, target string) error {
	return os.Rename(temp, target)
}

func syncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
