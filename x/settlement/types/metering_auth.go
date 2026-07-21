// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	// SignatureVersionV1 identifies the Task 84B canonical binary envelope.
	SignatureVersionV1 uint32 = 1

	UsageProviderDomainV1 = "virtengine.settlement.usage.provider.v1"
	UsageCustomerDomainV1 = "virtengine.settlement.usage.customer.v1"
	SignerRoleProvider    = "provider"
	SignerRoleCustomer    = "customer"

	DigestSize    = sha256.Size
	ReplayKeySize = sha256.Size

	MaxCanonicalFieldBytes   = 2048
	MaxUsagePeriodSeconds    = int64(24 * 60 * 60)
	MaxProofLifetimeBlocks   = int64(200)
	MaxProofLifetimeSeconds  = int64(60 * 60)
	MaxProofPastBlocks       = int64(1000)
	MaxProofFutureBlocks     = int64(2)
	MaxProofPastSeconds      = int64(2 * 60 * 60)
	MaxProofFutureSeconds    = int64(2 * 60)
	MaxUsagePeriodGapSeconds = int64(24 * 60 * 60)
	MaxRawMetricValue        = int64(1_000_000_000_000_000_000)
	MaxUsageUnits            = uint64(1_000_000_000_000_000)
)

var (
	usageMagic  = [8]byte{'V', 'E', 'U', 'S', 'A', 'G', 'E', 0x01}
	ackMagic    = [8]byte{'V', 'E', 'U', 'S', 'A', 'C', 'K', 0x01}
	streamMagic = [8]byte{'V', 'E', 'U', 'S', 'T', 'R', 'M', 0x01}
)

// RawUsageMetrics contains exact integer dimensions. Values are period deltas,
// never cumulative floating-point observations.
type RawUsageMetrics struct {
	CPUMilliSeconds    int64 `json:"cpu_milli_seconds"`
	MemoryByteSeconds  int64 `json:"memory_byte_seconds"`
	StorageByteSeconds int64 `json:"storage_byte_seconds"`
	NetworkBytesIn     int64 `json:"network_bytes_in"`
	NetworkBytesOut    int64 `json:"network_bytes_out"`
	GPUSeconds         int64 `json:"gpu_seconds"`
}

// CanonicalUsagePayload is the semantic input to the version-1 binary
// provider sign-byte encoder. Field order in CanonicalUsageSignBytes is part of
// the consensus-independent cryptographic protocol and must not change.
type CanonicalUsagePayload struct {
	SignatureVersion uint32
	ChainID          string
	Domain           string
	SignerRole       string
	Provider         string
	Customer         string
	OrderID          string
	LeaseID          string
	AllocationID     string
	PeriodStart      int64
	PeriodEnd        int64
	Metrics          RawUsageMetrics
	PricingVersion   uint32
	UsageUnits       uint64
	UsageType        string
	UnitPriceDenom   string
	UnitPriceAmount  string
	FormulaVersion   uint32
	ModelVersion     uint32
	Sequence         uint64
	Nonce            []byte
	IdempotencyKey   []byte
	ProviderKeyEpoch uint64
	ProviderKeyID    string
	IssuedAtHeight   int64
	ExpiresAtHeight  int64
	IssuedAtUnix     int64
	ExpiresAtUnix    int64
}

// CanonicalAcknowledgmentPayload binds a customer's detached signature to the
// exact authenticated usage digest and an independent replay key.
type CanonicalAcknowledgmentPayload struct {
	SignatureVersion uint32
	ChainID          string
	Domain           string
	SignerRole       string
	Customer         string
	UsageID          string
	UsageDigest      []byte
	ReplayKey        []byte
	IssuedAtHeight   int64
	ExpiresAtHeight  int64
	IssuedAtUnix     int64
	ExpiresAtUnix    int64
}

