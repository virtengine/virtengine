package types

import (
	"bytes"
	"encoding/hex"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func canonicalUsageFixture() CanonicalUsagePayload {
	return CanonicalUsagePayload{
		SignatureVersion: SignatureVersionV1,
		ChainID:          "virtengine-test-1",
		Domain:           UsageProviderDomainV1,
		SignerRole:       SignerRoleProvider,
		Provider:         "virtengine1provider",
		Customer:         "virtengine1customer",
		OrderID:          "order-42",
		LeaseID:          "lease-7",
		AllocationID:     "allocation-3",
		PeriodStart:      1_700_000_000,
		PeriodEnd:        1_700_003_600,
		Metrics: RawUsageMetrics{
			CPUMilliSeconds:    3_600_000,
			MemoryByteSeconds:  4_294_967_296,
			StorageByteSeconds: 8_589_934_592,
			NetworkBytesIn:     1_024,
			NetworkBytesOut:    2_048,
			GPUSeconds:         3_600,
		},
		PricingVersion:   1,
		UsageUnits:       1,
		UsageType:        "cpu",
		UnitPriceDenom:   "uve",
		UnitPriceAmount:  "12.500000000000000000",
		FormulaVersion:   1,
		ModelVersion:     1,
		Sequence:         9,
		Nonce:            bytes.Repeat([]byte{0x11}, ReplayKeySize),
		IdempotencyKey:   bytes.Repeat([]byte{0x22}, ReplayKeySize),
		ProviderKeyEpoch: 4,
		ProviderKeyID:    "ed25519:0011223344556677",
		IssuedAtHeight:   100,
		ExpiresAtHeight:  120,
		IssuedAtUnix:     1_700_003_601,
		ExpiresAtUnix:    1_700_007_201,
	}
}

func TestCanonicalUsageSignBytesGoldenAndUnambiguous(t *testing.T) {
	payload := canonicalUsageFixture()
	got, err := CanonicalUsageSignBytes(payload)
	require.NoError(t, err)

	const goldenHex = "5645555341474501000000010000001176697274656e67696e652d746573742d310000002776697274656e67696e652e736574746c656d656e742e75736167652e70726f76696465722e76310000000870726f76696465720000001376697274656e67696e653170726f76696465720000001376697274656e67696e6531637573746f6d6572000000086f726465722d3432000000076c656173652d370000000c616c6c6f636174696f6e2d33000000006553f100000000006553ff10000000000036ee8000000001000000000000000200000000000000000000040000000000000008000000000000000e1000000001000000000000000100000003637075000000037576650000001531322e35303030303030303030303030303030303000000001000000010000000000000009000000201111111111111111111111111111111111111111111111111111111111111111000000202222222222222222222222222222222222222222222222222222222222222222000000000000000400000018656432353531393a3030313132323333343435353636373700000000000000640000000000000078000000006553ff110000000065540d21"
	require.Equal(t, goldenHex, hex.EncodeToString(got))

	altered := payload
	altered.OrderID = payload.OrderID + "\x00lease-7"
	alteredBytes, err := CanonicalUsageSignBytes(altered)
	require.NoError(t, err)
	require.NotEqual(t, got, alteredBytes)

	digest, err := CanonicalUsageDigest(payload)
	require.NoError(t, err)
	require.Len(t, digest, DigestSize)
}

func TestCanonicalUsageValidationRejectsOverflowAndInvalidBounds(t *testing.T) {
	base := canonicalUsageFixture()

	tests := []struct {
		name   string
		mutate func(*CanonicalUsagePayload)
	}{
		{"unsupported version", func(p *CanonicalUsagePayload) { p.SignatureVersion = 99 }},
		{"wrong domain", func(p *CanonicalUsagePayload) { p.Domain = UsageCustomerDomainV1 }},
		{"negative metric", func(p *CanonicalUsagePayload) { p.Metrics.GPUSeconds = -1 }},
		{"metric too large", func(p *CanonicalUsagePayload) { p.Metrics.CPUMilliSeconds = MaxRawMetricValue + 1 }},
		{"zero sequence", func(p *CanonicalUsagePayload) { p.Sequence = 0 }},
		{"bad nonce", func(p *CanonicalUsagePayload) { p.Nonce = []byte{1} }},
		{"bad idempotency key", func(p *CanonicalUsagePayload) { p.IdempotencyKey = []byte{2} }},
		{"period reversed", func(p *CanonicalUsagePayload) { p.PeriodEnd = p.PeriodStart - 1 }},
		{"period too long", func(p *CanonicalUsagePayload) { p.PeriodEnd = p.PeriodStart + MaxUsagePeriodSeconds + 1 }},
		{"expiry before issue", func(p *CanonicalUsagePayload) { p.ExpiresAtHeight = p.IssuedAtHeight - 1 }},
		{"height lifetime too long", func(p *CanonicalUsagePayload) { p.ExpiresAtHeight = p.IssuedAtHeight + MaxProofLifetimeBlocks + 1 }},
		{"time lifetime too long", func(p *CanonicalUsagePayload) { p.ExpiresAtUnix = p.IssuedAtUnix + MaxProofLifetimeSeconds + 1 }},
		{"noncanonical price", func(p *CanonicalUsagePayload) { p.UnitPriceAmount = "01.0" }},
		{"price overflow", func(p *CanonicalUsagePayload) { p.UsageUnits = math.MaxUint64 }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := base
			payload.Nonce = append([]byte(nil), base.Nonce...)
			payload.IdempotencyKey = append([]byte(nil), base.IdempotencyKey...)
			tc.mutate(&payload)
			_, err := CanonicalUsageSignBytes(payload)
			require.Error(t, err)
		})
	}
}

