package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pkg.akt.dev/node/x/veid/types"
)

// ============================================================================
// SMS Rate Limit Tests (VE-910: Rate Limiting)
// ============================================================================

func TestSMSRateLimitConfig_Default(t *testing.T) {
	config := types.DefaultSMSRateLimitConfig()

	assert.Equal(t, types.DefaultAccountRateLimitPerHour, config.AccountLimitPerHour)
	assert.Equal(t, types.DefaultAccountRateLimitPerDay, config.AccountLimitPerDay)
	assert.Equal(t, types.DefaultPhoneRateLimitPerHour, config.PhoneLimitPerHour)
	assert.Equal(t, types.DefaultPhoneRateLimitPerDay, config.PhoneLimitPerDay)
	assert.Equal(t, types.DefaultGlobalRateLimitPerMinute, config.GlobalLimitPerMinute)
	assert.Equal(t, types.DefaultCooldownSeconds, config.CooldownSeconds)
	assert.Equal(t, types.DefaultBlockDurationHours, config.BlockDurationHours)
	assert.True(t, config.EnableIPRateLimit)
}

func TestSMSRateLimitConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  types.SMSRateLimitConfig
		wantErr bool
	}{
		{
			name:    "default config",
			config:  types.DefaultSMSRateLimitConfig(),
			wantErr: false,
		},
		{
			name: "zero account hourly limit",
			config: types.SMSRateLimitConfig{
				AccountLimitPerHour:  0,
				AccountLimitPerDay:   10,
				PhoneLimitPerHour:    3,
				PhoneLimitPerDay:     5,
				GlobalLimitPerMinute: 100,
				CooldownSeconds:      60,
				BlockDurationHours:   24,
			},
			wantErr: true,
		},
		{
			name: "hourly greater than daily",
			config: types.SMSRateLimitConfig{
				AccountLimitPerHour:  15,
				AccountLimitPerDay:   10,
				PhoneLimitPerHour:    3,
				PhoneLimitPerDay:     5,
				GlobalLimitPerMinute: 100,
				CooldownSeconds:      60,
				BlockDurationHours:   24,
			},
			wantErr: true,
		},
		{
			name: "zero phone hourly limit",
			config: types.SMSRateLimitConfig{
				AccountLimitPerHour:  5,
				AccountLimitPerDay:   10,
				PhoneLimitPerHour:    0,
				PhoneLimitPerDay:     5,
				GlobalLimitPerMinute: 100,
				CooldownSeconds:      60,
				BlockDurationHours:   24,
			},
			wantErr: true,
		},
		{
			name: "phone hourly greater than daily",
			config: types.SMSRateLimitConfig{
				AccountLimitPerHour:  5,
				AccountLimitPerDay:   10,
				PhoneLimitPerHour:    10,
				PhoneLimitPerDay:     5,
				GlobalLimitPerMinute: 100,
				CooldownSeconds:      60,
				BlockDurationHours:   24,
			},
			wantErr: true,
		},
		{
			name: "zero global limit",
			config: types.SMSRateLimitConfig{
				AccountLimitPerHour:  5,
				AccountLimitPerDay:   10,
				PhoneLimitPerHour:    3,
				PhoneLimitPerDay:     5,
				GlobalLimitPerMinute: 0,
				CooldownSeconds:      60,
				BlockDurationHours:   24,
			},
			wantErr: true,
		},
		{
			name: "cooldown too short",
			config: types.SMSRateLimitConfig{
				AccountLimitPerHour:  5,
				AccountLimitPerDay:   10,
				PhoneLimitPerHour:    3,
				PhoneLimitPerDay:     5,
				GlobalLimitPerMinute: 100,
				CooldownSeconds:      10,
				BlockDurationHours:   24,
			},
			wantErr: true,
		},
		{
			name: "block duration too short",
			config: types.SMSRateLimitConfig{
				AccountLimitPerHour:  5,
				AccountLimitPerDay:   10,
				PhoneLimitPerHour:    3,
				PhoneLimitPerDay:     5,
				GlobalLimitPerMinute: 100,
				CooldownSeconds:      60,
				BlockDurationHours:   0,
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

func TestSMSRateLimitState_Creation(t *testing.T) {
	now := time.Now()
	state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "hash123", now)

	require.NotNil(t, state)
	assert.Equal(t, types.RateLimitTypeAccount, state.EntityType)
	assert.Equal(t, "hash123", state.EntityHash)
	assert.Equal(t, uint32(0), state.HourlyCount)
	assert.Equal(t, uint32(0), state.DailyCount)
	assert.Equal(t, uint64(0), state.TotalRequests)
	assert.False(t, state.IsBlocked)
}

func TestSMSRateLimitState_RecordRequest(t *testing.T) {
	now := time.Now()
	state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "hash123", now)

	// Record requests
	state.RecordRequest(now)
	assert.Equal(t, uint32(1), state.HourlyCount)
	assert.Equal(t, uint32(1), state.DailyCount)
	assert.Equal(t, uint64(1), state.TotalRequests)

	state.RecordRequest(now.Add(time.Minute))
	assert.Equal(t, uint32(2), state.HourlyCount)
	assert.Equal(t, uint32(2), state.DailyCount)
	assert.Equal(t, uint64(2), state.TotalRequests)
}

func TestSMSRateLimitState_WindowReset(t *testing.T) {
	now := time.Now()
	state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "hash123", now)

	// Record some requests
	state.RecordRequest(now)
	state.RecordRequest(now)
	assert.Equal(t, uint32(2), state.HourlyCount)
	assert.Equal(t, uint32(2), state.DailyCount)

	// After 1 hour, hourly should reset
	oneHourLater := now.Add(time.Hour + time.Minute)
	state.ResetHourlyWindow(oneHourLater)
	assert.Equal(t, uint32(0), state.HourlyCount)
	assert.Equal(t, uint32(2), state.DailyCount) // Daily not reset

	// After 24 hours, daily should reset
	oneDayLater := now.Add(25 * time.Hour)
	state.ResetDailyWindow(oneDayLater)
	assert.Equal(t, uint32(0), state.DailyCount)
}

