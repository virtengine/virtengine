package keeper

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	enclave "github.com/virtengine/virtengine/pkg/enclave_runtime"
	"github.com/virtengine/virtengine/x/enclave/types"
)

const (
	sgxQuoteVersionDCAP = 3
)

// verifyAttestation performs platform-specific cryptographic verification.
func (k Keeper) verifyAttestation(ctx sdk.Context, identity *types.EnclaveIdentity, measurement *types.MeasurementRecord) error {
	switch identity.TeeType {
	case types.TEETypeSGX:
		return k.verifySGXAttestation(ctx, identity, measurement)
	case types.TEETypeSEVSNP:
		return k.verifySEVSNPAttestation(ctx, identity, measurement)
	case types.TEETypeNitro:
		return k.verifyNitroAttestation(ctx, identity, measurement)
	default:
		return types.ErrAttestationInvalid.Wrapf("unsupported TEE type: %s", identity.TeeType)
	}
}

type attestationChain struct {
	certs []*x509.Certificate
	crls  []*x509.RevocationList
	json  [][]byte
}

func parseAttestationChain(chain [][]byte) (*attestationChain, error) {
	parsed := &attestationChain{}

	for _, entry := range chain {
		if len(entry) == 0 {
			continue
		}

		if bytes.HasPrefix(bytes.TrimSpace(entry), []byte("{")) {
			parsed.json = append(parsed.json, entry)
			continue
		}

		if parsePEMBlocks(entry, parsed); len(parsed.certs) > 0 || len(parsed.crls) > 0 {
			continue
		}

		if cert, err := x509.ParseCertificate(entry); err == nil {
			parsed.certs = append(parsed.certs, cert)
			continue
		}

		if crl, err := x509.ParseRevocationList(entry); err == nil {
			parsed.crls = append(parsed.crls, crl)
		}
	}

	if len(parsed.certs) == 0 && len(parsed.json) == 0 {
		return parsed, fmt.Errorf("attestation chain contains no usable entries")
	}

	return parsed, nil
}

func parsePEMBlocks(data []byte, out *attestationChain) int {
	parsed := 0
	for len(data) > 0 {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}
		parsed++
		switch block.Type {
		case "CERTIFICATE":
			if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
				out.certs = append(out.certs, cert)
			}
		case "X509 CRL", "CRL":
			if crl, err := x509.ParseRevocationList(block.Bytes); err == nil {
				out.crls = append(out.crls, crl)
			}
		default:
			// ignore other PEM blocks
		}
	}
	return parsed
}

func verifyChainWithRoots(ctx sdk.Context, leaf *x509.Certificate, chain []*x509.Certificate, rootsPEM [][]byte) error {
	if leaf == nil {
		return fmt.Errorf("missing leaf certificate")
	}

	verifier := enclave.NewCertificateChainVerifier()
	verifier.CurrentTime = ctx.BlockTime()

	for _, root := range rootsPEM {
		if err := verifier.AddRootCA(root); err != nil {
			return err
		}
	}
	for _, cert := range chain {
		if cert.Equal(leaf) {
			continue
		}
		if err := verifier.AddIntermediateCA(cert.Raw); err != nil {
			return err
		}
	}

	return verifier.Verify([]*x509.Certificate{leaf})
}

func checkRevocations(ctx sdk.Context, leaf *x509.Certificate, issuerPool []*x509.Certificate, crls []*x509.RevocationList) error {
	if len(crls) == 0 || leaf == nil {
		return nil
	}

	verifyTime := ctx.BlockTime()
	for _, crl := range crls {
		if verifyTime.Before(crl.ThisUpdate) || (!crl.NextUpdate.IsZero() && verifyTime.After(crl.NextUpdate)) {
			continue
		}

		for _, issuer := range issuerPool {
			if err := crl.CheckSignatureFrom(issuer); err == nil {
				for _, revoked := range crl.RevokedCertificateEntries {
					if revoked.SerialNumber.Cmp(leaf.SerialNumber) == 0 {
						return fmt.Errorf("certificate revoked by CRL issued by %s", issuer.Subject.CommonName)
					}
				}
			}
		}
	}

	return nil
}

func findJSONWithKey(entries [][]byte, key string) []byte {
	for _, entry := range entries {
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(entry, &payload); err != nil {
			continue
		}
		if _, ok := payload[key]; ok {
			return entry
		}
	}
	return nil
}

func findCertificateBySubject(certs []*x509.Certificate, contains string) *x509.Certificate {
	for _, cert := range certs {
		if strings.Contains(strings.ToUpper(cert.Subject.CommonName), strings.ToUpper(contains)) {
			return cert
		}
	}
	if len(certs) > 0 {
		return certs[0]
	}
	return nil
}

func selectLeafCertificate(certs []*x509.Certificate) *x509.Certificate {
	for _, cert := range certs {
		if !cert.IsCA {
			return cert
		}
	}
	if len(certs) > 0 {
		return certs[0]
	}
	return nil
}
