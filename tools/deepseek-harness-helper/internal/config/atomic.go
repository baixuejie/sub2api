package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const lockTimeout = 2 * time.Second

func withFileLock(filename string, operation func() error) error {
	if err := EnsurePrivateDir(filepath.Dir(filename)); err != nil {
		return err
	}
	lockPath := filename + ".lock"
	deadline := time.Now().Add(lockTimeout)
	delay := 20 * time.Millisecond
	var lock *os.File
	for {
		var err error
		lock, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			break
		}
		if !os.IsExist(err) {
			return fmt.Errorf("acquire file lock: %w", err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("acquire file lock: timed out waiting for %s", lockPath)
		}
		time.Sleep(delay)
		if delay < 200*time.Millisecond {
			delay *= 2
			if delay > 200*time.Millisecond {
				delay = 200 * time.Millisecond
			}
		}
	}
	_, writeErr := fmt.Fprintf(lock, "%d\n", os.Getpid())
	closeErr := lock.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		_ = os.Remove(lockPath)
		return fmt.Errorf("initialize file lock: %w", err)
	}
	operationErr := operation()
	removeErr := os.Remove(lockPath)
	return errors.Join(operationErr, removeErr)
}

func backupExisting(filename string) error {
	content, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return writeAtomic(filename+".bak", string(content))
}

func writeAtomic(filename, content string) error {
	dir := filepath.Dir(filename)
	if err := EnsurePrivateDir(dir); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, "."+filepath.Base(filename)+".tmp-")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	committed := false
	defer func() {
		_ = temp.Close()
		if !committed {
			_ = os.Remove(tempName)
		}
	}()
	if err := temp.Chmod(0o600); err != nil {
		return err
	}
	if _, err := temp.WriteString(content); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := replaceFile(tempName, filename); err != nil {
		return err
	}
	committed = true
	if err := os.Chmod(filename, 0o600); err != nil {
		return err
	}
	return syncDirectory(dir)
}
