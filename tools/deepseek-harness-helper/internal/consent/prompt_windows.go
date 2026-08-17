//go:build windows

package consent

import (
	"context"
	"errors"
	"syscall"
	"unsafe"
)

var consentMessageBox = syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")

func ConfirmServer(_ context.Context, request TrustRequest) (bool, error) {
	title, err := syscall.UTF16PtrFromString("Trust Sub2API site")
	if err != nil {
		return false, err
	}
	message, err := syscall.UTF16PtrFromString(
		"A Sub2API site is requesting permission to run a local tool setup task on this computer.\r\n\r\n" +
			request.Origin + "\r\nTool: " + request.ExtensionID + "\r\n\r\nOnly choose Yes if you recognize and trust this exact site and tool.",
	)
	if err != nil {
		return false, err
	}
	result, _, callErr := consentMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(message)),
		uintptr(unsafe.Pointer(title)),
		0x00000004|0x00000030|0x00000100|0x00010000,
	)
	if result == 0 {
		if callErr == nil || errors.Is(callErr, syscall.Errno(0)) {
			return false, errors.New("MessageBoxW failed")
		}
		return false, callErr
	}
	return result == 6, nil
}
