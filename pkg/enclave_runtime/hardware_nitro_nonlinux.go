//go:build !linux

package enclave_runtime

import (
	"context"
	"errors"
	"io"
)

func dialNitroVsock(_ context.Context, _, _ uint32) (io.ReadWriteCloser, error) {
	return nil, errors.New("real Nitro vsock transport requires Linux")
}
