package keeper

import (
	"time"

	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	"github.com/virtengine/virtengine/x/settlement/types"
)

func toProtoEscrowAccount(escrow types.EscrowAccount) settlementv1.EscrowAccount {
	return settlementv1.EscrowAccount{
		EscrowId:        escrow.EscrowID,
		OrderId:         escrow.OrderID,
		LeaseId:         escrow.LeaseID,
		Depositor:       escrow.Depositor,
		Recipient:       escrow.Recipient,
		Amount:          escrow.Amount,
		Balance:         escrow.Balance,
		State:           string(escrow.State),
		Conditions:      toProtoReleaseConditions(escrow.Conditions),
		CreatedAt:       unixTimestamp(escrow.CreatedAt),
		ExpiresAt:       unixTimestamp(escrow.ExpiresAt),
		ActivatedAt:     unixTimestampPtr(escrow.ActivatedAt),
		ClosedAt:        unixTimestampPtr(escrow.ClosedAt),
		ClosureReason:   escrow.ClosureReason,
		TotalSettled:    escrow.TotalSettled,
		SettlementCount: escrow.SettlementCount,
		BlockHeight:     escrow.BlockHeight,
	}
}

func toProtoReleaseConditions(conditions []types.ReleaseCondition) []settlementv1.ReleaseCondition {
	if len(conditions) == 0 {
		return nil
	}

	resp := make([]settlementv1.ReleaseCondition, 0, len(conditions))
	for _, condition := range conditions {
		resp = append(resp, settlementv1.ReleaseCondition{
			Type:                      string(condition.Type),
			UnlockAfter:               unixTimestampPtr(condition.UnlockAfter),
			RequiredSigners:           condition.RequiredSigners,
			SignatureThreshold:        condition.SignatureThreshold,
			MinUsageUnits:             condition.MinUsageUnits,
			RequiredVerificationScore: condition.RequiredVerificationScore,
			Satisfied:                 condition.Satisfied,
			SatisfiedAt:               unixTimestampPtr(condition.SatisfiedAt),
		})
	}
	return resp
}

func toProtoSettlementRecord(settlement types.SettlementRecord) settlementv1.SettlementRecord {
	return settlementv1.SettlementRecord{
		SettlementId:    settlement.SettlementID,
		EscrowId:        settlement.EscrowID,
		OrderId:         settlement.OrderID,
		LeaseId:         settlement.LeaseID,
		Provider:        settlement.Provider,
		Customer:        settlement.Customer,
		TotalAmount:     settlement.TotalAmount,
		ProviderShare:   settlement.ProviderShare,
		PlatformFee:     settlement.PlatformFee,
		ValidatorFee:    settlement.ValidatorFee,
		SettledAt:       unixTimestamp(settlement.SettledAt),
		UsageRecordIds:  settlement.UsageRecordIDs,
		TotalUsageUnits: settlement.TotalUsageUnits,
		PeriodStart:     unixTimestamp(settlement.PeriodStart),
		PeriodEnd:       unixTimestamp(settlement.PeriodEnd),
		BlockHeight:     settlement.BlockHeight,
		SettlementType:  string(settlement.SettlementType),
		IsFinal:         settlement.IsFinal,
	}
}

func toProtoUsageRecord(usage types.UsageRecord) settlementv1.UsageRecord {
	return settlementv1.UsageRecord{
		UsageId:              usage.UsageID,
		OrderId:              usage.OrderID,
		LeaseId:              usage.LeaseID,
		Provider:             usage.Provider,
		Customer:             usage.Customer,
		UsageUnits:           usage.UsageUnits,
		UsageType:            usage.UsageType,
		PeriodStart:          unixTimestamp(usage.PeriodStart),
		PeriodEnd:            unixTimestamp(usage.PeriodEnd),
		UnitPrice:            usage.UnitPrice,
		TotalCost:            usage.TotalCost,
		ProviderSignature:    usage.ProviderSignature,
		CustomerAcknowledged: usage.CustomerAcknowledged,
		CustomerSignature:    usage.CustomerSignature,
		Settled:              usage.Settled,
		SettlementId:         usage.SettlementID,
		SubmittedAt:          unixTimestamp(usage.SubmittedAt),
		BlockHeight:          usage.BlockHeight,
		Metadata:             usage.Metadata,
	}
}

