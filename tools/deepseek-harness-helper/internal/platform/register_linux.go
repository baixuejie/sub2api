//go:build linux

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func registerProtocol(reg Registration) error {
	executable, err := filepath.Abs(reg.Executable)
	if err != nil {
		return err
	}
	if strings.ContainsAny(executable, "\r\n") {
		return fmt.Errorf("executable path contains an invalid control character")
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	applications := filepath.Join(dataHome, "applications")
	if err := os.MkdirAll(applications, 0o700); err != nil {
		return err
	}
	desktop := "[Desktop Entry]\nType=Application\nName=Sub2API DeepSeek Harness Helper\nNoDisplay=true\nTerminal=false\nExec=" + desktopQuote(executable) + " %u\nMimeType=x-scheme-handler/sub2api-harness;\n"
	filename := filepath.Join(applications, "deepseek-harness-helper.desktop")
	if err := os.WriteFile(filename, []byte(desktop), 0o600); err != nil {
		return err
	}
	xdgMime, err := exec.LookPath("xdg-mime")
	if err != nil {
		return fmt.Errorf("xdg-mime is required to register the protocol: %w", err)
	}
	cmd := exec.Command(xdgMime, "default", filepath.Base(filename), "x-scheme-handler/sub2api-harness")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("xdg-mime registration failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

func desktopQuote(value string) string {
	replacer := strings.NewReplacer(`\\`, `\\\\`, `"`, `\\"`, "`", "\\`", "$", "\\$", "%", "%%")
	return `"` + replacer.Replace(value) + `"`
}
