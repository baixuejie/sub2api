//go:build windows

package consent

import (
	"errors"
	"syscall"
	"unsafe"
)

var consentMoveFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("MoveFileExW")

func replaceFile(temp, target string) error {
	from, err := syscall.UTF16PtrFromString(temp)
	if err != nil {
		return err
	}
	to, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	result, _, callErr := consentMoveFileEx.Call(
		uintptr(unsafe.Pointer(from)),
		uintptr(unsafe.Pointer(to)),
		0x1|0x8,
	)
	if result != 0 {
		return nil
	}
	if callErr == nil || errors.Is(callErr, syscall.Errno(0)) {
		return errors.New("MoveFileExW failed")
	}
	return callErr
}
