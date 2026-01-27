//go:build ignore
// +build ignore

// TODO: This test file is excluded until CarrierType API is stabilized.

package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// VoIP Detection Tests (VE-910: VoIP Detection)
// ============================================================================

func TestCarrierType_Validation(t *testing.T) {
	tests := []struct {
		name    string
		carrier types.CarrierType
		valid   bool
	}{
		{"mobile", types.CarrierTypeMobile, true},
		{"landline", types.CarrierTypeLandline, true},
		{"voip", types.CarrierTypeVoIP, true},
		{"toll_free", types.CarrierTypeTollFree, true},
		{"premium", types.CarrierTypePremium, true},
		{"unknown", types.CarrierTypeUnknown, true},
		{"invalid", types.CarrierType("invalid"), false},
		{"empty", types.CarrierType(""), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsValidCarrierType(tc.carrier)
			assert.Equal(t, tc.valid, result)
		})
	}
}

func TestCarrierType_Blocked(t *testing.T) {
	tests := []struct {
		name     string
		carrier  types.CarrierType
		expected bool
	}{
		{"mobile - not blocked", types.CarrierTypeMobile, false},
		{"landline - not blocked", types.CarrierTypeLandline, false},
		{"voip - blocked", types.CarrierTypeVoIP, true},
		{"toll_free - blocked", types.CarrierTypeTollFree, true},
		{"premium - blocked", types.CarrierTypePremium, true},
		{"unknown - not blocked", types.CarrierTypeUnknown, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IsBlockedCarrierType(tc.carrier)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCarrierLookupResult_ShouldBlock(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		result    *types.CarrierLookupResult
		blockVoIP bool
		expected  bool
	}{
		{
			name: "mobile - no block",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash1",
				CarrierType:     types.CarrierTypeMobile,
				IsVoIP:          false,
				RiskLevel:       types.VoIPRiskLow,
				RiskScore:       10,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			expected:  false,
		},
		{
			name: "voip with blocking enabled",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash2",
				CarrierType:     types.CarrierTypeVoIP,
				IsVoIP:          true,
				RiskLevel:       types.VoIPRiskMedium,
				RiskScore:       50,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			expected:  true,
		},
		{
			name: "voip with blocking disabled",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash3",
				CarrierType:     types.CarrierTypeVoIP,
				IsVoIP:          true,
				RiskLevel:       types.VoIPRiskMedium,
				RiskScore:       50,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: false,
			expected:  true, // Still blocked because CarrierType is VoIP
		},
		{
			name: "toll free - blocked",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash4",
				CarrierType:     types.CarrierTypeTollFree,
				IsVoIP:          false,
				RiskLevel:       types.VoIPRiskLow,
				RiskScore:       20,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			expected:  true,
		},
		{
			name: "premium - blocked",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash5",
				CarrierType:     types.CarrierTypePremium,
				IsVoIP:          false,
				RiskLevel:       types.VoIPRiskLow,
				RiskScore:       30,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			expected:  true,
		},
		{
			name: "critical risk level",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash6",
				CarrierType:     types.CarrierTypeMobile,
				IsVoIP:          false,
				RiskLevel:       types.VoIPRiskCritical,
				RiskScore:       95,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			expected:  true,
		},
		{
			name: "high risk score",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash7",
				CarrierType:     types.CarrierTypeMobile,
				IsVoIP:          false,
				RiskLevel:       types.VoIPRiskHigh,
				RiskScore:       85,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			expected:  true,
		},
		{
			name: "failed lookup - not blocked",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash8",
				CarrierType:     types.CarrierTypeUnknown,
				LookupTimestamp: now,
				Success:         false,
				ErrorCode:       "LOOKUP_FAILED",
			},
			blockVoIP: true,
			expected:  false,
		},
		{
			name: "landline - not blocked",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash9",
				CarrierType:     types.CarrierTypeLandline,
				IsVoIP:          false,
				RiskLevel:       types.VoIPRiskLow,
				RiskScore:       5,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			expected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.result.ShouldBlock(tc.blockVoIP)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCarrierLookupResult_GetBlockReason(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		result    *types.CarrierLookupResult
		blockVoIP bool
		contains  string
		empty     bool
	}{
		{
			name: "voip blocked",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash1",
				CarrierType:     types.CarrierTypeMobile,
				IsVoIP:          true,
				RiskLevel:       types.VoIPRiskMedium,
				RiskScore:       50,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			contains:  "VoIP",
		},
		{
			name: "toll free blocked",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash2",
				CarrierType:     types.CarrierTypeTollFree,
				IsVoIP:          false,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			contains:  "toll_free",
		},
		{
			name: "critical risk",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash3",
				CarrierType:     types.CarrierTypeMobile,
				IsVoIP:          false,
				RiskLevel:       types.VoIPRiskCritical,
				RiskScore:       95,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			contains:  "fraud",
		},
		{
			name: "high score",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash4",
				CarrierType:     types.CarrierTypeMobile,
				IsVoIP:          false,
				RiskLevel:       types.VoIPRiskHigh,
				RiskScore:       85,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			contains:  "risk",
		},
		{
			name: "not blocked",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash5",
				CarrierType:     types.CarrierTypeMobile,
				IsVoIP:          false,
				RiskLevel:       types.VoIPRiskLow,
				RiskScore:       10,
				LookupTimestamp: now,
				Success:         true,
			},
			blockVoIP: true,
			empty:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reason := tc.result.GetBlockReason(tc.blockVoIP)
			if tc.empty {
				assert.Empty(t, reason)
			} else {
				assert.Contains(t, reason, tc.contains)
			}
		})
	}
}

func TestCarrierLookupResult_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		result  *types.CarrierLookupResult
		wantErr bool
	}{
		{
			name: "valid result",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash123",
				CarrierType:     types.CarrierTypeMobile,
				LookupTimestamp: now,
			},
			wantErr: false,
		},
		{
			name: "empty phone hash ref",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "",
				CarrierType:     types.CarrierTypeMobile,
				LookupTimestamp: now,
			},
			wantErr: true,
		},
		{
			name: "zero timestamp",
			result: &types.CarrierLookupResult{
				PhoneHashRef: "hash123",
				CarrierType:  types.CarrierTypeMobile,
			},
			wantErr: true,
		},
		{
			name: "invalid carrier type",
			result: &types.CarrierLookupResult{
				PhoneHashRef:    "hash123",
				CarrierType:     types.CarrierType("invalid"),
				LookupTimestamp: now,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.result.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVoIPDetectorConfig_Default(t *testing.T) {
	config := types.DefaultVoIPDetectorConfig()

	assert.Equal(t, "numverify", config.Provider)
	assert.True(t, config.Enabled)
	assert.True(t, config.BlockVoIP)
	assert.Equal(t, int64(86400), config.CacheTTLSeconds)
	assert.Equal(t, uint32(60), config.RateLimitPerMinute)
	assert.Equal(t, uint32(60), config.HighRiskThreshold)
	assert.Equal(t, uint32(80), config.BlockThreshold)
}

func TestVoIPDetectorConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  types.VoIPDetectorConfig
		wantErr bool
	}{
		{
			name:    "default config",
			config:  types.DefaultVoIPDetectorConfig(),
			wantErr: false,
		},
		{
			name: "enabled with empty provider",
			config: types.VoIPDetectorConfig{
				Enabled:            true,
				Provider:           "",
				CacheTTLSeconds:    3600,
				RateLimitPerMinute: 60,
				HighRiskThreshold:  60,
				BlockThreshold:     80,
			},
			wantErr: true,
		},
		{
			name: "disabled with empty provider - ok",
			config: types.VoIPDetectorConfig{
				Enabled:            false,
				Provider:           "",
				CacheTTLSeconds:    3600,
				RateLimitPerMinute: 60,
				HighRiskThreshold:  60,
				BlockThreshold:     80,
			},
			wantErr: false,
		},
		{
			name: "negative cache TTL",
			config: types.VoIPDetectorConfig{
				Enabled:            true,
				Provider:           "twilio",
				CacheTTLSeconds:    -1,
				RateLimitPerMinute: 60,
				HighRiskThreshold:  60,
				BlockThreshold:     80,
			},
			wantErr: true,
		},
		{
			name: "zero rate limit",
			config: types.VoIPDetectorConfig{
				Enabled:            true,
				Provider:           "twilio",
				CacheTTLSeconds:    3600,
				RateLimitPerMinute: 0,
				HighRiskThreshold:  60,
				BlockThreshold:     80,
			},
			wantErr: true,
		},
		{
			name: "block threshold less than high risk",
			config: types.VoIPDetectorConfig{
				Enabled:            true,
				Provider:           "twilio",
				CacheTTLSeconds:    3600,
				RateLimitPerMinute: 60,
				HighRiskThreshold:  80,
				BlockThreshold:     60,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVoIPDetectorConfig_IsCountryAllowed(t *testing.T) {
	tests := []struct {
		name     string
		config   types.VoIPDetectorConfig
		country  string
		expected bool
	}{
		{
			name: "empty lists - all allowed",
			config: types.VoIPDetectorConfig{
				AllowedCountries: []string{},
				BlockedCountries: []string{},
			},
			country:  "US",
			expected: true,
		},
		{
			name: "in allowed list",
			config: types.VoIPDetectorConfig{
				AllowedCountries: []string{"US", "GB", "CA"},
				BlockedCountries: []string{},
			},
			country:  "US",
			expected: true,
		},
		{
			name: "not in allowed list",
			config: types.VoIPDetectorConfig{
				AllowedCountries: []string{"US", "GB", "CA"},
				BlockedCountries: []string{},
			},
			country:  "RU",
			expected: false,
		},
		{
			name: "in blocked list",
			config: types.VoIPDetectorConfig{
				AllowedCountries: []string{},
				BlockedCountries: []string{"RU", "KP"},
			},
			country:  "RU",
			expected: false,
		},
		{
			name: "not in blocked list",
			config: types.VoIPDetectorConfig{
				AllowedCountries: []string{},
				BlockedCountries: []string{"RU", "KP"},
			},
			country:  "US",
			expected: true,
		},
		{
			name: "in both lists - blocked wins",
			config: types.VoIPDetectorConfig{
				AllowedCountries: []string{"US", "RU"},
				BlockedCountries: []string{"RU"},
			},
			country:  "RU",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.config.IsCountryAllowed(tc.country)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMockVoIPDetector(t *testing.T) {
	detector := types.NewMockVoIPDetector()

	require.NotNil(t, detector)
	assert.Equal(t, "mock", detector.GetProviderName())
	assert.True(t, detector.IsAvailable())
	assert.Equal(t, uint32(100), detector.GetRateLimit())
	assert.Equal(t, int64(3600), detector.GetCacheTTL())
}

func TestMockVoIPDetector_LookupCarrier(t *testing.T) {
	detector := types.NewMockVoIPDetector()

	t.Run("default result", func(t *testing.T) {
		result, err := detector.LookupCarrier("+14155551234", "hash123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "hash123", result.PhoneHashRef)
		assert.Equal(t, types.CarrierTypeMobile, result.CarrierType)
		assert.False(t, result.IsVoIP)
		assert.Equal(t, types.VoIPRiskLow, result.RiskLevel)
		assert.True(t, result.Success)
	})

	t.Run("custom result", func(t *testing.T) {
		customResult := &types.CarrierLookupResult{
			PhoneHashRef:    "hash456",
			CarrierType:     types.CarrierTypeVoIP,
			IsVoIP:          true,
			RiskLevel:       types.VoIPRiskHigh,
			RiskScore:       85,
			LookupTimestamp: time.Now(),
			Success:         true,
		}
		detector.SetResult("hash456", customResult)

		result, err := detector.LookupCarrier("+19999999999", "hash456")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, types.CarrierTypeVoIP, result.CarrierType)
		assert.True(t, result.IsVoIP)
		assert.Equal(t, types.VoIPRiskHigh, result.RiskLevel)
		assert.Equal(t, uint32(85), result.RiskScore)
	})
}

func TestMockVoIPDetector_NeverStoresPhone(t *testing.T) {
	detector := types.NewMockVoIPDetector()
	phone := "+14155551234"

	result, err := detector.LookupCarrier(phone, "hash123")
	require.NoError(t, err)

	// Ensure phone number is not stored in result
	assert.NotContains(t, result.PhoneHashRef, phone)
	assert.NotEqual(t, result.PhoneHashRef, phone)
}

func TestVoIPRiskLevel_Values(t *testing.T) {
	levels := []types.VoIPRiskLevel{
		types.VoIPRiskLow,
		types.VoIPRiskMedium,
		types.VoIPRiskHigh,
		types.VoIPRiskCritical,
	}

	// Ensure all levels are unique
	seen := make(map[types.VoIPRiskLevel]bool)
	for _, level := range levels {
		assert.False(t, seen[level], "duplicate risk level: %s", level)
		seen[level] = true
	}
}

func TestAllCarrierTypes(t *testing.T) {
	types := types.AllCarrierTypes()

	assert.Len(t, types, 6)
	assert.Contains(t, types, types.CarrierTypeMobile)
	assert.Contains(t, types, types.CarrierTypeLandline)
	assert.Contains(t, types, types.CarrierTypeVoIP)
	assert.Contains(t, types, types.CarrierTypeTollFree)
	assert.Contains(t, types, types.CarrierTypePremium)
	assert.Contains(t, types, types.CarrierTypeUnknown)
}
