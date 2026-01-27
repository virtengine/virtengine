package v1beta3

import (
	"errors"
)

var (
	ErrCertificate = errors.New("certificate error")
)

var (
	// ErrCertificateNotFound certificate not found
	ErrCertificateNotFound = errors.New("certificate not found")

	// ErrInvalidAddress invalid trusted auditor address
	ErrInvalidAddress = errors.New("invalid address")

	// ErrCertificateExists certificate already exists
	ErrCertificateExists = errors.New("certificate exists")

	// ErrCertificateAlreadyRevoked certificate already revoked
	ErrCertificateAlreadyRevoked = errors.New("certificate already revoked")

	// ErrInvalidSerialNumber invalid serial number
	ErrInvalidSerialNumber = errors.New("invalid serial number")

	// ErrInvalidCertificateValue certificate content is not valid
	ErrInvalidCertificateValue = errors.New("invalid certificate value")

	// ErrInvalidPubkeyValue public key is not valid
	ErrInvalidPubkeyValue = errors.New("invalid pubkey value")

	// ErrInvalidState invalid certificate state
	ErrInvalidState = errors.New("invalid state")

	// ErrInvalidKeySize invalid certificate state
	ErrInvalidKeySize = errors.New("invalid key size")
)
