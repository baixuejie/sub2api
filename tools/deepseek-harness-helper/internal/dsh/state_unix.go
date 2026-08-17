//go:build !windows

package dsh

import "os"

func replaceState(temp, target string) error { return os.Rename(temp, target) }
