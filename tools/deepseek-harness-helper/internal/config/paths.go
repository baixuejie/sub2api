package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

type Paths struct {
	DataDir          string
	InstallDir       string
	DSHHome          string
	SettingsFile     string
	CredentialsFile  string
	StateFile        string
	LogFile          string
	TrustedSitesFile string
}

func ResolvePaths() (Paths, error) {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		return Paths{}, errors.New("resolve user configuration directory")
	}
	if runtime.GOOS == "windows" {
		base = filepath.Join(base, "Sub2API", "DeepSeekHarnessHelper")
	} else {
		base = filepath.Join(base, "sub2api", "deepseek-harness-helper")
	}
	return PathsFor(base), nil
}

func PathsFor(base string) Paths {
	install := filepath.Join(base, "runtime")
	home := filepath.Join(base, "dsh-home")
	return Paths{
		DataDir:          base,
		InstallDir:       install,
		DSHHome:          home,
		SettingsFile:     filepath.Join(home, "settings.yaml"),
		CredentialsFile:  filepath.Join(home, ".credentials.yaml"),
		StateFile:        filepath.Join(base, "process.json"),
		LogFile:          filepath.Join(base, "dsh.log"),
		TrustedSitesFile: filepath.Join(base, "trusted-sites.json"),
	}
}

// ToolDataDir gives each adapter a stable private namespace without adding
// tool-specific fields to the shared workflow contract.
func (p Paths) ToolDataDir(toolID string) (string, error) {
	if p.DataDir == "" || !validToolID(toolID) {
		return "", errors.New("resolve tool data directory")
	}
	return filepath.Join(p.DataDir, "tools", toolID), nil
}

func validToolID(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			continue
		}
		return false
	}
	return true
}
