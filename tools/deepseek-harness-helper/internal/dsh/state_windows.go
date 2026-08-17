//go:build windows

package dsh

import (
	"os"
)

func replaceState(temp, target string) error {
	_ = os.Remove(target)
	return os.Rename(temp, target)
}
