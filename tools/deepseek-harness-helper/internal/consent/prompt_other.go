//go:build !windows && !darwin && !linux

package consent

import (
	"context"
	"errors"
)

func ConfirmServer(context.Context, string) (bool, error) {
	return false, errors.New("server trust confirmation is not supported on this platform")
}
