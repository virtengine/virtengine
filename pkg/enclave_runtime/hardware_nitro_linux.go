//go:build linux

package enclave_runtime

import (
	"context"
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

func dialNitroVsock(ctx context.Context, cid, port uint32) (io.ReadWriteCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM, 0)
	if err != nil {
		return nil, fmt.Errorf("create vsock socket: %w", err)
	}

	addr := &unix.SockaddrVM{
		CID:  cid,
		Port: port,
	}
	if err := unix.Connect(fd, addr); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("connect vsock: %w", err)
	}

	return os.NewFile(uintptr(fd), fmt.Sprintf("nitro-vsock-%d-%d", cid, port)), nil
}
