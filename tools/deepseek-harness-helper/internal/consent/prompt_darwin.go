//go:build darwin

package consent

import (
	"context"
	"errors"
	"os/exec"
	"strings"
)

const trustAppleScript = `on run argv
  set serverOrigin to item 1 of argv
  try
    set answer to display dialog "A Sub2API site is requesting permission to run a local tool setup task on this Mac." & return & return & serverOrigin & return & return & "Only choose Trust if you recognize this exact site." buttons {"Cancel", "Trust"} default button "Cancel" with icon caution
    if button returned of answer is "Trust" then return "approved"
  on error number -128
    return "declined"
  end try
  return "declined"
end run`

func ConfirmServer(ctx context.Context, request TrustRequest) (bool, error) {
	osascript, err := exec.LookPath("osascript")
	if err != nil {
		return false, errors.New("osascript is required to confirm server trust")
	}
	description := request.Origin + "\nTool: " + request.ExtensionID
	output, err := exec.CommandContext(ctx, osascript, "-e", trustAppleScript, "--", description).Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(output)) == "approved", nil
}
