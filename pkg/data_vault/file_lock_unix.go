//go:build !windows

package data_vault

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func tryLockFile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600) // #nosec G304 -- path is the fixture store's configured state file supplied by this function's own caller, not remote input; opening it is the purpose of this call
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil { // #nosec G115 -- file.Fd() is an OS file descriptor: a small non-negative index far below 2^31, and unix.Flock takes an int
		_ = file.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, errFixtureStoreInUse
		}
		return nil, err
	}
	return file, nil
}

func unlockFile(file *os.File) error {
	if file == nil {
		return nil
	}
	unlockErr := unix.Flock(int(file.Fd()), unix.LOCK_UN) // #nosec G115 -- file.Fd() is an OS file descriptor: a small non-negative index far below 2^31, and unix.Flock takes an int
	closeErr := file.Close()
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}
