//go:build !windows

package dsh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func processStartTime(pid int, nodePath, dshBin string) (time.Time, error) {
	if pid <= 0 {
		return time.Time{}, errProcessIdentityUnavailable
	}
	environment := append(os.Environ(), "LC_ALL=C", "TZ=UTC")
	command := exec.Command("ps", "-ww", "-p", strconv.Itoa(pid), "-o", "command=")
	command.Env = environment
	commandOutput, err := command.Output()
	if err != nil {
		return time.Time{}, err
	}
	commandLine := strings.TrimSpace(string(commandOutput))
	if !strings.Contains(commandLine, filepath.Base(nodePath)) || !strings.Contains(commandLine, filepath.Base(dshBin)) {
		return time.Time{}, fmt.Errorf("%w: command line mismatch", errProcessIdentityUnavailable)
	}

	started := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart=")
	started.Env = environment
	startedOutput, err := started.Output()
	if err != nil {
		return time.Time{}, err
	}
	value := strings.Join(strings.Fields(string(startedOutput)), " ")
	parsed, err := time.Parse("Mon Jan 2 15:04:05 2006", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: parse process start time: %v", errProcessIdentityUnavailable, err)
	}
	return parsed.UTC(), nil
}
