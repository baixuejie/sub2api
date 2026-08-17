//go:build windows

package platform

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

var (
	kernel32Console = windows.NewLazySystemDLL("kernel32.dll")
	allocConsole    = kernel32Console.NewProc("AllocConsole")
)

// EnsureInteractiveConsole binds the process and its children to real console
// handles. Protocol handlers can otherwise inherit pipes or invalid handles,
// which NVM for Windows explicitly rejects.
func EnsureInteractiveConsole() error {
	stdin, stdout, stderr, err := openConsoleFiles()
	if err != nil {
		if allocErr := callAllocConsole(); allocErr != nil && !errors.Is(allocErr, windows.ERROR_ACCESS_DENIED) {
			return fmt.Errorf("allocate Windows console: %w", allocErr)
		}
		stdin, stdout, stderr, err = openConsoleFiles()
		if err != nil {
			return err
		}
	}

	if err := windows.SetStdHandle(windows.STD_INPUT_HANDLE, windows.Handle(stdin.Fd())); err != nil {
		closeConsoleFiles(stdin, stdout, stderr)
		return fmt.Errorf("bind Windows console input: %w", err)
	}
	if err := windows.SetStdHandle(windows.STD_OUTPUT_HANDLE, windows.Handle(stdout.Fd())); err != nil {
		closeConsoleFiles(stdin, stdout, stderr)
		return fmt.Errorf("bind Windows console output: %w", err)
	}
	if err := windows.SetStdHandle(windows.STD_ERROR_HANDLE, windows.Handle(stderr.Fd())); err != nil {
		closeConsoleFiles(stdin, stdout, stderr)
		return fmt.Errorf("bind Windows console error output: %w", err)
	}

	os.Stdin = stdin
	os.Stdout = stdout
	os.Stderr = stderr
	return nil
}

func callAllocConsole() error {
	result, _, callErr := allocConsole.Call()
	if result != 0 {
		return nil
	}
	if callErr != nil && !errors.Is(callErr, windows.ERROR_SUCCESS) {
		return callErr
	}
	return errors.New("AllocConsole returned false without a Windows error")
}

func openConsoleFiles() (*os.File, *os.File, *os.File, error) {
	stdin, err := os.OpenFile("CONIN$", os.O_RDONLY, 0)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open Windows console input: %w", err)
	}
	stdout, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		_ = stdin.Close()
		return nil, nil, nil, fmt.Errorf("open Windows console output: %w", err)
	}
	stderr, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, nil, nil, fmt.Errorf("open Windows console error output: %w", err)
	}
	return stdin, stdout, stderr, nil
}

func closeConsoleFiles(files ...*os.File) {
	for _, file := range files {
		if file != nil {
			_ = file.Close()
		}
	}
}

func WaitForExitPrompt() {
	var mode uint32
	if err := windows.GetConsoleMode(windows.Handle(os.Stdin.Fd()), &mode); err != nil {
		return
	}
	_, _ = fmt.Fprint(os.Stdout, "\nPress Enter to close this window...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}
