// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	settlementtypes "github.com/virtengine/virtengine/x/settlement/types"
)

const (
	providerUsageProofLifetimeBlocks = int64(20)
	providerUsageProofLifetime       = 10 * time.Minute
)

// ProviderSigningStateResolver reads the actual governed x/provider active key
// and current committed chain height/time. Production usage submission fails
// closed when it is unavailable.
type ProviderSigningStateResolver interface {
	ResolveProviderSigningState(ctx context.Context, providerAddress string) (ActiveProviderKeyBinding, error)
}

// OnChainUsageStreamState is the committed sequence state for one canonical
// provider/allocation/order/lease stream.
type OnChainUsageStreamState struct {
	LastSequence uint64
}

// UsageStreamStateResolver reconciles producer sequence allocation with the
// authoritative replay state before a new proof is allocated.
type UsageStreamStateResolver interface {
	ResolveUsageStreamState(ctx context.Context, provider, allocationID, orderID, leaseID string) (OnChainUsageStreamState, error)
}

func (s *ChainUsageSubmitterImpl) prepareAuthenticatedUsageReport(ctx context.Context, report *ChainUsageReport) error {
	if report == nil {
		return ErrInvalidReport
	}
	if report.OrderID == "" || report.LeaseID == "" || report.UsageUnits == 0 || report.UsageType == "" ||
		report.PeriodEnd.Before(report.PeriodStart) || report.UnitPrice.Amount.IsZero() {
		return ErrInvalidReport
	}
	if s.cfg.ProviderSigningState == nil {
		return fmt.Errorf("provider signing state resolver not configured")
	}
	if s.keyManager == nil {
		return fmt.Errorf("key manager not configured")
	}
	if report.CustomerAddress == "" {
		return fmt.Errorf("customer address is required for authenticated usage")
	}
	binding, err := s.cfg.ProviderSigningState.ResolveProviderSigningState(ctx, s.cfg.ProviderAddress)
	if err != nil {
		return fmt.Errorf("resolve governed provider signing state: %w", err)
	}
	if binding.ProviderAddress != s.cfg.ProviderAddress || binding.BlockHeight <= 0 || binding.BlockTime.IsZero() {
		return ErrProviderKeyMismatch
	}

	identity, err := usageReportIdentity(s.cfg.ChainID, s.cfg.ProviderAddress, report)
	if err != nil {
		return err
	}
	streamID, err := settlementtypes.UsageStreamID(s.cfg.ProviderAddress, report.AllocationID, report.OrderID, report.LeaseID)
	if err != nil {
		return err
	}
	streamKey := hex.EncodeToString(streamID)
	identityKey := hex.EncodeToString(identity)
	var chainState OnChainUsageStreamState
	if s.cfg.UsageStreamState != nil {
		chainState, err = s.cfg.UsageStreamState.ResolveUsageStreamState(
			ctx,
			s.cfg.ProviderAddress,
			report.AllocationID,
			report.OrderID,
			report.LeaseID,
		)
		if err != nil {
			return fmt.Errorf("resolve committed usage stream state: %w", err)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.queueState == nil {
		s.queueState = newTxSubmissionQueueState()
	}
	allocation, exists := s.queueState.UsageProofs[identityKey]
	if exists && proofAllocationExpired(allocation, binding) && !s.proofAllocationQueued(allocation) {
		reservedSequence := allocation.Sequence
		allocation = nil
		exists = false
		if reservedSequence > 0 && s.queueState.UsageSequences[streamKey] == reservedSequence {
			s.queueState.UsageSequences[streamKey] = reservedSequence - 1
		}
	}
	if !exists {
		last := s.queueState.UsageSequences[streamKey]
		if chainState.LastSequence > last {
			last = chainState.LastSequence
		}
		if last == ^uint64(0) {
			return fmt.Errorf("usage stream sequence exhausted")
		}
		sequence := last + 1
		nonce, err := settlementtypes.DeriveReplayKey("virtengine.provider.usage.nonce.v1", s.cfg.ChainID, s.cfg.ProviderAddress, identityKey)
		if err != nil {
			return err
		}
		idempotencyKey, err := settlementtypes.DeriveReplayKey("virtengine.provider.usage.idempotency.v1", s.cfg.ChainID, s.cfg.ProviderAddress, identityKey)
		if err != nil {
			return err
		}
		allocation = &usageProofAllocation{
			StreamID:        streamKey,
			Sequence:        sequence,
			Nonce:           nonce,
			IdempotencyKey:  idempotencyKey,
			KeyID:           binding.KeyID,
			KeyEpoch:        binding.Epoch,
			IssuedAtHeight:  binding.BlockHeight,
			ExpiresAtHeight: binding.BlockHeight + providerUsageProofLifetimeBlocks,
			IssuedAtUnix:    binding.BlockTime.Unix(),
			ExpiresAtUnix:   binding.BlockTime.Add(providerUsageProofLifetime).Unix(),
		}
	}
	if allocation.KeyID != binding.KeyID || allocation.KeyEpoch != binding.Epoch {
		return ErrProviderKeyMismatch
	}

	report.ChainID = s.cfg.ChainID
	report.PricingVersion = defaultVersion(report.PricingVersion)
	report.FormulaVersion = defaultVersion(report.FormulaVersion)
	report.ModelVersion = defaultVersion(report.ModelVersion)
	report.StreamSequence = allocation.Sequence
	report.Nonce = append([]byte(nil), allocation.Nonce...)
	report.IdempotencyKey = append([]byte(nil), allocation.IdempotencyKey...)
	report.ProviderKeyEpoch = allocation.KeyEpoch
	report.ProviderKeyID = allocation.KeyID
	report.IssuedAtHeight = allocation.IssuedAtHeight
	report.ExpiresAtHeight = allocation.ExpiresAtHeight
	report.IssuedAtUnix = allocation.IssuedAtUnix
	report.ExpiresAtUnix = allocation.ExpiresAtUnix
	report.SignatureVersion = settlementtypes.SignatureVersionV1

	payload := canonicalPayloadForReport(s.cfg.ProviderAddress, report)
	signBytes, err := settlementtypes.CanonicalUsageSignBytes(payload)
	if err != nil {
		return err
	}
	signature, err := s.keyManager.SignForProviderKey(signBytes, binding)
	if err != nil {
		return err
	}
	report.Signature, err = hex.DecodeString(signature.Signature)
	if err != nil {
		return fmt.Errorf("decode provider signature: %w", err)
	}
	if !exists {
		previousSequence, hadSequence := s.queueState.UsageSequences[streamKey]
		previousProof, hadProof := s.queueState.UsageProofs[identityKey]
		s.queueState.UsageSequences[streamKey] = allocation.Sequence
		s.queueState.UsageProofs[identityKey] = allocation
		if err := s.queueStore.Save(s.queueState); err != nil {
			if hadSequence {
				s.queueState.UsageSequences[streamKey] = previousSequence
			} else {
				delete(s.queueState.UsageSequences, streamKey)
			}
			if hadProof {
				s.queueState.UsageProofs[identityKey] = previousProof
			} else {
				delete(s.queueState.UsageProofs, identityKey)
			}
			return err
		}
	}
	return nil
}

func proofAllocationExpired(allocation *usageProofAllocation, binding ActiveProviderKeyBinding) bool {
	return allocation == nil ||
		allocation.ExpiresAtHeight < binding.BlockHeight ||
		allocation.ExpiresAtUnix < binding.BlockTime.Unix()
}

func (s *ChainUsageSubmitterImpl) proofAllocationQueued(allocation *usageProofAllocation) bool {
	if allocation == nil || s.queueState == nil {
		return false
	}
	for _, item := range s.queueState.Items {
		if item.Kind != queueItemKindUsage {
			continue
		}
		var msg MsgRecordUsageWrapper
		if json.Unmarshal(item.Payload, &msg) == nil && bytes.Equal(msg.IdempotencyKey, allocation.IdempotencyKey) {
			return true
		}
	}
	return false
}

func canonicalPayloadForReport(provider string, report *ChainUsageReport) settlementtypes.CanonicalUsagePayload {
	return settlementtypes.CanonicalUsagePayload{
		SignatureVersion: report.SignatureVersion,
		ChainID:          report.ChainID,
		Domain:           settlementtypes.UsageProviderDomainV1,
		SignerRole:       settlementtypes.SignerRoleProvider,
		Provider:         provider,
		Customer:         report.CustomerAddress,
		OrderID:          report.OrderID,
		LeaseID:          report.LeaseID,
		AllocationID:     report.AllocationID,
		PeriodStart:      report.PeriodStart.Unix(),
		PeriodEnd:        report.PeriodEnd.Unix(),
		Metrics: settlementtypes.RawUsageMetrics{
			CPUMilliSeconds:    report.RawMetrics.CPUMilliSeconds,
			MemoryByteSeconds:  report.RawMetrics.MemoryByteSeconds,
			StorageByteSeconds: report.RawMetrics.StorageByteSeconds,
			NetworkBytesIn:     report.RawMetrics.NetworkBytesIn,
			NetworkBytesOut:    report.RawMetrics.NetworkBytesOut,
			GPUSeconds:         report.RawMetrics.GPUSeconds,
		},
		PricingVersion:   report.PricingVersion,
		UsageUnits:       report.UsageUnits,
		UsageType:        report.UsageType,
		UnitPriceDenom:   report.UnitPrice.Denom,
		UnitPriceAmount:  report.UnitPrice.Amount.String(),
		FormulaVersion:   report.FormulaVersion,
		ModelVersion:     report.ModelVersion,
		Sequence:         report.StreamSequence,
		Nonce:            report.Nonce,
		IdempotencyKey:   report.IdempotencyKey,
		ProviderKeyEpoch: report.ProviderKeyEpoch,
		ProviderKeyID:    report.ProviderKeyID,
		IssuedAtHeight:   report.IssuedAtHeight,
		ExpiresAtHeight:  report.ExpiresAtHeight,
		IssuedAtUnix:     report.IssuedAtUnix,
		ExpiresAtUnix:    report.ExpiresAtUnix,
	}
}

func usageReportIdentity(chainID, provider string, report *ChainUsageReport) ([]byte, error) {
	values := []string{
		chainID,
		provider,
		report.CustomerAddress,
		report.OrderID,
		report.LeaseID,
		report.AllocationID,
		report.UsageType,
		report.PeriodStart.UTC().Format(time.RFC3339Nano),
		report.PeriodEnd.UTC().Format(time.RFC3339Nano),
		strconv.FormatUint(report.UsageUnits, 10),
		report.UnitPrice.String(),
		strconv.FormatInt(report.RawMetrics.CPUMilliSeconds, 10),
		strconv.FormatInt(report.RawMetrics.MemoryByteSeconds, 10),
		strconv.FormatInt(report.RawMetrics.StorageByteSeconds, 10),
		strconv.FormatInt(report.RawMetrics.NetworkBytesIn, 10),
		strconv.FormatInt(report.RawMetrics.NetworkBytesOut, 10),
		strconv.FormatInt(report.RawMetrics.GPUSeconds, 10),
	}
	return settlementtypes.DeriveReplayKey("virtengine.provider.usage.identity.v1", values...)
}

func defaultVersion(version uint32) uint32 {
	if version == 0 {
		return 1
	}
	return version
}
