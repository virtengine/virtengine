package keeper

import (
	"bytes"
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	enclave "github.com/virtengine/virtengine/pkg/enclave_runtime"
	"github.com/virtengine/virtengine/x/enclave/types"
)

func (k Keeper) verifySGXAttestation(ctx sdk.Context, identity *types.EnclaveIdentity, measurement *types.MeasurementRecord) error {
	quote, err := types.ParseSGXDCAPQuoteV3(identity.AttestationQuote)
	if err != nil {
		return types.ErrAttestationInvalid.Wrapf("parse SGX quote: %v", err)
	}

	if quote.Header.Version < sgxQuoteVersionDCAP {
		return types.ErrInvalidQuoteVersion.Wrapf("quote version %d is below minimum %d", quote.Header.Version, sgxQuoteVersionDCAP)
	}

	if quote.Report.DebugEnabled() {
		return types.ErrDebugModeEnabled
	}

	if !bytes.Equal(identity.MeasurementHash, quote.Report.MRENCLAVE[:]) {
		return types.ErrAttestationInvalid.Wrap("MRENCLAVE does not match measurement hash")
	}

	if len(identity.SignerHash) > 0 && !bytes.Equal(identity.SignerHash, quote.Report.MRSIGNER[:]) {
		return types.ErrAttestationInvalid.Wrap("MRSIGNER does not match signer hash")
	}

	if identity.ISVProdID != 0 && identity.ISVProdID != quote.Report.ISVProdID {
		return types.ErrAttestationInvalid.Wrapf("ISVProdID mismatch: got %d expected %d", quote.Report.ISVProdID, identity.ISVProdID)
	}

	if identity.ISVSVN != 0 && identity.ISVSVN != quote.Report.ISVSVN {
		return types.ErrAttestationInvalid.Wrapf("ISVSVN mismatch: got %d expected %d", quote.Report.ISVSVN, identity.ISVSVN)
	}

	if err := k.verifySGXCryptographic(ctx, identity, quote); err != nil {
		return err
	}

	return nil
}

func (k Keeper) verifySGXCryptographic(ctx sdk.Context, identity *types.EnclaveIdentity, quote *types.SGXDCAPQuote) error {
	dcapVerifier, err := enclave.NewDCAPVerifier()
	if err != nil {
		return types.ErrAttestationInvalid.Wrapf("DCAP verifier init: %v", err)
	}

	dcapResult, err := dcapVerifier.Verify(identity.AttestationQuote)
	if err != nil {
		return types.ErrAttestationInvalid.Wrapf("DCAP verification error: %v", err)
	}
	if dcapResult == nil || !dcapResult.Valid {
		return types.ErrAttestationInvalid.Wrap("DCAP verification failed")
	}

	if len(identity.AttestationChain) == 0 {
		return nil
	}

	chain, err := parseAttestationChain(identity.AttestationChain)
	if err != nil {
		return types.ErrAttestationInvalid.Wrapf("attestation chain parse: %v", err)
	}

	if len(chain.certs) > 0 {
		leaf := selectLeafCertificate(chain.certs)
		rootPEM := [][]byte{
			[]byte(enclave.IntelSGXRootCAPEM),
			[]byte(enclave.IntelSGXPCKProcessorCAPEM),
		}
		if err := verifyChainWithRoots(ctx, leaf, chain.certs, rootPEM); err != nil {
			return types.ErrAttestationInvalid.Wrapf("SGX chain verify: %v", err)
		}
		if err := checkRevocations(ctx, leaf, chain.certs, chain.crls); err != nil {
			return types.ErrAttestationInvalid.Wrapf("SGX revocation: %v", err)
		}
	}

	if err := verifySGXCollateral(ctx, quote, chain); err != nil {
		return types.ErrAttestationInvalid.Wrapf("SGX collateral: %v", err)
	}

	return nil
}

func verifySGXCollateral(ctx sdk.Context, quote *types.SGXDCAPQuote, chain *attestationChain) error {
	if chain == nil || len(chain.json) == 0 {
		return nil
	}

	// Optional TCB Info evaluation.
	if tcbInfoJSON := findJSONWithKey(chain.json, "tcbInfo"); len(tcbInfoJSON) > 0 {
		tcbVerifier := enclave.NewTCBInfoVerifier()
		tcbInfo, err := tcbVerifier.ParseTCBInfo(tcbInfoJSON)
		if err != nil {
			return fmt.Errorf("parse TCB info: %w", err)
		}
		status, err := tcbVerifier.GetTCBStatus(tcbInfo, quote.Report.CPUSVN[:], quote.Header.PCESVN)
		if err != nil {
			return fmt.Errorf("TCB status lookup: %w", err)
		}
		if !enclave.IsTCBStatusAcceptable(status, false) {
			return fmt.Errorf("TCB status %s is not acceptable", status)
		}
	}

	// Optional QE Identity validation.
	if qeIdentityJSON := findJSONWithKey(chain.json, "enclaveIdentity"); len(qeIdentityJSON) > 0 {
		qeVerifier := enclave.NewQEIdentityVerifier()
		qeIdentity, err := qeVerifier.ParseQEIdentity(qeIdentityJSON)
		if err != nil {
			return fmt.Errorf("parse QE identity: %w", err)
		}
		qeReport := convertSGXReport(quote.QEReport)
		if err := qeVerifier.VerifyQEReport(&qeReport, qeIdentity); err != nil {
			return fmt.Errorf("QE report verification failed: %w", err)
		}
	}

	return nil
}

func convertSGXReport(report types.SGXReportBody) enclave.SGXReportBody {
	var body enclave.SGXReportBody
	copy(body.CPUSVN[:], report.CPUSVN[:])
	body.Attributes.Flags = binary.LittleEndian.Uint64(report.Attributes[0:8])
	body.Attributes.Xfrm = binary.LittleEndian.Uint64(report.Attributes[8:16])
	copy(body.MREnclave[:], report.MRENCLAVE[:])
	copy(body.MRSigner[:], report.MRSIGNER[:])
	body.ISVProdID = report.ISVProdID
	body.ISVSVN = report.ISVSVN
	copy(body.ReportData[:], report.ReportData[:])
	return body
}