func toProtoUsageSummary(summary types.UsageSummary) settlementv1.UsageSummary {
	return settlementv1.UsageSummary{
		Provider:        summary.Provider,
		OrderId:         summary.OrderID,
		PeriodStart:     unixTimestamp(summary.PeriodStart),
		PeriodEnd:       unixTimestamp(summary.PeriodEnd),
		TotalUsageUnits: summary.TotalUsage,
		TotalCost:       summary.TotalCost,
		ByUsageType:     toProtoUsageTypeSummaries(summary.ByUsageType),
		GeneratedAt:     unixTimestamp(summary.GeneratedAt),
		BlockHeight:     summary.BlockHeight,
		UsageRecordIds:  summary.UsageRecordIDs,
	}
}

func toProtoUsageTypeSummaries(entries []types.UsageTypeSummary) []settlementv1.UsageTypeSummary {
	if len(entries) == 0 {
		return nil
	}

	resp := make([]settlementv1.UsageTypeSummary, 0, len(entries))
	for _, entry := range entries {
		resp = append(resp, settlementv1.UsageTypeSummary{
			UsageType:  entry.UsageType,
			UsageUnits: entry.UsageUnits,
			TotalCost:  entry.TotalCost,
		})
	}
	return resp
}

func toProtoRewardDistribution(dist types.RewardDistribution) settlementv1.RewardDistribution {
	return settlementv1.RewardDistribution{
		DistributionId:    dist.DistributionID,
		EpochNumber:       dist.EpochNumber,
		TotalRewards:      dist.TotalRewards,
		Recipients:        toProtoRewardRecipients(dist.Recipients),
		Source:            string(dist.Source),
		DistributedAt:     unixTimestamp(dist.DistributedAt),
		BlockHeight:       dist.BlockHeight,
		ReferenceTxHashes: dist.ReferenceTxHashes,
		Metadata:          dist.Metadata,
	}
}

func toProtoRewardRecipients(recipients []types.RewardRecipient) []settlementv1.RewardRecipient {
	if len(recipients) == 0 {
		return nil
	}

	resp := make([]settlementv1.RewardRecipient, 0, len(recipients))
	for _, recipient := range recipients {
		resp = append(resp, settlementv1.RewardRecipient{
			Address:           recipient.Address,
			Amount:            recipient.Amount,
			Reason:            recipient.Reason,
			UsageUnits:        recipient.UsageUnits,
			VerificationScore: recipient.VerificationScore,
			StakingWeight:     recipient.StakingWeight,
			ReferenceId:       recipient.ReferenceID,
		})
	}
	return resp
}

func toProtoRewardHistoryEntry(entry types.RewardHistoryEntry) settlementv1.RewardHistoryEntry {
	return settlementv1.RewardHistoryEntry{
		DistributionId: entry.DistributionID,
		EpochNumber:    entry.EpochNumber,
		Source:         string(entry.Source),
		Amount:         entry.Amount,
		Reason:         entry.Reason,
		UsageUnits:     entry.UsageUnits,
		ReferenceId:    entry.ReferenceID,
		DistributedAt:  unixTimestamp(entry.DistributedAt),
	}
}

func toProtoClaimableRewards(rewards types.ClaimableRewards) settlementv1.ClaimableRewards {
	return settlementv1.ClaimableRewards{
		Address:        rewards.Address,
		TotalClaimable: rewards.TotalClaimable,
		RewardEntries:  toProtoRewardEntries(rewards.RewardEntries),
		LastUpdated:    unixTimestamp(rewards.LastUpdated),
		TotalClaimed:   rewards.TotalClaimed,
	}
}

func toProtoRewardEntries(entries []types.RewardEntry) []settlementv1.RewardEntry {
	if len(entries) == 0 {
		return nil
	}

	resp := make([]settlementv1.RewardEntry, 0, len(entries))
	for _, entry := range entries {
		resp = append(resp, settlementv1.RewardEntry{
			DistributionId: entry.DistributionID,
			Source:         string(entry.Source),
			Amount:         entry.Amount,
			CreatedAt:      unixTimestamp(entry.CreatedAt),
			ExpiresAt:      unixTimestampPtr(entry.ExpiresAt),
			Reason:         entry.Reason,
		})
	}
	return resp
}

