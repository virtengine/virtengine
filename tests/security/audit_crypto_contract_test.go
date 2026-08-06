//go:build security

package security

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	enclavetypes "github.com/virtengine/virtengine/x/enclave/types"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

const (
	auditTestSGXQuoteHeaderSize     = 48
	auditTestSGXReportBodySize      = 384
	auditTestSGXQuoteSigDataOffset  = auditTestSGXQuoteHeaderSize + auditTestSGXReportBodySize + 4
	auditTestSGXQuoteSigDataMinSize = 64 + 64 + auditTestSGXReportBodySize + 64 + 2 + 2 + 4
	auditTestSEVSNPMinReportSize    = 0x4A0
)

func TestAuditCrypto_EmbeddingEnvelopeValidation(t *testing.T) {
	account := sdk.AccAddress(bytes.Repeat([]byte{0x42}, 20)).String()
	embedding := []byte("deterministic-embedding-vector")

	envelope := veidtypes.NewEmbeddingEnvelope(
		"env-audit-001",
		account,
		veidtypes.EmbeddingTypeFace,
		veidtypes.ComputeEmbeddingHash(embedding),
		"model-v1",
		"sha256:model-v1-hash",
		512,
		"scope-face",
		time.Unix(1_700_000_000, 0).UTC(),
		128,
		"validator-audit",
	)

	require.NoError(t, envelope.Validate())
	require.True(t, envelope.MatchesEmbedding(embedding))
	require.False(t, envelope.MatchesEmbedding([]byte("tampered-embedding-vector")))

	badHash := *envelope
	badHash.EmbeddingHash = []byte("too-short")
	err := badHash.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "embedding_hash")

	badDimension := *envelope
	badDimension.Dimension = 0
	err = badDimension.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "dimension cannot be zero")
}