func TestSMSRateLimitState_Cooldown(t *testing.T) {
	now := time.Now()
	state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "hash123", now)

	cooldownSeconds := int64(60)

	// No request yet - cooldown passed
	assert.True(t, state.CheckCooldown(now, cooldownSeconds))

	// Record a request
	state.RecordRequest(now)

	// Immediately after - cooldown not passed
	assert.False(t, state.CheckCooldown(now, cooldownSeconds))
	assert.False(t, state.CheckCooldown(now.Add(30*time.Second), cooldownSeconds))

	// After cooldown - passed
	assert.True(t, state.CheckCooldown(now.Add(61*time.Second), cooldownSeconds))
}

func TestSMSRateLimitState_Blocking(t *testing.T) {
	now := time.Now()
	state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "hash123", now)

	// Not blocked initially
	assert.False(t, state.IsBlocked)
	assert.True(t, state.IsBlockExpired(now))

	// Block for 24 hours
	state.Block(now, 24, "rate limit exceeded")
	assert.True(t, state.IsBlocked)
	assert.Equal(t, "rate limit exceeded", state.BlockReason)
	assert.Equal(t, uint32(1), state.TotalBlocks)
	assert.False(t, state.IsBlockExpired(now))
	assert.False(t, state.IsBlockExpired(now.Add(23*time.Hour)))

	// After block period
	assert.True(t, state.IsBlockExpired(now.Add(25*time.Hour)))

	// Unblock
	state.Unblock()
	assert.False(t, state.IsBlocked)
	assert.Empty(t, state.BlockReason)
}

