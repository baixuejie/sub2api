//go:build windows

package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

func tightenPlatformPermissions(paths ...string) error {
	current, err := user.Current()
	if err != nil || strings.TrimSpace(current.Username) == "" {
		return errors.New("resolve current Windows user for private ACL")
	}
	icacls, err := exec.LookPath("icacls.exe")
	if err != nil {
		return errors.New("icacls.exe is required to secure Helper credentials")
	}
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if err := runICACLS(icacls,
			path,
			"/inheritance:r",
			"/grant:r", current.Username+":F",
			"/grant:r", "*S-1-5-18:F",
			"/Q",
		); err != nil {
			return err
		}
		if info.IsDir() {
			if err := runICACLS(icacls,
				path,
				"/grant", current.Username+":(OI)(CI)F",
				"/grant", "*S-1-5-18:(OI)(CI)F",
				"/Q",
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func runICACLS(executable string, args ...string) error {
	output, err := exec.Command(executable, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("secure Helper data ACL: %w (%s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}
