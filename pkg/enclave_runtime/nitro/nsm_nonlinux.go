//go:build !linux

package nitro

import (
	"errors"
	"os"
)

// nsmIoctl performs an ioctl call to the NSM device.
func nsmIoctl(_ *os.File, _ []byte, _ []byte) (int, error) {
	return 0, errors.New("real NSM ioctl requires Linux")
}
