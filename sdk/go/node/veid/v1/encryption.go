package v1

import (
	"fmt"
)

// Validate checks basic integrity of the envelope
func (e *EncryptedPayloadEnvelope) Validate() error {
	if e == nil {
		return fmt.Errorf("empty envelope")
	}

	if e.AlgorithmId == "" {
		return fmt.Errorf("algorithm_id required")
	}

	if len(e.Nonce) == 0 {
		return fmt.Errorf("nonce cannot be empty")
	}

	if len(e.Ciphertext) == 0 {
		return fmt.Errorf("ciphertext cannot be empty")
	}

	if len(e.SenderPubKey) == 0 {
		return fmt.Errorf("sender public key required")
	}

	if len(e.SenderSignature) == 0 {
		return fmt.Errorf("sender signature required")
	}
	
	return nil
}