// CanonicalUsageSignBytes encodes a usage statement as fixed-width integers and
// uint32-length-prefixed byte strings. No JSON, maps, field omission, varints,
// locale, or implementation-specific decimal formatting participates.
func CanonicalUsageSignBytes(payload CanonicalUsagePayload) ([]byte, error) {
	if err := validateCanonicalUsage(payload); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	out.Write(usageMagic[:])
	writeUint32(&out, payload.SignatureVersion)
	for _, value := range []string{
		payload.ChainID,
		payload.Domain,
		payload.SignerRole,
		payload.Provider,
		payload.Customer,
		payload.OrderID,
		payload.LeaseID,
		payload.AllocationID,
	} {
		writeString(&out, value)
	}
	writeInt64(&out, payload.PeriodStart)
	writeInt64(&out, payload.PeriodEnd)
	writeInt64(&out, payload.Metrics.CPUMilliSeconds)
	writeInt64(&out, payload.Metrics.MemoryByteSeconds)
	writeInt64(&out, payload.Metrics.StorageByteSeconds)
	writeInt64(&out, payload.Metrics.NetworkBytesIn)
	writeInt64(&out, payload.Metrics.NetworkBytesOut)
	writeInt64(&out, payload.Metrics.GPUSeconds)
	writeUint32(&out, payload.PricingVersion)
	writeUint64(&out, payload.UsageUnits)
	writeString(&out, payload.UsageType)
	writeString(&out, payload.UnitPriceDenom)
	writeString(&out, payload.UnitPriceAmount)
	writeUint32(&out, payload.FormulaVersion)
	writeUint32(&out, payload.ModelVersion)
	writeUint64(&out, payload.Sequence)
	writeBytes(&out, payload.Nonce)
	writeBytes(&out, payload.IdempotencyKey)
	writeUint64(&out, payload.ProviderKeyEpoch)
	writeString(&out, payload.ProviderKeyID)
	writeInt64(&out, payload.IssuedAtHeight)
	writeInt64(&out, payload.ExpiresAtHeight)
	writeInt64(&out, payload.IssuedAtUnix)
	writeInt64(&out, payload.ExpiresAtUnix)
	return out.Bytes(), nil
}

// CanonicalUsageDigest returns SHA-256 over the complete canonical sign bytes.
func CanonicalUsageDigest(payload CanonicalUsagePayload) ([]byte, error) {
	signBytes, err := CanonicalUsageSignBytes(payload)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(signBytes)
	return digest[:], nil
}

// CanonicalAcknowledgmentSignBytes encodes the customer acknowledgment.
func CanonicalAcknowledgmentSignBytes(payload CanonicalAcknowledgmentPayload) ([]byte, error) {
	if err := validateCanonicalAcknowledgment(payload); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	out.Write(ackMagic[:])
	writeUint32(&out, payload.SignatureVersion)
	writeString(&out, payload.ChainID)
	writeString(&out, payload.Domain)
	writeString(&out, payload.SignerRole)
	writeString(&out, payload.Customer)
	writeString(&out, payload.UsageID)
	writeBytes(&out, payload.UsageDigest)
	writeBytes(&out, payload.ReplayKey)
	writeInt64(&out, payload.IssuedAtHeight)
	writeInt64(&out, payload.ExpiresAtHeight)
	writeInt64(&out, payload.IssuedAtUnix)
	writeInt64(&out, payload.ExpiresAtUnix)
	return out.Bytes(), nil
}

// UsageStreamID creates a collision-safe stream identity. Allocation is the
// primary lineage when present; order and lease remain bound to prevent a
// reused allocation identifier from crossing economic subjects.
func UsageStreamID(provider, allocationID, orderID, leaseID string) ([]byte, error) {
	if provider == "" || orderID == "" || leaseID == "" {
		return nil, fmt.Errorf("provider, order_id, and lease_id are required")
	}
	if err := validateCanonicalStrings(provider, allocationID, orderID, leaseID); err != nil {
		return nil, err
	}
	lineageKind := "lease_order"
	if allocationID != "" {
		lineageKind = "allocation"
	}
	var out bytes.Buffer
	out.Write(streamMagic[:])
	writeString(&out, provider)
	writeString(&out, lineageKind)
	writeString(&out, allocationID)
	writeString(&out, orderID)
	writeString(&out, leaseID)
	digest := sha256.Sum256(out.Bytes())
	return digest[:], nil
}

// DeriveReplayKey derives a 32-byte key from a domain and length-prefixed
// values. It is suitable for durable non-random producer identities.
func DeriveReplayKey(domain string, values ...string) ([]byte, error) {
	if domain == "" {
		return nil, fmt.Errorf("replay key domain is required")
	}
	all := append([]string{domain}, values...)
	if err := validateCanonicalStrings(all...); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	writeString(&out, domain)
	for _, value := range values {
		writeString(&out, value)
	}
	digest := sha256.Sum256(out.Bytes())
	return digest[:], nil
}

