//go:build !windows

package provider_daemon

import (
	"os"

	"golang.org/x/sys/unix"
)

func lockQueueStateFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB) // #nosec G115 -- file.Fd() is an OS file descriptor: a small non-negative index far below 2^31, and unix.Flock takes an int
}

func unlockQueueStateFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_UN) // #nosec G115 -- file.Fd() is an OS file descriptor: a small non-negative index far below 2^31, and unix.Flock takes an int
}
