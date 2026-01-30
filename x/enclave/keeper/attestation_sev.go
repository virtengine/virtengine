package keeper

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	enclave "github.com/virtengine/virtengine/pkg/enclave_runtime"
	"github.com/virtengine/virtengine/x/enclave/types"
)

func (k Keeper) verifySEVSNPAttestation(ctx sdk.Context, identity *types.EnclaveIdentity, measurement *types.MeasurementRecord) error {
	report, err := types.ParseSEVSNPReport(identity.AttestationQuote)
	if err != nil {
		return types.ErrAttestationInvalid.Wrapf("parse SEV-SNP report: %v", err)
	}

	if report.Version != 1 && report.Version != 2 {
		return types.ErrInvalidQuoteVersion.Wrapf("unsupported SEV-SNP report version: %d", report.Version)
	}

	if report.DebugEnabled() {
		return types.ErrDebugModeEnabled
	}

	if err := verifySEVSNPMeasurement(identity, report.Measurement[:]); err != nil {
		return err
	}

	if len(identity.SignerHash) > 0 && !bytes.Equal(identity.SignerHash, report.IDKeyDigest[:]) {
		return types.ErrAttestationInvalid.Wrap("ID key digest does not match signer hash")
	}

	if len(identity.AttestationChain) == 0 {
		return types.ErrAttestationInvalid.Wrap("attestation chain required for SEV-SNP verification")
	}

	chain, err := parseAttestationChain(identity.AttestationChain)
	if err != nil {
		return types.ErrAttestationInvalid.Wrapf("attestation chain parse: %v", err)
	}

	vcek := findCertificateBySubject(chain.certs, "VCEK")
	if vcek == nil {
		return types.ErrAttestationInvalid.Wrap("VCEK certificate not provided")
	}

	snpVerifier, err := enclave.NewSNPVerifier()
	if err != nil {
		return types.ErrAttestationInvalid.Wrapf("SNP verifier init: %v", err)
	}

	result, err := snpVerifier.VerifyWithCert(identity.AttestationQuote, vcek)
	if err != nil {
		return types.ErrAttestationInvalid.Wrapf("SNP verification error: %v", err)
	}
	if result == nil || !result.Valid {
		return types.ErrAttestationInvalid.Wrap("SNP verification failed")
	}

	if err := checkRevocations(ctx, vcek, chain.certs, chain.crls); err != nil {
		return types.ErrAttestationInvalid.Wrapf("SEV revocation: %v", err)
	}

	return nil
}

func verifySEVSNPMeasurement(identity *types.EnclaveIdentity, measurement []byte) error {
	switch len(identity.MeasurementHash) {
	case 32:
		sum := sha256.Sum256(measurement)
		if !bytes.Equal(identity.MeasurementHash, sum[:]) {
			return types.ErrAttestationInvalid.Wrap("measurement hash does not match launch digest")
		}
	case 48:
		if !bytes.Equal(identity.MeasurementHash, measurement) {
			return types.ErrAttestationInvalid.Wrap("measurement hash does not match launch digest")
		}
	default:
		return types.ErrAttestationInvalid.Wrap("unsupported measurement hash length for SEV-SNP")
	}
	return nil
}