func validateCanonicalUsage(payload CanonicalUsagePayload) error {
	if payload.SignatureVersion != SignatureVersionV1 {
		return fmt.Errorf("unsupported usage signature version: %d", payload.SignatureVersion)
	}
	if payload.Domain != UsageProviderDomainV1 || payload.SignerRole != SignerRoleProvider {
		return fmt.Errorf("invalid provider domain or signer role")
	}
	if payload.ChainID == "" || payload.Provider == "" || payload.Customer == "" || payload.OrderID == "" || payload.LeaseID == "" {
		return fmt.Errorf("chain and economic lineage fields are required")
	}
	if payload.UsageType == "" || payload.UnitPriceDenom == "" || payload.ProviderKeyID == "" {
		return fmt.Errorf("usage, price, and key identifiers are required")
	}
	if err := validateCanonicalStrings(
		payload.ChainID, payload.Domain, payload.SignerRole, payload.Provider,
		payload.Customer, payload.OrderID, payload.LeaseID, payload.AllocationID,
		payload.UsageType, payload.UnitPriceDenom, payload.UnitPriceAmount, payload.ProviderKeyID,
	); err != nil {
		return err
	}
	if payload.PeriodStart < 0 || payload.PeriodEnd <= payload.PeriodStart {
		return fmt.Errorf("usage period must be positive and increasing")
	}
	if payload.PeriodEnd-payload.PeriodStart > MaxUsagePeriodSeconds {
		return fmt.Errorf("usage period exceeds %d seconds", MaxUsagePeriodSeconds)
	}
	if err := validateRawMetrics(payload.Metrics); err != nil {
		return err
	}
	if payload.PricingVersion == 0 || payload.FormulaVersion == 0 || payload.ModelVersion == 0 {
		return fmt.Errorf("pricing, formula, and model versions must be non-zero")
	}
	if payload.UsageUnits == 0 || payload.UsageUnits > MaxUsageUnits {
		return fmt.Errorf("usage_units outside supported range")
	}
	expectedUnits, err := ExpectedUsageUnitsV1(payload.UsageType, payload.Metrics)
	if err != nil {
		return err
	}
	if payload.UsageUnits != expectedUnits {
		return fmt.Errorf("usage_units %d do not match version-1 raw metric formula result %d", payload.UsageUnits, expectedUnits)
	}
	if payload.Sequence == 0 || payload.ProviderKeyEpoch == 0 {
		return fmt.Errorf("sequence and key epoch must be non-zero")
	}
	if len(payload.Nonce) != ReplayKeySize || len(payload.IdempotencyKey) != ReplayKeySize {
		return fmt.Errorf("nonce and idempotency key must each be %d bytes", ReplayKeySize)
	}
	if err := validateProofBounds(payload.IssuedAtHeight, payload.ExpiresAtHeight, payload.IssuedAtUnix, payload.ExpiresAtUnix); err != nil {
		return err
	}

	price, err := sdkmath.LegacyNewDecFromStr(payload.UnitPriceAmount)
	if err != nil || price.IsNegative() || price.IsZero() || price.String() != payload.UnitPriceAmount {
		return fmt.Errorf("unit price amount is not a positive canonical SDK decimal")
	}
	quantity := sdkmath.NewIntFromUint64(payload.UsageUnits)
	total := price.MulInt(quantity).TruncateInt()
	if total.IsNegative() || !total.IsInt64() || total.GT(sdkmath.NewInt(math.MaxInt64)) {
		return fmt.Errorf("unit price multiplication exceeds supported integer range")
	}
	return nil
}

