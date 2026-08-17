//go:build !windows

package dsh

import (
	"errors"
	"os"
	"syscall"
	"time"
)

type startLock struct {
	file *os.File
}

func acquireStartLock(path string, timeout time.Duration) (*startLock, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(timeout)
	for {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return &startLock{file: file}, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			file.Close()
			return nil, err
		}
		if time.Now().After(deadline) {
			file.Close()
			return nil, errors.New("timeout acquiring DSH startup lock")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (lock *startLock) release() {
	if lock == nil || lock.file == nil {
		return
	}
	_ = syscall.Flock(int(lock.file.Fd()), syscall.LOCK_UN)
	_ = lock.file.Close()
}
