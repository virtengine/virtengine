package types

import (
	"bytes"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

// GenesisState is the genesis state for the settlement module
type GenesisState struct {
	// Params are the module parameters
	Params Params `json:"params"`

	// EscrowAccounts are the initial escrow accounts
	EscrowAccounts []EscrowAccount `json:"escrow_accounts"`

	// SettlementRecords are the initial settlement records
	SettlementRecords []SettlementRecord `json:"settlement_records"`

	// RewardDistributions are the initial reward distributions
	RewardDistributions []RewardDistribution `json:"reward_distributions"`

	// UsageRecords are the initial usage records
	UsageRecords []UsageRecord `json:"usage_records"`

	// ClaimableRewards are the initial claimable rewards
	ClaimableRewards []ClaimableRewards `json:"claimable_rewards"`

	// PayoutRecords are the initial payout records
	PayoutRecords []PayoutRecord `json:"payout_records"`

	// FiatConversionRecords are the initial fiat conversion records
	FiatConversionRecords []FiatConversionRecord `json:"fiat_conversion_records"`

	// FiatPayoutPreferences are the initial provider fiat payout preferences
	FiatPayoutPreferences []FiatPayoutPreference `json:"fiat_payout_preferences"`

	// EscrowSequence is the next escrow sequence number
	EscrowSequence uint64 `json:"escrow_sequence"`

	// SettlementSequence is the next settlement sequence number
	SettlementSequence uint64 `json:"settlement_sequence"`

	// DistributionSequence is the next distribution sequence number
	DistributionSequence uint64 `json:"distribution_sequence"`

	// UsageSequence is the next usage sequence number
	UsageSequence uint64 `json:"usage_sequence"`

	// PayoutSequence is the next payout sequence number
	PayoutSequence uint64 `json:"payout_sequence"`

	// FiatConversionSequence is the next fiat conversion sequence number
	FiatConversionSequence uint64 `json:"fiat_conversion_sequence"`

	// UsageAuthenticationActive enables Task 84B on fresh chains. Old genesis
	// documents omit this field and remain migration inputs rather than being
	// retroactively authenticated.
	UsageAuthenticationActive bool `json:"usage_authentication_active"`

	// FinancialCases are canonical Task 84D financial dispute aggregates.
	FinancialCases []FinancialCase `json:"financial_cases"`
	// FinancialCasesActive enables v1.7.0 non-owner mutation fencing on fresh chains.
	FinancialCasesActive bool `json:"financial_cases_active"`

	// FiatConversionCustodyBalance is the exact bank balance expected in the
	// governed fiat custody sink at genesis import/export.
	FiatConversionCustodyBalance sdk.Coins `json:"fiat_conversion_custody_balance"`
}

// Params defines the parameters for the settlement module
type Params struct {
	// PlatformFeeRate is the platform fee rate (e.g., 0.05 for 5%)
	PlatformFeeRate string `json:"platform_fee_rate"`

	// ValidatorFeeRate is the validator fee rate (e.g., 0.01 for 1%)
	ValidatorFeeRate string `json:"validator_fee_rate"`

	// MinEscrowDuration is the minimum escrow duration in seconds
	MinEscrowDuration uint64 `json:"min_escrow_duration"`

	// MaxEscrowDuration is the maximum escrow duration in seconds
	MaxEscrowDuration uint64 `json:"max_escrow_duration"`

	// SettlementPeriod is the default settlement period in seconds
	SettlementPeriod uint64 `json:"settlement_period"`

	// RewardClaimExpiry is how long rewards can be claimed (in seconds)
	RewardClaimExpiry uint64 `json:"reward_claim_expiry"`

	// MinSettlementAmount is the minimum amount for a settlement
	MinSettlementAmount string `json:"min_settlement_amount"`

	// UsageGracePeriod is the grace period for usage disputes (in seconds)
	UsageGracePeriod uint64 `json:"usage_grace_period"`

	// StakingRewardEpochLength is the length of staking reward epochs in blocks
	StakingRewardEpochLength uint64 `json:"staking_reward_epoch_length"`

	// RewardPoolAddress is the address where staking rewards are aggregated
	RewardPoolAddress string `json:"reward_pool_address"`

	// VerificationRewardAmount is the base reward for identity verifications
	VerificationRewardAmount string `json:"verification_reward_amount"`

	// PayoutHoldbackRate is the holdback rate for payouts (e.g., 0.0 for no holdback)
	PayoutHoldbackRate string `json:"payout_holdback_rate"`

	// MaxPayoutRetries is the maximum number of retry attempts for failed payouts
	MaxPayoutRetries uint32 `json:"max_payout_retries"`

	// DisputeWindowDuration is the dispute window duration in seconds
	DisputeWindowDuration uint64 `json:"dispute_window_duration"`

	// UsageRewardRateBps is the base reward rate for usage rewards (basis points)
	UsageRewardRateBps uint32 `json:"usage_reward_rate_bps"`

	// UsageRewardCPUMultiplierBps is the CPU usage reward multiplier in basis points
	UsageRewardCPUMultiplierBps uint32 `json:"usage_reward_cpu_multiplier_bps"`

	// UsageRewardMemoryMultiplierBps is the memory usage reward multiplier in basis points
	UsageRewardMemoryMultiplierBps uint32 `json:"usage_reward_memory_multiplier_bps"`

	// UsageRewardStorageMultiplierBps is the storage usage reward multiplier in basis points
	UsageRewardStorageMultiplierBps uint32 `json:"usage_reward_storage_multiplier_bps"`

	// UsageRewardGPUMultiplierBps is the GPU usage reward multiplier in basis points
	UsageRewardGPUMultiplierBps uint32 `json:"usage_reward_gpu_multiplier_bps"`

	// UsageRewardNetworkMultiplierBps is the network usage reward multiplier in basis points
	UsageRewardNetworkMultiplierBps uint32 `json:"usage_reward_network_multiplier_bps"`

	// UsageRewardSLAOnTimeMultiplierBps is the on-time reporting SLA multiplier in basis points
	UsageRewardSLAOnTimeMultiplierBps uint32 `json:"usage_reward_sla_ontime_multiplier_bps"`

	// UsageRewardSLALateMultiplierBps is the late reporting SLA multiplier in basis points
	UsageRewardSLALateMultiplierBps uint32 `json:"usage_reward_sla_late_multiplier_bps"`

	// UsageRewardAcknowledgedMultiplierBps is the customer-acknowledged quality multiplier in basis points
	UsageRewardAcknowledgedMultiplierBps uint32 `json:"usage_reward_ack_multiplier_bps"`

	// UsageRewardUnacknowledgedMultiplierBps is the unacknowledged quality multiplier in basis points
	UsageRewardUnacknowledgedMultiplierBps uint32 `json:"usage_reward_unack_multiplier_bps"`

	// FiatConversionEnabled enables fiat conversion flow
	FiatConversionEnabled bool `json:"fiat_conversion_enabled"`

	// FiatConversionMinAmount is the minimum stablecoin amount eligible for conversion
	FiatConversionMinAmount string `json:"fiat_conversion_min_amount"`

	// FiatConversionMaxAmount is the maximum stablecoin amount eligible for conversion
	FiatConversionMaxAmount string `json:"fiat_conversion_max_amount"`

	// FiatConversionDailyLimit is the daily stablecoin conversion cap per provider
	FiatConversionDailyLimit string `json:"fiat_conversion_daily_limit"`

	// FiatConversionStableDenom is the stablecoin denom used for off-ramp
	FiatConversionStableDenom string `json:"fiat_conversion_stable_denom"`

	// FiatConversionStableSymbol is the stablecoin symbol used for swaps
	FiatConversionStableSymbol string `json:"fiat_conversion_stable_symbol"`

	// FiatConversionStableDecimals is the stablecoin decimals
	FiatConversionStableDecimals uint32 `json:"fiat_conversion_stable_decimals"`

	// FiatConversionDefaultFiat is the default fiat currency
	FiatConversionDefaultFiat string `json:"fiat_conversion_default_fiat"`

	// FiatConversionDefaultMethod is the default payment method
	FiatConversionDefaultMethod string `json:"fiat_conversion_default_method"`

	// FiatConversionMaxSlippage is the maximum slippage allowed (string decimal)
	FiatConversionMaxSlippage string `json:"fiat_conversion_max_slippage"`

	// FiatConversionRiskScoreThreshold is the compliance risk score threshold
	FiatConversionRiskScoreThreshold int32 `json:"fiat_conversion_risk_score_threshold"`

	// FiatConversionMinComplianceStatus is the minimum compliance status required
	FiatConversionMinComplianceStatus string `json:"fiat_conversion_min_compliance_status"`

	// FiatConversionSpreadBps is the spread applied to fiat conversion rates (basis points)
	FiatConversionSpreadBps uint32 `json:"fiat_conversion_spread_bps"`

	// OracleSources defines the configured oracle sources
	OracleSources []OracleSourceConfig `json:"oracle_sources"`

	// OracleStalenessThresholdSeconds defines staleness threshold in seconds
	OracleStalenessThresholdSeconds uint64 `json:"oracle_staleness_threshold_seconds"`

	// OracleMinSources defines minimum oracle sources required for aggregation
	OracleMinSources uint32 `json:"oracle_min_sources"`

	// OracleManualPrices defines governance-set emergency prices
	OracleManualPrices []ManualPriceOverride `json:"oracle_manual_prices"`

	// OracleDeviationThresholdBps defines deviation threshold for alerts (basis points)
	OracleDeviationThresholdBps uint32 `json:"oracle_deviation_threshold_bps"`

	// OracleDeviationWindowSeconds defines the alert evaluation window in seconds
	OracleDeviationWindowSeconds uint64 `json:"oracle_deviation_window_seconds"`

	FinancialCaseFilingWindowSeconds       uint64 `json:"financial_case_filing_window_seconds"`
	FinancialCaseEvidenceWindowSeconds     uint64 `json:"financial_case_evidence_window_seconds"`
	FinancialCaseReviewWindowSeconds       uint64 `json:"financial_case_review_window_seconds"`
	FinancialCaseAppealWindowSeconds       uint64 `json:"financial_case_appeal_window_seconds"`
	FinancialCaseEscalationWindowSeconds   uint64 `json:"financial_case_escalation_window_seconds"`
	FinancialCaseFilingWindowBlocks        int64  `json:"financial_case_filing_window_blocks"`
	FinancialCaseEvidenceWindowBlocks      int64  `json:"financial_case_evidence_window_blocks"`
	FinancialCaseReviewWindowBlocks        int64  `json:"financial_case_review_window_blocks"`
	FinancialCaseAppealWindowBlocks        int64  `json:"financial_case_appeal_window_blocks"`
	FinancialCaseEscalationWindowBlocks    int64  `json:"financial_case_escalation_window_blocks"`
	FinancialCaseMaxClaims                 uint32 `json:"financial_case_max_claims"`
	FinancialCaseMaxAppeals                uint32 `json:"financial_case_max_appeals"`
	FinancialCaseMaxEvidenceReferenceBytes uint32 `json:"financial_case_max_evidence_reference_bytes"`
	FinancialCaseTimeoutBatchLimit         uint32 `json:"financial_case_timeout_batch_limit"`

	// FiatConversionDEXProfileID and digest commit the exact production DEX
	// route profile selected by governance.
	FiatConversionDEXProfileID     string                                  `json:"fiat_conversion_dex_profile_id"`
	FiatConversionDEXProfileDigest []byte                                  `json:"fiat_conversion_dex_profile_digest"`
	FiatConversionDEXProfileState  settlementv1.FiatConversionProfileState `json:"fiat_conversion_dex_profile_state"`
	// FiatConversionPayoutProfileID and digest commit the exact production
	// payout corridor profile selected by governance.
	FiatConversionPayoutProfileID     string                                  `json:"fiat_conversion_payout_profile_id"`
	FiatConversionPayoutProfileDigest []byte                                  `json:"fiat_conversion_payout_profile_digest"`
	FiatConversionPayoutProfileState  settlementv1.FiatConversionProfileState `json:"fiat_conversion_payout_profile_state"`
	// FiatConversionMinSwapFinalityConfirmations is external-chain finality
	// evidence required before stable output can be accepted.
	FiatConversionMinSwapFinalityConfirmations uint32 `json:"fiat_conversion_min_swap_finality_confirmations"`
	FiatConversionObservationMaxPastSeconds    uint64 `json:"fiat_conversion_observation_max_past_seconds"`
	FiatConversionObservationMaxFutureSeconds  uint64 `json:"fiat_conversion_observation_max_future_seconds"`
	FiatConversionMaxObservations              uint32 `json:"fiat_conversion_max_observations"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                       DefaultParams(),
		EscrowAccounts:               []EscrowAccount{},
		SettlementRecords:            []SettlementRecord{},
		RewardDistributions:          []RewardDistribution{},
		UsageRecords:                 []UsageRecord{},
		ClaimableRewards:             []ClaimableRewards{},
		PayoutRecords:                []PayoutRecord{},
		FiatConversionRecords:        []FiatConversionRecord{},
		FiatPayoutPreferences:        []FiatPayoutPreference{},
		EscrowSequence:               1,
		SettlementSequence:           1,
		DistributionSequence:         1,
		UsageSequence:                1,
		PayoutSequence:               1,
		FiatConversionSequence:       1,
		UsageAuthenticationActive:    true,
		FinancialCases:               []FinancialCase{},
		FinancialCasesActive:         true,
		FiatConversionCustodyBalance: sdk.NewCoins(),
	}
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		PlatformFeeRate:                        "0.05",   // 5%
		ValidatorFeeRate:                       "0.01",   // 1%
		MinEscrowDuration:                      3600,     // 1 hour
		MaxEscrowDuration:                      31536000, // 1 year
		SettlementPeriod:                       86400,    // 1 day
		RewardClaimExpiry:                      2592000,  // 30 days
		MinSettlementAmount:                    "1000",   // Minimum tokens for settlement
		UsageGracePeriod:                       86400,    // 1 day grace period
		StakingRewardEpochLength:               100,      // 100 blocks per epoch
		VerificationRewardAmount:               "100",    // Base reward for verification
		PayoutHoldbackRate:                     "0.0",    // No holdback by default
		MaxPayoutRetries:                       3,        // 3 retry attempts
		DisputeWindowDuration:                  604800,   // 7 days
		UsageRewardRateBps:                     1000,     // 10% base reward on usage value
		UsageRewardCPUMultiplierBps:            10000,    // 1.0x
		UsageRewardMemoryMultiplierBps:         10000,    // 1.0x
		UsageRewardStorageMultiplierBps:        10000,    // 1.0x
		UsageRewardGPUMultiplierBps:            12000,    // 1.2x
		UsageRewardNetworkMultiplierBps:        9000,     // 0.9x
		UsageRewardSLAOnTimeMultiplierBps:      10000,    // 1.0x
		UsageRewardSLALateMultiplierBps:        8000,     // 0.8x
		UsageRewardAcknowledgedMultiplierBps:   10000,    // 1.0x
		UsageRewardUnacknowledgedMultiplierBps: 9000,     // 0.9x
		FiatConversionEnabled:                  false,
		FiatConversionMinAmount:                "1000",
		FiatConversionMaxAmount:                "100000000",
		FiatConversionDailyLimit:               "1000000000",
		FiatConversionStableDenom:              "uusdc",
		FiatConversionStableSymbol:             "USDC",
		FiatConversionStableDecimals:           6,
		FiatConversionDefaultFiat:              "USD",
		FiatConversionDefaultMethod:            "bank_transfer",
		FiatConversionMaxSlippage:              "0.02",
		FiatConversionRiskScoreThreshold:       75,
		FiatConversionMinComplianceStatus:      "CLEARED",
		FiatConversionSpreadBps:                50,
		OracleSources: []OracleSourceConfig{
			{ID: "cosmos-oracle", Type: OracleSourceTypeCosmosOracle, Enabled: true, Priority: 1},
			{ID: "band-ibc", Type: OracleSourceTypeBandIBC, Enabled: true, Priority: 2},
			{ID: "chainlink-ibc", Type: OracleSourceTypeChainlinkIBC, Enabled: true, Priority: 3},
			{ID: "manual", Type: OracleSourceTypeManual, Enabled: true, Priority: 100},
		},
		OracleStalenessThresholdSeconds:        300,
		OracleMinSources:                       3,
		OracleManualPrices:                     []ManualPriceOverride{},
		OracleDeviationThresholdBps:            500,
		OracleDeviationWindowSeconds:           60,
		FinancialCaseFilingWindowSeconds:       604800,
		FinancialCaseEvidenceWindowSeconds:     604800,
		FinancialCaseReviewWindowSeconds:       604800,
		FinancialCaseAppealWindowSeconds:       86400,
		FinancialCaseEscalationWindowSeconds:   2592000,
		FinancialCaseFilingWindowBlocks:        120960,
		FinancialCaseEvidenceWindowBlocks:      120960,
		FinancialCaseReviewWindowBlocks:        120960,
		FinancialCaseAppealWindowBlocks:        17280,
		FinancialCaseEscalationWindowBlocks:    518400,
		FinancialCaseMaxClaims:                 32,
		FinancialCaseMaxAppeals:                1,
		FinancialCaseMaxEvidenceReferenceBytes: 512,
		FinancialCaseTimeoutBatchLimit:         100,
		// No external route/corridor is certified by local engineering work.
		FiatConversionDEXProfileState:              settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_ENGINEERING_COMPLETE_EXTERNAL_BLOCKED,
		FiatConversionPayoutProfileState:           settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_ENGINEERING_COMPLETE_EXTERNAL_BLOCKED,
		FiatConversionMinSwapFinalityConfirmations: 2,
		FiatConversionObservationMaxPastSeconds:    3600,
		FiatConversionObservationMaxFutureSeconds:  30,
		FiatConversionMaxObservations:              16,
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate escrow accounts
	seenEscrows := make(map[string]bool)
	for _, escrow := range gs.EscrowAccounts {
		if err := escrow.Validate(); err != nil {
			return err
		}
		if seenEscrows[escrow.EscrowID] {
			return ErrEscrowExists.Wrapf("duplicate escrow_id: %s", escrow.EscrowID)
		}
		seenEscrows[escrow.EscrowID] = true
	}

	// Validate settlement records
	seenSettlements := make(map[string]bool)
	for _, settlement := range gs.SettlementRecords {
		if err := settlement.Validate(); err != nil {
			return err
		}
		if seenSettlements[settlement.SettlementID] {
			return ErrSettlementExists.Wrapf("duplicate settlement_id: %s", settlement.SettlementID)
		}
		seenSettlements[settlement.SettlementID] = true
	}

	// Validate reward distributions
	seenDistributions := make(map[string]bool)
	for _, dist := range gs.RewardDistributions {
		if err := dist.Validate(); err != nil {
			return err
		}
		if seenDistributions[dist.DistributionID] {
			return ErrInvalidReward.Wrapf("duplicate distribution_id: %s", dist.DistributionID)
		}
		seenDistributions[dist.DistributionID] = true
	}

	// Validate usage records
	seenUsage := make(map[string]bool)
	for _, usage := range gs.UsageRecords {
		if err := usage.Validate(); err != nil {
			return err
		}
		if seenUsage[usage.UsageID] {
			return ErrUsageRecordExists.Wrapf("duplicate usage_id: %s", usage.UsageID)
		}
		seenUsage[usage.UsageID] = true
		if gs.UsageAuthenticationActive && !usage.IsAuthenticated() {
			return ErrUsageAuthenticationRequired.Wrapf("genesis usage %s is not authenticated", usage.UsageID)
		}
	}

	// Validate payout records
	seenPayouts := make(map[string]bool)
	for _, payout := range gs.PayoutRecords {
		if err := payout.Validate(); err != nil {
			return err
		}
		if seenPayouts[payout.PayoutID] {
			return ErrPayoutExists.Wrapf("duplicate payout_id: %s", payout.PayoutID)
		}
		seenPayouts[payout.PayoutID] = true
	}

	seenCases := make(map[string]bool)
	seenActiveAliases := make(map[string]string)
	for i, financialCase := range gs.FinancialCases {
		if financialCase.CaseId == "" || seenCases[financialCase.CaseId] {
			return ErrInvalidFinancialCase.Wrapf("duplicate or empty financial case at %d", i)
		}
		seenCases[financialCase.CaseId] = true
		if !financialCase.Exposure.OriginalHeld.IsValid() || financialCase.Exposure.OriginalHeld.IsZero() {
			return ErrInvalidFinancialCase.Wrapf("financial case %s has invalid exposure", financialCase.CaseId)
		}
		if _, err := sdk.AccAddressFromBech32(financialCase.Provider); err != nil {
			return ErrInvalidFinancialCase.Wrapf("financial case %s has invalid provider", financialCase.CaseId)
		}
		if _, err := sdk.AccAddressFromBech32(financialCase.Customer); err != nil || financialCase.Provider == financialCase.Customer {
			return ErrInvalidFinancialCase.Wrapf("financial case %s has invalid customer", financialCase.CaseId)
		}
		if IsActiveFinancialCaseStatus(financialCase.Status) && financialCase.ActiveHoldCount == 0 {
			return ErrFinancialCaseHold.Wrapf("financial case %s has no hold", financialCase.CaseId)
		}
		if IsTerminalFinancialCaseStatus(financialCase.Status) && financialCase.ActiveHoldCount != 0 {
			return ErrFinancialCaseHold.Wrapf("terminal financial case %s retains hold", financialCase.CaseId)
		}
		if financialCase.Exposure.PayoutId != "" {
			found := false
			for _, payout := range gs.PayoutRecords {
				if payout.PayoutID == financialCase.Exposure.PayoutId {
					found = true
					if IsActiveFinancialCaseStatus(financialCase.Status) && (payout.State != PayoutStateHeld || payout.DisputeID != financialCase.CaseId) {
						return ErrFinancialCaseHold.Wrapf("financial case %s payout hold is inconsistent", financialCase.CaseId)
					}
					break
				}
			}
			if !found {
				return ErrFinancialCaseHold.Wrapf("financial case %s payout is missing", financialCase.CaseId)
			}
		}
		if financialCase.Exposure.EscrowId != "" {
			found := false
			for _, escrow := range gs.EscrowAccounts {
				if escrow.EscrowID == financialCase.Exposure.EscrowId {
					found = true
					if IsActiveFinancialCaseStatus(financialCase.Status) && escrow.State != EscrowStateDisputed {
						return ErrFinancialCaseHold.Wrapf("financial case %s escrow hold is inconsistent", financialCase.CaseId)
					}
					break
				}
			}
			if !found {
				return ErrFinancialCaseHold.Wrapf("financial case %s escrow is missing", financialCase.CaseId)
			}
		}
		if IsActiveFinancialCaseStatus(financialCase.Status) {
			aliases := []string{fmt.Sprintf("subject:%d/%s", financialCase.Subject.Type, financialCase.Subject.PrimaryId)}
			for _, alias := range []struct{ kind, value string }{
				{"order", financialCase.Subject.OrderId}, {"invoice", financialCase.Subject.InvoiceId},
				{"usage", financialCase.Subject.UsageId}, {"job", financialCase.Subject.HpcJobId},
				{"settlement", financialCase.Subject.SettlementId}, {"escrow", financialCase.Exposure.EscrowId},
				{"reservation", financialCase.Subject.ReservationId}, {"lease", financialCase.Subject.LeaseId},
			} {
				if alias.value != "" {
					aliases = append(aliases, alias.kind+":"+alias.value)
				}
			}
			for _, alias := range aliases {
				if owner, exists := seenActiveAliases[alias]; exists && owner != financialCase.CaseId {
					return ErrInvalidFinancialCase.Wrapf("active alias %s has cases %s and %s", alias, owner, financialCase.CaseId)
				}
				seenActiveAliases[alias] = financialCase.CaseId
			}
		}
	}

	// Validate fiat conversion records
	seenConversions := make(map[string]bool)
	expectedCustody := sdk.NewCoins()
	for _, conversion := range gs.FiatConversionRecords {
		if err := conversion.Validate(); err != nil {
			return err
		}
		if conversion.ProtocolVersion == 0 || (conversion.ProtocolVersion > 0 && len(conversion.RequestDigest) != 32) {
			return ErrFiatConversionQuarantined.Wrapf("conversion %s is not protocol-migrated", conversion.ConversionID)
		}
		if !conversion.LegacyQuarantined && conversion.PayoutID == "" {
			return ErrInvalidPayout.Wrapf("conversion %s has no payout hold", conversion.ConversionID)
		}
		if seenConversions[conversion.ConversionID] {
			return ErrInvalidSettlement.Wrapf("duplicate conversion_id: %s", conversion.ConversionID)
		}
		seenConversions[conversion.ConversionID] = true
		if conversion.ValueMovementApplied && !conversion.LegacyQuarantined {
			expectedCustody = expectedCustody.Add(conversion.CustodySinkAmount)
		}
		if conversion.PayoutID != "" {
			linked := false
			for _, payout := range gs.PayoutRecords {
				if payout.PayoutID == conversion.PayoutID {
					linked = payout.FiatConversionID == conversion.ConversionID && payout.Provider == conversion.Provider && payout.Customer == conversion.Customer &&
						payout.InvoiceID == conversion.InvoiceID && payout.SettlementID == conversion.SettlementID && len(payout.NetAmount) == 1 && payout.NetAmount[0].IsEqual(conversion.CryptoAmount)
					if linked && conversion.State == FiatConversionStatePayoutCompleted && !conversion.LegacyQuarantined {
						linked = conversion.ValueMovementApplied && payout.ValueMovementApplied && conversion.CustodySinkAmount.IsEqual(conversion.CryptoAmount) &&
							len(conversion.CustodySinkEffectHash) == 32 && bytes.Equal(conversion.CustodySinkEffectHash, payout.ValueMovementEffectHash)
					}
					break
				}
			}
			if !linked {
				return ErrInvalidPayout.Wrapf("conversion %s payout linkage mismatch", conversion.ConversionID)
			}
		}
	}
	if !gs.FiatConversionCustodyBalance.Equal(expectedCustody) {
		return ErrInvalidSettlement.Wrapf("fiat custody balance mismatch: expected %s got %s", expectedCustody, gs.FiatConversionCustodyBalance)
	}

	// Validate fiat payout preferences
	seenPrefs := make(map[string]bool)
	for _, pref := range gs.FiatPayoutPreferences {
		if err := pref.Validate(); err != nil {
			return err
		}
		if seenPrefs[pref.Provider] {
			return ErrInvalidParams.Wrapf("duplicate fiat payout preference: %s", pref.Provider)
		}
		seenPrefs[pref.Provider] = true
	}

	return nil
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.FinancialCaseMaxClaims > 256 || p.FinancialCaseMaxAppeals > 8 || p.FinancialCaseMaxEvidenceReferenceBytes > 4096 || p.FinancialCaseTimeoutBatchLimit > 1000 {
		return ErrInvalidParams.Wrap("financial case limits exceed protocol maximum")
	}
	for _, blocks := range []int64{p.FinancialCaseFilingWindowBlocks, p.FinancialCaseEvidenceWindowBlocks, p.FinancialCaseReviewWindowBlocks, p.FinancialCaseAppealWindowBlocks, p.FinancialCaseEscalationWindowBlocks} {
		if blocks < 0 {
			return ErrInvalidParams.Wrap("financial case block windows cannot be negative")
		}
	}
	// Validate fee rates are between 0 and 1
	// We'll do basic validation here; more sophisticated parsing would be needed in production
	if p.PlatformFeeRate != "" {
		fee, err := sdkmath.LegacyNewDecFromStr(p.PlatformFeeRate)
		if err != nil || fee.IsNegative() || fee.GT(sdkmath.LegacyOneDec()) {
			return ErrInvalidParams.Wrap("platform_fee_rate must be between 0 and 1")
		}
	}

	if p.ValidatorFeeRate != "" {
		fee, err := sdkmath.LegacyNewDecFromStr(p.ValidatorFeeRate)
		if err != nil || fee.IsNegative() || fee.GT(sdkmath.LegacyOneDec()) {
			return ErrInvalidParams.Wrap("validator_fee_rate must be between 0 and 1")
		}
	}

	if p.MinEscrowDuration == 0 {
		return ErrInvalidParams.Wrap("min_escrow_duration must be greater than zero")
	}

	if p.MaxEscrowDuration <= p.MinEscrowDuration {
		return ErrInvalidParams.Wrap("max_escrow_duration must be greater than min_escrow_duration")
	}

	if p.SettlementPeriod == 0 {
		return ErrInvalidParams.Wrap("settlement_period must be greater than zero")
	}

	if p.StakingRewardEpochLength == 0 {
		return ErrInvalidParams.Wrap("staking_reward_epoch_length must be greater than zero")
	}

	if p.RewardPoolAddress != "" {
		if _, err := sdk.AccAddressFromBech32(p.RewardPoolAddress); err != nil {
			return ErrInvalidParams.Wrap("reward_pool_address must be a valid bech32 address")
		}
	}

	if p.UsageRewardRateBps > 10000 {
		return ErrInvalidParams.Wrap("usage_reward_rate_bps cannot exceed 10000")
	}

	if err := validateRewardMultiplierBps(p.UsageRewardCPUMultiplierBps, "usage_reward_cpu_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardMemoryMultiplierBps, "usage_reward_memory_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardStorageMultiplierBps, "usage_reward_storage_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardGPUMultiplierBps, "usage_reward_gpu_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardNetworkMultiplierBps, "usage_reward_network_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardSLAOnTimeMultiplierBps, "usage_reward_sla_ontime_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardSLALateMultiplierBps, "usage_reward_sla_late_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardAcknowledgedMultiplierBps, "usage_reward_ack_multiplier_bps"); err != nil {
		return err
	}
	if err := validateRewardMultiplierBps(p.UsageRewardUnacknowledgedMultiplierBps, "usage_reward_unack_multiplier_bps"); err != nil {
		return err
	}

	if p.FiatConversionEnabled {
		if err := validateCertifiedFiatProfile(p.FiatConversionDEXProfileID, p.FiatConversionDEXProfileDigest, p.FiatConversionDEXProfileState, "dex"); err != nil {
			return err
		}
		if err := validateCertifiedFiatProfile(p.FiatConversionPayoutProfileID, p.FiatConversionPayoutProfileDigest, p.FiatConversionPayoutProfileState, "payout"); err != nil {
			return err
		}
		if p.FiatConversionMinSwapFinalityConfirmations == 0 {
			return ErrInvalidParams.Wrap("fiat conversion swap finality confirmations must be positive")
		}
		if p.FiatConversionStableDenom == "" || p.FiatConversionStableSymbol == "" {
			return ErrInvalidParams.Wrap("fiat conversion stable token must be configured")
		}

		if p.FiatConversionStableDecimals > 18 {
			return ErrInvalidParams.Wrap("fiat conversion stable decimals must be <= 18")
		}

		minAmount, ok := sdkmath.NewIntFromString(p.FiatConversionMinAmount)
		if !ok || minAmount.IsNegative() {
			return ErrInvalidParams.Wrap("fiat_conversion_min_amount must be a valid non-negative integer")
		}

		maxAmount, ok := sdkmath.NewIntFromString(p.FiatConversionMaxAmount)
		if !ok || maxAmount.IsNegative() {
			return ErrInvalidParams.Wrap("fiat_conversion_max_amount must be a valid non-negative integer")
		}

		if maxAmount.IsPositive() && minAmount.GT(maxAmount) {
			return ErrInvalidParams.Wrap("fiat_conversion_min_amount cannot exceed max amount")
		}

		dailyLimit, ok := sdkmath.NewIntFromString(p.FiatConversionDailyLimit)
		if !ok || dailyLimit.IsNegative() {
			return ErrInvalidParams.Wrap("fiat_conversion_daily_limit must be a valid non-negative integer")
		}

		if p.FiatConversionDefaultFiat == "" {
			return ErrInvalidParams.Wrap("fiat_conversion_default_fiat required")
		}

		if p.FiatConversionDefaultMethod == "" {
			return ErrInvalidParams.Wrap("fiat_conversion_default_method required")
		}

		if p.FiatConversionMaxSlippage == "" {
			return ErrInvalidParams.Wrap("fiat_conversion_max_slippage required")
		}

		if _, err := sdkmath.LegacyNewDecFromStr(p.FiatConversionMaxSlippage); err != nil {
			return ErrInvalidParams.Wrapf("invalid fiat_conversion_max_slippage: %s", err)
		}

		if p.FiatConversionRiskScoreThreshold < 0 || p.FiatConversionRiskScoreThreshold > 100 {
			return ErrInvalidParams.Wrap("fiat_conversion_risk_score_threshold must be between 0 and 100")
		}

		if p.FiatConversionMinComplianceStatus == "" {
			return ErrInvalidParams.Wrap("fiat_conversion_min_compliance_status required")
		}
	}
	if !validFiatProfileState(p.FiatConversionDEXProfileState) || !validFiatProfileState(p.FiatConversionPayoutProfileState) {
		return ErrInvalidParams.Wrap("invalid fiat conversion profile state")
	}
	if p.FiatConversionObservationMaxPastSeconds == 0 || p.FiatConversionObservationMaxFutureSeconds > 300 {
		return ErrInvalidParams.Wrap("invalid fiat conversion observation time bounds")
	}
	if p.FiatConversionMaxObservations == 0 || p.FiatConversionMaxObservations > 64 {
		return ErrInvalidParams.Wrap("fiat conversion max observations must be between 1 and 64")
	}

	if p.FiatConversionSpreadBps > 10000 {
		return ErrInvalidParams.Wrap("fiat_conversion_spread_bps cannot exceed 10000")
	}

	if p.OracleStalenessThresholdSeconds == 0 {
		return ErrInvalidParams.Wrap("oracle_staleness_threshold_seconds must be greater than zero")
	}

	if p.OracleMinSources == 0 {
		return ErrInvalidParams.Wrap("oracle_min_sources must be greater than zero")
	}

	if p.OracleDeviationThresholdBps == 0 || p.OracleDeviationThresholdBps > 10000 {
		return ErrInvalidParams.Wrap("oracle_deviation_threshold_bps must be between 1 and 10000")
	}

	if p.OracleDeviationWindowSeconds == 0 {
		return ErrInvalidParams.Wrap("oracle_deviation_window_seconds must be greater than zero")
	}

	seenOracle := make(map[string]bool)
	enabledSources := 0
	for _, source := range p.OracleSources {
		if err := source.Validate(); err != nil {
			return err
		}
		if seenOracle[source.ID] {
			return ErrInvalidParams.Wrapf("duplicate oracle source: %s", source.ID)
		}
		seenOracle[source.ID] = true
		if source.Enabled {
			enabledSources++
		}
	}
	if enabledSources == 0 {
		return ErrInvalidParams.Wrap("at least one oracle source must be enabled")
	}

	for _, override := range p.OracleManualPrices {
		if err := override.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func validFiatProfileState(state settlementv1.FiatConversionProfileState) bool {
	switch state {
	case settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_UNSUPPORTED,
		settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_ENGINEERING_INCOMPLETE,
		settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_ENGINEERING_COMPLETE_EXTERNAL_BLOCKED,
		settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED,
		settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_PAUSED:
		return true
	default:
		return false
	}
}

func validateCertifiedFiatProfile(id string, digest []byte, state settlementv1.FiatConversionProfileState, kind string) error {
	if state != settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED {
		return ErrInvalidParams.Wrapf("fiat conversion %s profile is not certified_enabled", kind)
	}
	if id == "" || len(id) > 128 {
		return ErrInvalidParams.Wrapf("fiat conversion %s profile id is invalid", kind)
	}
	if len(digest) != 32 {
		return ErrInvalidParams.Wrapf("fiat conversion %s profile digest must be SHA-256", kind)
	}
	return nil
}

func validateRewardMultiplierBps(value uint32, name string) error {
	if value == 0 {
		return ErrInvalidParams.Wrapf("%s must be greater than zero", name)
	}
	if value > 20000 {
		return ErrInvalidParams.Wrapf("%s cannot exceed 20000", name)
	}
	return nil
}

// ProtoMessage implements proto.Message
func (*GenesisState) ProtoMessage() {}

// Reset implements proto.Message
func (gs *GenesisState) Reset() { *gs = GenesisState{} }

// String implements proto.Message
func (gs *GenesisState) String() string {
	return fmt.Sprintf("%+v", *gs)
}