func toProtoPayoutRecord(payout types.PayoutRecord) settlementv1.PayoutRecord {
	return settlementv1.PayoutRecord{
		PayoutId:          payout.PayoutID,
		FiatConversionId:  payout.FiatConversionID,
		InvoiceId:         payout.InvoiceID,
		SettlementId:      payout.SettlementID,
		EscrowId:          payout.EscrowID,
		OrderId:           payout.OrderID,
		LeaseId:           payout.LeaseID,
		Provider:          payout.Provider,
		Customer:          payout.Customer,
		GrossAmount:       payout.GrossAmount,
		PlatformFee:       payout.PlatformFee,
		ValidatorFee:      payout.ValidatorFee,
		HoldbackAmount:    payout.HoldbackAmount,
		NetAmount:         payout.NetAmount,
		State:             string(payout.State),
		DisputeId:         payout.DisputeID,
		HoldReason:        payout.HoldReason,
		IdempotencyKey:    payout.IdempotencyKey,
		ExecutionAttempts: payout.ExecutionAttempts,
		LastAttemptAt:     unixTimestampPtr(payout.LastAttemptAt),
		LastError:         payout.LastError,
		TxHash:            payout.TxHash,
		CreatedAt:         unixTimestamp(payout.CreatedAt),
		ProcessedAt:       unixTimestampPtr(payout.ProcessedAt),
		CompletedAt:       unixTimestampPtr(payout.CompletedAt),
		BlockHeight:       payout.BlockHeight,
	}
}

func toProtoParams(params types.Params) settlementv1.Params {
	return settlementv1.Params{
		PlatformFeeRate:                   params.PlatformFeeRate,
		ValidatorFeeRate:                  params.ValidatorFeeRate,
		MinEscrowDuration:                 params.MinEscrowDuration,
		MaxEscrowDuration:                 params.MaxEscrowDuration,
		SettlementPeriod:                  params.SettlementPeriod,
		RewardClaimExpiry:                 params.RewardClaimExpiry,
		MinSettlementAmount:               params.MinSettlementAmount,
		UsageGracePeriod:                  params.UsageGracePeriod,
		StakingRewardEpochLength:          params.StakingRewardEpochLength,
		VerificationRewardAmount:          params.VerificationRewardAmount,
		PayoutHoldbackRate:                params.PayoutHoldbackRate,
		MaxPayoutRetries:                  params.MaxPayoutRetries,
		DisputeWindowDuration:             params.DisputeWindowDuration,
		UsageRewardRateBps:                params.UsageRewardRateBps,
		UsageRewardCpuMultiplierBps:       params.UsageRewardCPUMultiplierBps,
		UsageRewardMemoryMultiplierBps:    params.UsageRewardMemoryMultiplierBps,
		UsageRewardStorageMultiplierBps:   params.UsageRewardStorageMultiplierBps,
		UsageRewardGpuMultiplierBps:       params.UsageRewardGPUMultiplierBps,
		UsageRewardNetworkMultiplierBps:   params.UsageRewardNetworkMultiplierBps,
		UsageRewardSlaOntimeMultiplierBps: params.UsageRewardSLAOnTimeMultiplierBps,
		UsageRewardSlaLateMultiplierBps:   params.UsageRewardSLALateMultiplierBps,
		UsageRewardAckMultiplierBps:       params.UsageRewardAcknowledgedMultiplierBps,
		UsageRewardUnackMultiplierBps:     params.UsageRewardUnacknowledgedMultiplierBps,
		FiatConversionEnabled:             params.FiatConversionEnabled,
		FiatConversionMinAmount:           params.FiatConversionMinAmount,
		FiatConversionMaxAmount:           params.FiatConversionMaxAmount,
		FiatConversionDailyLimit:          params.FiatConversionDailyLimit,
		FiatConversionStableDenom:         params.FiatConversionStableDenom,
		FiatConversionStableSymbol:        params.FiatConversionStableSymbol,
		FiatConversionStableDecimals:      params.FiatConversionStableDecimals,
		FiatConversionDefaultFiat:         params.FiatConversionDefaultFiat,
		FiatConversionDefaultMethod:       params.FiatConversionDefaultMethod,
		FiatConversionMaxSlippage:         params.FiatConversionMaxSlippage,
		FiatConversionRiskScoreThreshold:  params.FiatConversionRiskScoreThreshold,
		FiatConversionMinComplianceStatus: params.FiatConversionMinComplianceStatus,
	}
}

