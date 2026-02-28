//go:build !linux

package enclave_runtime

import (
	"errors"
	"os"
)

func requestSEVHardwareReport(_ *os.File, _ [64]byte, _ uint32) ([]byte, error) {
	return nil, errors.New("real SEV-SNP guest report requests require Linux")
}

func requestSEVDerivedKey(_ *os.File, _ int, _ uint64, _ uint32) ([]byte, error) {
	return nil, errors.New("real SEV-SNP derived key requests require Linux")
}
