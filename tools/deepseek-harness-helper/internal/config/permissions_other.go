//go:build !windows

package config

func tightenPlatformPermissions(_ ...string) error {
	return nil
}
