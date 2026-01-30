package keeper

import (
	"bytes"
	"crypto/sha256"

	sdk "github.com/cosmos/cosmos-sdk/types"
	enclave "github.com/virtengine/virtengine/pkg/enclave_runtime"
	"github.com/virtengine/virtengine/x/enclave/types"
)

func (k Keeper) verifyNitroAttestation(ctx sdk.Context, identity *types.EnclaveIdentity, measurement *types.MeasurementRecord) error {
	nitroVerifier, err := enclave.NewNitroCryptoVerifier()
	if err != nil {
		return types.ErrAttestationInvalid.Wrapf("Nitro verifier init: %v", err)
	}

	// If measurement hash is a full PCR value, set it as expected PCR0.
	if len(identity.MeasurementHash) == 48 {
		nitroVerifier.SetExpectedPCR(0, identity.MeasurementHash)
	}

	result, err := nitroVerifier.Verify(identity.AttestationQuote)
	if err != nil {
		return types.ErrAttestationInvalid.Wrapf("Nitro verification error: %v", err)
	}
	if result == nil || !result.Valid || result.Document == nil {
		return types.ErrAttestationInvalid.Wrap("Nitro verification failed")
	}

	pcr0, ok := enclave.GetPCR0(result.Document)
	if !ok {
		return types.ErrAttestationInvalid.Wrap("PCR0 missing in attestation document")
	}

	if err := verifyNitroMeasurement(identity, pcr0); err != nil {
		return err
	}

	return nil
}

func verifyNitroMeasurement(identity *types.EnclaveIdentity, pcr0 []byte) error {
	switch len(identity.MeasurementHash) {
	case 32:
		sum := sha256.Sum256(pcr0)
		if !bytes.Equal(identity.MeasurementHash, sum[:]) {
			return types.ErrAttestationInvalid.Wrap("measurement hash does not match PCR0")
		}
	case 48:
		if !bytes.Equal(identity.MeasurementHash, pcr0) {
			return types.ErrAttestationInvalid.Wrap("measurement hash does not match PCR0")
		}
	default:
		return types.ErrAttestationInvalid.Wrap("unsupported measurement hash length for Nitro")
	}
	return nil
}
