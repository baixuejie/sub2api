//go:build windows

package platform

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

func registerProtocol(reg Registration) error {
	executable, err := filepath.Abs(reg.Executable)
	if err != nil {
		return err
	}
	root := `Software\Classes\sub2api-harness`
	key, _, err := registry.CreateKey(registry.CURRENT_USER, root, registry.SET_VALUE)
	if err != nil {
		return err
	}
	if err := key.SetStringValue("", "URL:Sub2API DeepSeek Harness Bootstrap"); err != nil {
		key.Close()
		return err
	}
	if err := key.SetStringValue("URL Protocol", ""); err != nil {
		key.Close()
		return err
	}
	key.Close()
	command, _, err := registry.CreateKey(registry.CURRENT_USER, root+`\shell\open\command`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer command.Close()
	return command.SetStringValue("", fmt.Sprintf(`"%s" "%%1"`, executable))
}