func TestAuditCrypto_SaltBindingRequiresFreshTimestampAndValidSignatures(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	address := sdk.AccAddress(bytes.Repeat([]byte{0x55}, 20))

	t.Run("ed25519_round_trip", func(t *testing.T) {
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		binding := &veidkeeper.SaltBindingData{
			Salt:      []byte("audit-salt-ed25519"),
			Address:   address,
			ScopeID:   "scope-ed25519",
			Timestamp: now,
		}
		signature := ed25519.Sign(privKey, binding.Payload())

		err = veidkeeper.VerifySaltBindingWithParams(&veidkeeper.SaltBindingVerifyParams{
			BindingData: binding,
			Signature:   signature,
			PubKey:      pubKey,
			Algorithm:   veidkeeper.AlgorithmEd25519,
			CurrentTime: now,
		})
		require.NoError(t, err)
	})

	t.Run("secp256k1_round_trip", func(t *testing.T) {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey().Bytes()

		binding := &veidkeeper.SaltBindingData{
			Salt:      []byte("audit-salt-secp256k1"),
			Address:   address,
			ScopeID:   "scope-secp256k1",
			Timestamp: now,
		}
		payloadHash := sha256.Sum256(binding.Payload())
		signature, err := privKey.Sign(payloadHash[:])
		require.NoError(t, err)

		err = veidkeeper.VerifySaltBindingWithParams(&veidkeeper.SaltBindingVerifyParams{
			BindingData: binding,
			Signature:   signature,
			PubKey:      pubKey,
			Algorithm:   veidkeeper.AlgorithmSecp256k1,
			CurrentTime: now,
		})
		require.NoError(t, err)
	})

	t.Run("stale_timestamp_is_rejected", func(t *testing.T) {
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		binding := &veidkeeper.SaltBindingData{
			Salt:      []byte("audit-stale-salt"),
			Address:   address,
			ScopeID:   "scope-stale",
			Timestamp: now.Add(-(veidkeeper.SaltBindingMaxAge + time.Second)),
		}
		signature := ed25519.Sign(privKey, binding.Payload())

		err = veidkeeper.VerifySaltBindingWithParams(&veidkeeper.SaltBindingVerifyParams{
			BindingData: binding,
			Signature:   signature,
			PubKey:      pubKey,
			Algorithm:   veidkeeper.AlgorithmEd25519,
			CurrentTime: now,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "too old")
	})

	t.Run("future_timestamp_is_rejected", func(t *testing.T) {
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		binding := &veidkeeper.SaltBindingData{
			Salt:      []byte("audit-future-salt"),
			Address:   address,
			ScopeID:   "scope-future",
			Timestamp: now.Add(veidkeeper.SaltBindingMaxFuture + time.Second),
		}
		signature := ed25519.Sign(privKey, binding.Payload())

		err = veidkeeper.VerifySaltBindingWithParams(&veidkeeper.SaltBindingVerifyParams{
			BindingData: binding,
			Signature:   signature,
			PubKey:      pubKey,
			Algorithm:   veidkeeper.AlgorithmEd25519,
			CurrentTime: now,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "future")
	})
}

func TestAuditCrypto_AttestationParsersRejectMalformedReportsAndExposeDebugFlags(t *testing.T) {
	t.Run("sgx_quote_rejects_short_input", func(t *testing.T) {
		_, err := enclavetypes.ParseSGXDCAPQuoteV3([]byte("short"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "quote too small")
	})

	t.Run("sgx_quote_exposes_debug_flag", func(t *testing.T) {
		quote := makeAuditTestSGXQuote(true)

		parsed, err := enclavetypes.ParseSGXDCAPQuoteV3(quote)
		require.NoError(t, err)
		require.True(t, parsed.Report.DebugEnabled())
		require.Equal(t, uint16(3), parsed.Header.Version)
		require.Equal(t, uint16(7), parsed.Report.ISVSVN)
		require.Equal(t, bytes.Repeat([]byte{0xA1}, 32), parsed.Report.MRENCLAVE[:])
	})

	t.Run("sev_report_rejects_short_input", func(t *testing.T) {
		_, err := enclavetypes.ParseSEVSNPReport([]byte("tiny"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "report too small")
	})

	t.Run("sev_report_exposes_debug_policy_and_measurement_hash", func(t *testing.T) {
		report := makeAuditTestSEVSNPReport(true)

		parsed, err := enclavetypes.ParseSEVSNPReport(report)
		require.NoError(t, err)
		require.True(t, parsed.DebugEnabled())

		firstHash := enclavetypes.SEVSNPMeasurementHash(parsed.Measurement[:])
		secondHash := enclavetypes.SEVSNPMeasurementHash(parsed.Measurement[:])
		require.Equal(t, firstHash, secondHash)
	})
}

func makeAuditTestSGXQuote(debug bool) []byte {
	quote := make([]byte, auditTestSGXQuoteSigDataOffset+auditTestSGXQuoteSigDataMinSize)
	binary.LittleEndian.PutUint16(quote[0:2], 3)

	report := quote[auditTestSGXQuoteHeaderSize : auditTestSGXQuoteHeaderSize+auditTestSGXReportBodySize]
	if debug {
		binary.LittleEndian.PutUint64(report[48:56], 0x02)
	}
	copy(report[64:96], bytes.Repeat([]byte{0xA1}, 32))
	copy(report[128:160], bytes.Repeat([]byte{0xB2}, 32))
	binary.LittleEndian.PutUint16(report[256:258], 9)
	binary.LittleEndian.PutUint16(report[258:260], 7)
	copy(report[320:384], bytes.Repeat([]byte{0xC3}, 64))

	qeReportStart := auditTestSGXQuoteSigDataOffset + 64 + 64
	qeReport := quote[qeReportStart : qeReportStart+auditTestSGXReportBodySize]
	copy(qeReport[64:96], bytes.Repeat([]byte{0xD4}, 32))

	return quote
}

func makeAuditTestSEVSNPReport(debug bool) []byte {
	report := make([]byte, auditTestSEVSNPMinReportSize)
	binary.LittleEndian.PutUint32(report[0:4], 2)
	if debug {
		binary.LittleEndian.PutUint64(report[0x008:0x010], 1<<19)
	}
	copy(report[0x050:0x090], bytes.Repeat([]byte{0xE5}, 64))
	copy(report[0x090:0x0C0], bytes.Repeat([]byte{0xF6}, 48))
	copy(report[0x0E0:0x100], bytes.Repeat([]byte{0x17}, 32))
	copy(report[0x100:0x120], bytes.Repeat([]byte{0x28}, 32))
	return report
}