func TestSMSRateLimiter_CheckAccountLimit(t *testing.T) {
	config := types.DefaultSMSRateLimitConfig()
	limiter := types.NewSMSRateLimiter(config)
	now := time.Now()

	t.Run("allows first request", func(t *testing.T) {
		state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "account1", now)
		result := limiter.CheckAccountLimit(state, now)

		assert.True(t, result.Allowed)
		assert.Equal(t, config.AccountLimitPerHour-1, result.RemainingHourly)
		assert.Equal(t, config.AccountLimitPerDay-1, result.RemainingDaily)
	})

	t.Run("enforces cooldown", func(t *testing.T) {
		state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "account2", now)
		state.RecordRequest(now)

		// Try again immediately
		result := limiter.CheckAccountLimit(state, now.Add(10*time.Second))

		assert.False(t, result.Allowed)
		assert.Equal(t, types.RateLimitTypeAccount, result.LimitType)
		assert.Contains(t, result.LimitReason, "cooldown")
		assert.Greater(t, result.RetryAfter, int64(0))
	})

	t.Run("enforces hourly limit", func(t *testing.T) {
		state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "account3", now)

		// Exhaust hourly limit
		for i := uint32(0); i < config.AccountLimitPerHour; i++ {
			state.RecordRequest(now.Add(time.Duration(i) * 2 * time.Minute)) // Spread out to avoid cooldown
		}

		result := limiter.CheckAccountLimit(state, now.Add(time.Duration(config.AccountLimitPerHour)*2*time.Minute))

		assert.False(t, result.Allowed)
		assert.Equal(t, types.RateLimitTypeAccount, result.LimitType)
		assert.Contains(t, result.LimitReason, "hourly")
		assert.Equal(t, uint32(0), result.RemainingHourly)
	})

	t.Run("enforces daily limit", func(t *testing.T) {
		state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "account4", now)

		// Exhaust daily limit over multiple hours
		for i := uint32(0); i < config.AccountLimitPerDay; i++ {
			requestTime := now.Add(time.Duration(i) * 10 * time.Minute)
			state.ResetHourlyWindow(requestTime) // Simulate different hours
			state.RecordRequest(requestTime)
		}

		result := limiter.CheckAccountLimit(state, now.Add(time.Duration(config.AccountLimitPerDay)*10*time.Minute+time.Minute))

		assert.False(t, result.Allowed)
		assert.Equal(t, types.RateLimitTypeAccount, result.LimitType)
		assert.Contains(t, result.LimitReason, "daily")
		assert.Equal(t, uint32(0), result.RemainingDaily)
	})

	t.Run("respects blocking", func(t *testing.T) {
		state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "account5", now)
		state.Block(now, 24, "suspicious activity")

		result := limiter.CheckAccountLimit(state, now.Add(time.Hour))

		assert.False(t, result.Allowed)
		assert.Contains(t, result.LimitReason, "blocked")
		assert.Greater(t, result.RetryAfter, int64(0))
	})

	t.Run("unblocks after expiry", func(t *testing.T) {
		state := types.NewSMSRateLimitState(types.RateLimitTypeAccount, "account6", now)
		state.Block(now, 24, "suspicious activity")

		// Check after block expired
		result := limiter.CheckAccountLimit(state, now.Add(25*time.Hour))

		assert.True(t, result.Allowed)
		assert.False(t, state.IsBlocked)
	})
}

func TestSMSRateLimiter_CheckPhoneLimit(t *testing.T) {
	config := types.DefaultSMSRateLimitConfig()
	limiter := types.NewSMSRateLimiter(config)
	now := time.Now()

	t.Run("allows first request", func(t *testing.T) {
		state := types.NewSMSRateLimitState(types.RateLimitTypePhone, "phone1", now)
		result := limiter.CheckPhoneLimit(state, now)

		assert.True(t, result.Allowed)
		assert.Equal(t, config.PhoneLimitPerHour-1, result.RemainingHourly)
		assert.Equal(t, config.PhoneLimitPerDay-1, result.RemainingDaily)
	})

	t.Run("enforces hourly limit", func(t *testing.T) {
		state := types.NewSMSRateLimitState(types.RateLimitTypePhone, "phone2", now)

		// Exhaust hourly limit
		for i := uint32(0); i < config.PhoneLimitPerHour; i++ {
			state.RecordRequest(now.Add(time.Duration(i) * time.Minute))
		}

		result := limiter.CheckPhoneLimit(state, now.Add(time.Duration(config.PhoneLimitPerHour)*time.Minute))

		assert.False(t, result.Allowed)
		assert.Equal(t, types.RateLimitTypePhone, result.LimitType)
		assert.Contains(t, result.LimitReason, "hourly")
	})

	t.Run("enforces daily limit", func(t *testing.T) {
		state := types.NewSMSRateLimitState(types.RateLimitTypePhone, "phone3", now)

		// Exhaust daily limit over multiple hours
		for i := uint32(0); i < config.PhoneLimitPerDay; i++ {
			requestTime := now.Add(time.Duration(i) * 30 * time.Minute)
			state.ResetHourlyWindow(requestTime)
			state.RecordRequest(requestTime)
		}

		result := limiter.CheckPhoneLimit(state, now.Add(time.Duration(config.PhoneLimitPerDay)*30*time.Minute+time.Minute))

		assert.False(t, result.Allowed)
		assert.Equal(t, types.RateLimitTypePhone, result.LimitType)
		assert.Contains(t, result.LimitReason, "daily")
	})
}

