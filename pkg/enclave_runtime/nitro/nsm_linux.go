//go:build linux

package nitro

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const nsmIoctlRaw = 0xC0200A00

type nsmIovec struct {
	Addr uint64
	Len  uint64
}

type nsmRaw struct {
	Request  nsmIovec
	Response nsmIovec
}

// nsmIoctl performs an ioctl call to the NSM device using the Linux NSM ABI.
func nsmIoctl(fd *os.File, request, response []byte) (int, error) {
	if fd == nil {
		return 0, ErrNSMDeviceNotOpen
	}
	if len(request) == 0 {
		return 0, ErrNSMInvalidArgument
	}
	if len(request) > 0x1000 {
		return 0, fmt.Errorf("%w: request exceeds NSM maximum size", ErrNSMInputTooLarge)
	}
	if len(response) == 0 {
		return 0, ErrNSMBufferTooSmall
	}

	raw := nsmRaw{
		Request: nsmIovec{
			Addr: uint64(uintptr(unsafe.Pointer(&request[0]))),
			Len:  uint64(len(request)),
		},
		Response: nsmIovec{
			Addr: uint64(uintptr(unsafe.Pointer(&response[0]))),
			Len:  uint64(len(response)),
		},
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd.Fd(),
		uintptr(nsmIoctlRaw),
		uintptr(unsafe.Pointer(&raw)),
	)
	if errno != 0 {
		return 0, errno
	}
	return int(raw.Response.Len), nil
}
