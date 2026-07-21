//go:build !windows

package provider_daemon

import (
	"os"

	"golang.org/x/sys/unix"
)

func lockQueueStateFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
}

func unlockQueueStateFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_UN)
}
