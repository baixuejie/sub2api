//go:build windows

package dsh

import (
	"fmt"
	"path/filepath"
	"time"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

func processStartTime(pid int, nodePath, _ string) (time.Time, error) {
	if pid <= 0 {
		return time.Time{}, errProcessIdentityUnavailable
	}
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return time.Time{}, err
	}
	defer windows.CloseHandle(handle)

	buffer := make([]uint16, 32768)
	size := uint32(len(buffer))
	if err := windows.QueryFullProcessImageName(handle, 0, &buffer[0], &size); err != nil {
		return time.Time{}, err
	}
	imagePath := string(utf16.Decode(buffer[:size]))
	left, leftErr := filepath.Abs(imagePath)
	right, rightErr := filepath.Abs(nodePath)
	if leftErr != nil || rightErr != nil || !samePath(left, right) {
		return time.Time{}, fmt.Errorf("%w: executable path mismatch", errProcessIdentityUnavailable)
	}

	var creation, exit, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(handle, &creation, &exit, &kernel, &user); err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, creation.Nanoseconds()).UTC(), nil
}
