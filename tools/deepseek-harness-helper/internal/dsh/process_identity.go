package dsh

import "errors"

func processIdentityMatches(state ProcessState) bool {
	startedAt, err := processStartTime(state.PID, state.NodePath, state.DSHBin)
	return err == nil && startedAt.Equal(state.StartedAt)
}

var errProcessIdentityUnavailable = errors.New("managed DSH process identity is unavailable")