func TestCanonicalAcknowledgmentBindsStoredDigestAndReplayKey(t *testing.T) {
	usage := canonicalUsageFixture()
	digest, err := CanonicalUsageDigest(usage)
	require.NoError(t, err)

	ack := CanonicalAcknowledgmentPayload{
		SignatureVersion: SignatureVersionV1,
		ChainID:          usage.ChainID,
		Domain:           UsageCustomerDomainV1,
		SignerRole:       SignerRoleCustomer,
		Customer:         usage.Customer,
		UsageID:          "usage-1700003601-9",
		UsageDigest:      digest,
		ReplayKey:        bytes.Repeat([]byte{0x33}, ReplayKeySize),
		IssuedAtHeight:   110,
		ExpiresAtHeight:  120,
		IssuedAtUnix:     1_700_003_700,
		ExpiresAtUnix:    1_700_004_000,
	}

	first, err := CanonicalAcknowledgmentSignBytes(ack)
	require.NoError(t, err)
	ack.UsageDigest[0] ^= 0xff
	second, err := CanonicalAcknowledgmentSignBytes(ack)
	require.NoError(t, err)
	require.NotEqual(t, first, second)
}

func TestUsageStreamIdentityIsCollisionSafeAndLineageBound(t *testing.T) {
	first, err := UsageStreamID("provider", "allocation", "order", "lease")
	require.NoError(t, err)
	second, err := UsageStreamID("providera", "llocation", "order", "lease")
	require.NoError(t, err)
	crossOrder, err := UsageStreamID("provider", "allocation", "other-order", "lease")
	require.NoError(t, err)
	require.NotEqual(t, first, second)
	require.NotEqual(t, first, crossOrder)
}

func FuzzCanonicalUsageSignBytes(f *testing.F) {
	f.Add("order-42", "lease-7", "allocation-3", uint64(9), int64(3_600_000))
	f.Fuzz(func(t *testing.T, orderID, leaseID, allocationID string, sequence uint64, cpuMillis int64) {
		if len(orderID) > 128 || len(leaseID) > 128 || len(allocationID) > 128 {
			t.Skip()
		}
		payload := canonicalUsageFixture()
		payload.OrderID = orderID
		payload.LeaseID = leaseID
		payload.AllocationID = allocationID
		payload.Sequence = sequence
		payload.Metrics.CPUMilliSeconds = cpuMillis
		first, firstErr := CanonicalUsageSignBytes(payload)
		second, secondErr := CanonicalUsageSignBytes(payload)
		require.Equal(t, firstErr == nil, secondErr == nil)
		if firstErr == nil {
			require.Equal(t, first, second)
			mutated := payload
			mutated.OrderID += "x"
			mutatedBytes, err := CanonicalUsageSignBytes(mutated)
			if err == nil {
				require.NotEqual(t, first, mutatedBytes)
			}
		}
	})
}
