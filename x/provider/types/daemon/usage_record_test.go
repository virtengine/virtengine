//go:build ignore
// +build ignore

// TODO: This test file is excluded until provider daemon types compilation errors are fixed.

package daemon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageRecordTypeString(t *testing.T) {
	assert.Equal(t, "periodic", UsageRecordTypePeriodic.String())
	assert.Equal(t, "final", UsageRecordTypeFinal.String())
	assert.Contains(t, UsageRecordType(99).String(), "unknown")
}

func TestTimeWindowValidate(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name    string
		tw      TimeWindow
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid time window",
			tw: TimeWindow{
				StartTime:       now,
				EndTime:         now.Add(1 * time.Hour),
				DurationSeconds: 3600,
			},
			wantErr: false,
		},
		{
			name: "zero start time",
			tw: TimeWindow{
				EndTime:         now,
				DurationSeconds: 3600,
			},
			wantErr: true,
			errMsg:  "start_time is required",
		},
		{
			name: "zero end time",
			tw: TimeWindow{
				StartTime:       now,
				DurationSeconds: 3600,
			},
			wantErr: true,
			errMsg:  "end_time is required",
		},
		{
			name: "end before start",
			tw: TimeWindow{
				StartTime:       now,
				EndTime:         now.Add(-1 * time.Hour),
				DurationSeconds: 3600,
			},
			wantErr: true,
			errMsg:  "end_time cannot be before start_time",
		},
		{
			name: "duration mismatch",
			tw: TimeWindow{
				StartTime:       now,
				EndTime:         now.Add(1 * time.Hour),
				DurationSeconds: 1800, // Should be 3600
			},
			wantErr: true,
			errMsg:  "duration_seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tw.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResourceUsageValidate(t *testing.T) {
	tests := []struct {
		name    string
		ru      ResourceUsage
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid resource usage",
			ru: ResourceUsage{
				CPUMillicores:       1000,
				CPUSeconds:          3600,
				MemoryBytesAvg:      1024 * 1024 * 100,
				MemoryBytesMax:      1024 * 1024 * 200,
				StorageBytesUsed:    1024 * 1024 * 1024,
				NetworkIngressBytes: 1024 * 1024,
				NetworkEgressBytes:  1024 * 1024 * 2,
			},
			wantErr: false,
		},
		{
			name: "negative cpu millicores",
			ru: ResourceUsage{
				CPUMillicores: -100,
			},
			wantErr: true,
			errMsg:  "cpu_millicores cannot be negative",
		},
		{
			name: "max memory less than avg",
			ru: ResourceUsage{
				CPUMillicores:  1000,
				MemoryBytesAvg: 200,
				MemoryBytesMax: 100,
			},
			wantErr: true,
			errMsg:  "memory_bytes_max cannot be less than memory_bytes_avg",
		},
		{
			name: "negative network ingress",
			ru: ResourceUsage{
				CPUMillicores:       1000,
				NetworkIngressBytes: -100,
			},
			wantErr: true,
			errMsg:  "network_ingress_bytes cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ru.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPricingCalcInputsValidate(t *testing.T) {
	tests := []struct {
		name    string
		pci     PricingCalcInputs
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pricing inputs",
			pci: PricingCalcInputs{
				PricingModel:    "hourly",
				BaseRatePerHour: "0.10",
				Currency:        "uve",
			},
			wantErr: false,
		},
		{
			name: "missing pricing model",
			pci: PricingCalcInputs{
				Currency: "uve",
			},
			wantErr: true,
			errMsg:  "pricing_model is required",
		},
		{
			name: "missing currency",
			pci: PricingCalcInputs{
				PricingModel: "hourly",
			},
			wantErr: true,
			errMsg:  "currency is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pci.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUsageRecordValidate(t *testing.T) {
	now := time.Now().UTC()
	validTimeWindow := TimeWindow{
		StartTime:       now,
		EndTime:         now.Add(1 * time.Hour),
		DurationSeconds: 3600,
	}
	validResourceUsage := ResourceUsage{
		CPUMillicores:  1000,
		MemoryBytesAvg: 100,
		MemoryBytesMax: 200,
	}
	validPricing := PricingCalcInputs{
		PricingModel: "hourly",
		Currency:     "uve",
	}
	validSignature := DaemonSignature{
		PublicKey: "aabbccdd",
		Signature: "11223344",
		Algorithm: "ed25519",
		KeyID:     "key-1",
		SignedAt:  now,
	}

	tests := []struct {
		name    string
		record  UsageRecord
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid usage record",
			record: UsageRecord{
				RecordID:          "record-1",
				OrderID:           "order-1",
				AllocationID:      "alloc-1",
				ProviderAddress:   "provider1",
				RecordType:        UsageRecordTypePeriodic,
				TimeWindow:        validTimeWindow,
				ResourceUsage:     validResourceUsage,
				PricingCalcInputs: validPricing,
				Signature:         validSignature,
				CreatedAt:         now,
			},
			wantErr: false,
		},
		{
			name: "missing record id",
			record: UsageRecord{
				OrderID:           "order-1",
				AllocationID:      "alloc-1",
				ProviderAddress:   "provider1",
				RecordType:        UsageRecordTypePeriodic,
				TimeWindow:        validTimeWindow,
				ResourceUsage:     validResourceUsage,
				PricingCalcInputs: validPricing,
				Signature:         validSignature,
			},
			wantErr: true,
			errMsg:  "record_id is required",
		},
		{
			name: "invalid record type",
			record: UsageRecord{
				RecordID:          "record-1",
				OrderID:           "order-1",
				AllocationID:      "alloc-1",
				ProviderAddress:   "provider1",
				RecordType:        UsageRecordType(99),
				TimeWindow:        validTimeWindow,
				ResourceUsage:     validResourceUsage,
				PricingCalcInputs: validPricing,
				Signature:         validSignature,
			},
			wantErr: true,
			errMsg:  "invalid record_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUsageRecordHash(t *testing.T) {
	now := time.Now().UTC()
	record := UsageRecord{
		RecordID:        "record-1",
		OrderID:         "order-1",
		AllocationID:    "alloc-1",
		ProviderAddress: "provider1",
		RecordType:      UsageRecordTypePeriodic,
		TimeWindow: TimeWindow{
			StartTime:       now,
			EndTime:         now.Add(1 * time.Hour),
			DurationSeconds: 3600,
		},
		ResourceUsage: ResourceUsage{
			CPUMillicores: 1000,
		},
		PricingCalcInputs: PricingCalcInputs{
			PricingModel: "hourly",
			Currency:     "uve",
		},
	}

	hash1, err := record.Hash()
	require.NoError(t, err)
	require.Len(t, hash1, 32) // SHA-256

	// Same record should produce same hash
	hash2, err := record.Hash()
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2)

	// Different record should produce different hash
	record.RecordID = "record-2"
	hash3, err := record.Hash()
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)
}
