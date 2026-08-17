//go:build windows

package config

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32   = syscall.NewLazyDLL("kernel32.dll")
	moveFileEx = kernel32.NewProc("MoveFileExW")
)

const (
	moveFileReplaceExisting = 0x1
	moveFileWriteThrough    = 0x8
)

func replaceFile(temp, target string) error {
	from, err := syscall.UTF16PtrFromString(temp)
	if err != nil {
		return err
	}
	to, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	r1, _, callErr := moveFileEx.Call(
		uintptr(unsafe.Pointer(from)),
		uintptr(unsafe.Pointer(to)),
		moveFileReplaceExisting|moveFileWriteThrough,
	)
	if r1 == 0 {
		if callErr == nil || errors.Is(callErr, syscall.Errno(0)) {
			return errors.New("MoveFileExW failed")
		}
		return callErr
	}
	return nil
}

func syncDirectory(string) error { return nil }

var _ = os.ErrNotExist
