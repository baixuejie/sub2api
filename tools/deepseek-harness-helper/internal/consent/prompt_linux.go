//go:build linux

package consent

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

func ConfirmServer(ctx context.Context, origin string) (bool, error) {
	message := "A Sub2API site is requesting permission to install and configure DeepSeek Harness on this computer.\n\n" + origin + "\n\nOnly approve if you recognize and trust this exact site."
	if zenity, err := exec.LookPath("zenity"); err == nil {
		return runQuestion(exec.CommandContext(ctx, zenity, "--question", "--title=Trust Sub2API site", "--text="+message, "--ok-label=Trust", "--cancel-label=Cancel"))
	}
	if kdialog, err := exec.LookPath("kdialog"); err == nil {
		return runQuestion(exec.CommandContext(ctx, kdialog, "--warningyesno", message, "--title", "Trust Sub2API site", "--yes-label", "Trust", "--no-label", "Cancel"))
	}
	if xmessage, err := exec.LookPath("xmessage"); err == nil {
		return runQuestion(exec.CommandContext(ctx, xmessage, "-center", "-title", "Trust Sub2API site", "-buttons", "Trust:0,Cancel:1", "-default", "Cancel", message))
	}
	return false, errors.New("zenity, kdialog, or xmessage is required to confirm server trust")
}

func runQuestion(command *exec.Cmd) (bool, error) {
	err := command.Run()
	if err == nil {
		return true, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return false, fmt.Errorf("show server trust confirmation: %w", err)
}
