//go:build !windows

package platform

func EnsureInteractiveConsole() error { return nil }

func WaitForExitPrompt() {}
