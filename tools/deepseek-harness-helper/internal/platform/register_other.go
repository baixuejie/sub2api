//go:build !windows && !linux && !darwin

package platform

import "errors"

func registerProtocol(Registration) error {
	return errors.New("protocol registration is not supported on this platform")
}