func toProtoFiatConversionRecord(record types.FiatConversionRecord) settlementv1.FiatConversionRecord {
	return settlementv1.FiatConversionRecord{
		ConversionId:        record.ConversionID,
		InvoiceId:           record.InvoiceID,
		SettlementId:        record.SettlementID,
		PayoutId:            record.PayoutID,
		EscrowId:            record.EscrowID,
		OrderId:             record.OrderID,
		LeaseId:             record.LeaseID,
		Provider:            record.Provider,
		Customer:            record.Customer,
		RequestedBy:         record.RequestedBy,
		RequestedAt:         unixTimestamp(record.RequestedAt),
		UpdatedAt:           unixTimestamp(record.UpdatedAt),
		State:               string(record.State),
		CryptoToken:         toProtoTokenSpec(record.CryptoToken),
		StableToken:         toProtoTokenSpec(record.StableToken),
		CryptoAmount:        record.CryptoAmount,
		StableAmount:        record.StableAmount,
		FiatCurrency:        record.FiatCurrency,
		FiatAmount:          record.FiatAmount,
		PaymentMethod:       record.PaymentMethod,
		DestinationRef:      record.DestinationRef,
		DestinationHash:     record.DestinationHash,
		DestinationRegion:   record.DestinationRegion,
		SlippageTolerance:   record.SlippageTolerance,
		DexAdapter:          record.DexAdapter,
		SwapQuoteId:         record.SwapQuoteID,
		SwapTxHash:          record.SwapTxHash,
		SwapStatus:          record.SwapStatus,
		OffRampProvider:     record.OffRampProvider,
		OffRampQuoteId:      record.OffRampQuoteID,
		OffRampId:           record.OffRampID,
		OffRampStatus:       record.OffRampStatus,
		OffRampReference:    record.OffRampReference,
		ComplianceStatus:    record.ComplianceStatus,
		ComplianceRiskScore: record.ComplianceRiskScore,
		ComplianceCheckedAt: record.ComplianceCheckedAt,
		FailureReason:       record.FailureReason,
		AuditTrail:          toProtoFiatConversionAuditEntries(record.AuditTrail),
	}
}

func toProtoFiatConversionAuditEntries(entries []types.FiatConversionAuditEntry) []settlementv1.FiatConversionAuditEntry {
	if len(entries) == 0 {
		return nil
	}

	resp := make([]settlementv1.FiatConversionAuditEntry, 0, len(entries))
	for _, entry := range entries {
		resp = append(resp, settlementv1.FiatConversionAuditEntry{
			Action:    entry.Action,
			Actor:     entry.Actor,
			Reason:    entry.Reason,
			Timestamp: entry.Timestamp,
			Metadata:  entry.Metadata,
		})
	}
	return resp
}

func toProtoFiatPayoutPreference(pref types.FiatPayoutPreference) settlementv1.FiatPayoutPreference {
	return settlementv1.FiatPayoutPreference{
		Provider:          pref.Provider,
		Enabled:           pref.Enabled,
		FiatCurrency:      pref.FiatCurrency,
		PaymentMethod:     pref.PaymentMethod,
		DestinationRef:    pref.DestinationRef,
		DestinationHash:   pref.DestinationHash,
		DestinationRegion: pref.DestinationRegion,
		PreferredDex:      pref.PreferredDEX,
		PreferredOffRamp:  pref.PreferredOffRamp,
		SlippageTolerance: pref.SlippageTolerance,
		CryptoToken:       toProtoTokenSpec(pref.CryptoToken),
		StableToken:       toProtoTokenSpec(pref.StableToken),
		CreatedAt:         unixTimestamp(pref.CreatedAt),
		UpdatedAt:         unixTimestamp(pref.UpdatedAt),
	}
}

func toProtoTokenSpec(spec types.TokenSpec) settlementv1.TokenSpec {
	return settlementv1.TokenSpec{
		Symbol:   spec.Symbol,
		Denom:    spec.Denom,
		Decimals: uint32(spec.Decimals),
		ChainId:  spec.ChainID,
	}
}

func unixTimestamp(timestamp time.Time) int64 {
	if timestamp.IsZero() {
		return 0
	}
	return timestamp.Unix()
}

func unixTimestampPtr(timestamp *time.Time) int64 {
	if timestamp == nil {
		return 0
	}
	return timestamp.Unix()
}