// ExpectedUsageUnitsV1 applies the conservative integer metering formula. Each
// nonzero sub-unit period rounds up to one billable unit. An all-zero CPU report
// retains the documented one-unit minimum accounting footprint.
func ExpectedUsageUnitsV1(usageType string, metrics RawUsageMetrics) (uint64, error) {
	const (
		cpuHourMillis = int64(1000 * 60 * 60)
		gb            = int64(1024 * 1024 * 1024)
		gbHour        = gb * 60 * 60
	)
	ceilUnits := func(value, divisor int64) uint64 {
		if value <= 0 {
			return 0
		}
		return uint64((value-1)/divisor + 1) //nolint:gosec // positive bounded int64 result
	}

	switch strings.ToLower(usageType) {
	case "cpu", "compute", "cpu_core_hours":
		units := ceilUnits(metrics.CPUMilliSeconds, cpuHourMillis)
		if units == 0 {
			units = 1
		}
		return units, nil
	case "memory", "memory_gb_hours":
		return nonzeroFormulaUnits(ceilUnits(metrics.MemoryByteSeconds, gbHour), usageType)
	case "storage", "storage_gb_hours":
		return nonzeroFormulaUnits(ceilUnits(metrics.StorageByteSeconds, gbHour), usageType)
	case "gpu", "gpu_hours":
		return nonzeroFormulaUnits(ceilUnits(metrics.GPUSeconds, 60*60), usageType)
	case "network", "network_gb":
		if metrics.NetworkBytesIn > math.MaxInt64-metrics.NetworkBytesOut {
			return 0, fmt.Errorf("network metric sum overflow")
		}
		return nonzeroFormulaUnits(ceilUnits(metrics.NetworkBytesIn+metrics.NetworkBytesOut, gb), usageType)
	case "fixed":
		return 1, nil
	default:
		return 0, fmt.Errorf("usage type %q has no version-1 metric formula", usageType)
	}
}

func nonzeroFormulaUnits(units uint64, usageType string) (uint64, error) {
	if units == 0 {
		return 0, fmt.Errorf("usage type %q has no positive raw metric delta", usageType)
	}
	return units, nil
}

func validateCanonicalAcknowledgment(payload CanonicalAcknowledgmentPayload) error {
	if payload.SignatureVersion != SignatureVersionV1 {
		return fmt.Errorf("unsupported acknowledgment signature version: %d", payload.SignatureVersion)
	}
	if payload.Domain != UsageCustomerDomainV1 || payload.SignerRole != SignerRoleCustomer {
		return fmt.Errorf("invalid customer domain or signer role")
	}
	if payload.ChainID == "" || payload.Customer == "" || payload.UsageID == "" {
		return fmt.Errorf("chain, customer, and usage_id are required")
	}
	if err := validateCanonicalStrings(payload.ChainID, payload.Domain, payload.SignerRole, payload.Customer, payload.UsageID); err != nil {
		return err
	}
	if len(payload.UsageDigest) != DigestSize || len(payload.ReplayKey) != ReplayKeySize {
		return fmt.Errorf("usage digest and replay key must each be %d bytes", DigestSize)
	}
	return validateProofBounds(payload.IssuedAtHeight, payload.ExpiresAtHeight, payload.IssuedAtUnix, payload.ExpiresAtUnix)
}

func validateRawMetrics(metrics RawUsageMetrics) error {
	values := []int64{
		metrics.CPUMilliSeconds,
		metrics.MemoryByteSeconds,
		metrics.StorageByteSeconds,
		metrics.NetworkBytesIn,
		metrics.NetworkBytesOut,
		metrics.GPUSeconds,
	}
	for _, value := range values {
		if value < 0 || value > MaxRawMetricValue {
			return fmt.Errorf("raw metric outside supported range")
		}
	}
	return nil
}

func validateProofBounds(issuedHeight, expiresHeight, issuedUnix, expiresUnix int64) error {
	if issuedHeight <= 0 || expiresHeight < issuedHeight || expiresHeight-issuedHeight > MaxProofLifetimeBlocks {
		return fmt.Errorf("invalid proof height bounds")
	}
	if issuedUnix <= 0 || expiresUnix < issuedUnix || expiresUnix-issuedUnix > MaxProofLifetimeSeconds {
		return fmt.Errorf("invalid proof time bounds")
	}
	return nil
}

func validateCanonicalStrings(values ...string) error {
	for _, value := range values {
		if len(value) > MaxCanonicalFieldBytes {
			return fmt.Errorf("canonical field exceeds %d bytes", MaxCanonicalFieldBytes)
		}
	}
	return nil
}

func writeString(out *bytes.Buffer, value string) {
	writeBytes(out, []byte(value))
}

func writeBytes(out *bytes.Buffer, value []byte) {
	writeUint32(out, uint32(len(value))) //nolint:gosec // validated against a 2 KiB protocol maximum
	_, _ = out.Write(value)
}

func writeUint32(out *bytes.Buffer, value uint32) {
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], value)
	_, _ = out.Write(encoded[:])
}

func writeUint64(out *bytes.Buffer, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = out.Write(encoded[:])
}

func writeInt64(out *bytes.Buffer, value int64) {
	writeUint64(out, uint64(value)) //nolint:gosec // preserves signed two's-complement bits
}
