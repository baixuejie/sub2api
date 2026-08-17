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
