package dsh

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type ProcessState struct {
	PID             int       `json:"pid"`
	StartedAt       time.Time `json:"started_at"`
	NodePath        string    `json:"node_path"`
	DSHBin          string    `json:"dsh_bin"`
	DSHVersion      string    `json:"dsh_version"`
	ConfigurationID string    `json:"configuration_id"`
	URL             string    `json:"url"`
}

func readState(filename string) (ProcessState, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ProcessState{}, err
	}
	var state ProcessState
	if err := json.Unmarshal(data, &state); err != nil {
		return ProcessState{}, err
	}
	if state.PID <= 0 || state.StartedAt.IsZero() || state.NodePath == "" || state.DSHBin == "" || state.DSHVersion == "" || state.ConfigurationID == "" || state.URL == "" {
		return ProcessState{}, errors.New("invalid process state")
	}
	return state, nil
}

func removeStateForPID(filename string, pid int) {
	state, err := readState(filename)
	if err == nil && state.PID == pid {
		_ = os.Remove(filename)
	}
}

func writeState(filename string, state ProcessState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(filename), ".process-*.tmp")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(append(data, '\n')); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return replaceState(name, filename)
}