func TestGlobalRateLimitState(t *testing.T) {
	now := time.Now()
	state := types.NewGlobalRateLimitState(now)

	require.NotNil(t, state)
	assert.Equal(t, uint32(0), state.MinuteCount)
	assert.Equal(t, uint64(0), state.TotalRequests)

	// Record requests
	state.RecordRequest(now)
	assert.Equal(t, uint32(1), state.MinuteCount)
	assert.Equal(t, uint64(1), state.TotalRequests)

	// Reset after minute
	state.ResetMinuteWindow(now.Add(61 * time.Second))
	assert.Equal(t, uint32(0), state.MinuteCount)
	assert.Equal(t, uint64(1), state.TotalRequests) // Total persists
}

func TestSMSRateLimiter_CheckGlobalLimit(t *testing.T) {
	config := types.DefaultSMSRateLimitConfig()
	limiter := types.NewSMSRateLimiter(config)
	now := time.Now()

	t.Run("allows requests within limit", func(t *testing.T) {
		state := types.NewGlobalRateLimitState(now)
		result := limiter.CheckGlobalLimit(state, now)

		assert.True(t, result.Allowed)
	})

	t.Run("enforces global limit", func(t *testing.T) {
		state := types.NewGlobalRateLimitState(now)

		// Exhaust global limit
		for i := uint32(0); i < config.GlobalLimitPerMinute; i++ {
			state.RecordRequest(now)
		}

		result := limiter.CheckGlobalLimit(state, now)

		assert.False(t, result.Allowed)
		assert.Equal(t, types.RateLimitTypeGlobal, result.LimitType)
		assert.Contains(t, result.LimitReason, "global")
		assert.Greater(t, result.RetryAfter, int64(0))
	})

	t.Run("resets after minute", func(t *testing.T) {
		state := types.NewGlobalRateLimitState(now)

		// Exhaust limit
		for i := uint32(0); i < config.GlobalLimitPerMinute; i++ {
			state.RecordRequest(now)
		}

		// After a minute
		result := limiter.CheckGlobalLimit(state, now.Add(61*time.Second))

		assert.True(t, result.Allowed)
	})
}

func TestRateLimitCheckResult_Fields(t *testing.T) {
	result := types.RateLimitCheckResult{
		Allowed:         false,
		LimitType:       types.RateLimitTypeAccount,
		LimitReason:     "test reason",
		RetryAfter:      300,
		RemainingHourly: 2,
		RemainingDaily:  5,
	}

	assert.False(t, result.Allowed)
	assert.Equal(t, types.RateLimitTypeAccount, result.LimitType)
	assert.Equal(t, "test reason", result.LimitReason)
	assert.Equal(t, int64(300), result.RetryAfter)
	assert.Equal(t, uint32(2), result.RemainingHourly)
	assert.Equal(t, uint32(5), result.RemainingDaily)
}

func TestRateLimitType_Values(t *testing.T) {
	types := []types.RateLimitType{
		types.RateLimitTypeAccount,
		types.RateLimitTypePhone,
		types.RateLimitTypeIP,
		types.RateLimitTypeGlobal,
	}

	// Ensure all types are unique
	seen := make(map[types.RateLimitType]bool)
	for _, rt := range types {
		assert.False(t, seen[rt], "duplicate rate limit type: %s", rt)
		seen[rt] = true
	}
}
