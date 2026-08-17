//go:build !windows

package consent

import "os"

func replaceFile(temp, target string) error {
	return os.Rename(temp, target)
}
