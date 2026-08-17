//go:build darwin

package platform

import (
	"errors"
	"path/filepath"
	"strings"
)

func registerProtocol(reg Registration) error {
	path, err := filepath.Abs(reg.Executable)
	if err != nil {
		return err
	}
	marker := ".app" + string(filepath.Separator) + "Contents" + string(filepath.Separator) + "MacOS" + string(filepath.Separator)
	if !strings.Contains(path, marker) {
		return errors.New("macOS protocol registration requires the packaged .app path; build packaging/macos/DeepSeek Harness Helper.app and launch it once")
	}
	return errors.New("macOS LaunchServices registration is performed by launching the packaged .app once; run: open 'DeepSeek Harness Helper.app'")
}
