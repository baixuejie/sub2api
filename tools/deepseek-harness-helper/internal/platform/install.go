package platform

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func InstallSelf(executable, dataDir string) (string, error) {
	source, err := filepath.Abs(executable)
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		marker := ".app" + string(filepath.Separator) + "Contents" + string(filepath.Separator) + "MacOS" + string(filepath.Separator)
		if !strings.Contains(source, marker) {
			return "", errors.New("macOS installation requires the packaged DeepSeek Harness Helper.app")
		}
		// LaunchServices registers CFBundleURLTypes when the packaged app is opened.
		return source, nil
	}
	if strings.TrimSpace(dataDir) == "" {
		return "", errors.New("helper data directory is required")
	}
	binDir := filepath.Join(dataDir, "bin")
	if err := os.MkdirAll(binDir, 0o700); err != nil {
		return "", err
	}
	name := "deepseek-harness-helper"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	target := filepath.Join(binDir, name)
	if !sameInstallPath(source, target) {
		if err := copyExecutableAtomic(source, target); err != nil {
			return "", err
		}
	}
	if err := RegisterProtocol(target); err != nil {
		return "", err
	}
	return target, nil
}

func copyExecutableAtomic(source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	temp, err := os.CreateTemp(filepath.Dir(target), ".helper-install-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	committed := false
	defer func() {
		_ = temp.Close()
		if !committed {
			_ = os.Remove(tempName)
		}
	}()
	if err = temp.Chmod(0o700); err != nil {
		return err
	}
	if _, err = io.Copy(temp, input); err != nil {
		return err
	}
	if err = temp.Sync(); err != nil {
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(target)
	}
	if err = os.Rename(tempName, target); err != nil {
		return fmt.Errorf("install helper executable: %w", err)
	}
	committed = true
	return os.Chmod(target, 0o700)
}

func sameInstallPath(left, right string) bool {
	left, leftErr := filepath.Abs(left)
	right, rightErr := filepath.Abs(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
