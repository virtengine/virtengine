// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	decredsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"

	providertypes "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	settlementkeeper "github.com/virtengine/virtengine/x/settlement/keeper"
	"github.com/virtengine/virtengine/x/settlement/types"
)

type task84BProviderKeys struct {
	owner  string
	record providertypes.ProviderPublicKeyRecord
}

func (m task84BProviderKeys) GetProviderSigningKey(_ sdk.Context, owner sdk.AccAddress, keyID string, epoch uint64) (providertypes.ProviderPublicKeyRecord, bool) {
	if owner.String() != m.owner || keyID != m.record.KeyID || epoch != m.record.Epoch {
		return providertypes.ProviderPublicKeyRecord{}, false
	}
	return m.record, true
}

type task84BAccounts struct{ account sdk.AccountI }

func (m task84BAccounts) GetAccount(_ context.Context, addr sdk.AccAddress) sdk.AccountI {
	if m.account == nil || !m.account.GetAddress().Equals(addr) {
		return nil
	}
	return m.account
}

func configureTask84BAuthentication(t *testing.T, s *KeeperTestSuite) (ed25519.PrivateKey, *cosmosed25519.PrivKey) {
	t.Helper()
	s.ctx = s.ctx.WithChainID("virtengine-test-1").WithBlockHeight(100).WithBlockTime(time.Unix(1_700_000_000, 0).UTC())
	providerPublic, providerPrivate, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	providerRecord := providertypes.NewProviderPublicKeyRecord(providerPublic, providertypes.PublicKeyTypeEd25519, s.ctx.BlockHeight())
	providerRecord.ActivatedAtUnix = s.ctx.BlockTime().Unix()

	customerPrivate := cosmosed25519.GenPrivKey()
	account := authtypes.NewBaseAccountWithAddress(s.depositor)
	require.NoError(t, account.SetPubKey(customerPrivate.PubKey()))

	s.keeper.SetUsageAuthenticationKeepers(
		task84BProviderKeys{owner: s.provider.String(), record: providerRecord},
		task84BAccounts{account: account},
	)
	require.NoError(t, s.keeper.ActivateUsageAuthentication(s.ctx))
	return providerPrivate, customerPrivate
}

//nolint:unparam // all current signed keeper vectors intentionally exercise the CPU formula
func task84BSignedUsage(t *testing.T, s *KeeperTestSuite, privateKey ed25519.PrivateKey, orderID, leaseID, allocationID, usageType string, sequence uint64, start, end time.Time) *types.UsageRecord {
	t.Helper()
	keyRecord, found := task84BProviderKeysForKeeper(s, privateKey)
	require.True(t, found)
	sequenceText := strconv.FormatUint(sequence, 10)
	nonce, err := types.DeriveReplayKey("task84b-test-nonce", orderID, leaseID, allocationID, usageType, start.String(), end.String(), sequenceText)
	require.NoError(t, err)
	idempotencyKey, err := types.DeriveReplayKey("task84b-test-idempotency", orderID, leaseID, allocationID, usageType, start.String(), end.String(), sequenceText)
	require.NoError(t, err)

	record := &types.UsageRecord{
		OrderID:          orderID,
		LeaseID:          leaseID,
		AllocationID:     allocationID,
		Provider:         s.provider.String(),
		Customer:         s.depositor.String(),
		UsageUnits:       10,
		UsageType:        usageType,
		PeriodStart:      start,
		PeriodEnd:        end,
		UnitPrice:        sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDec(2)),
		Metrics:          types.RawUsageMetrics{CPUMilliSeconds: 36_000_000},
		PricingVersion:   1,
		FormulaVersion:   1,
		ModelVersion:     1,
		Sequence:         sequence,
		Nonce:            nonce,
		IdempotencyKey:   idempotencyKey,
		ProviderKeyEpoch: keyRecord.Epoch,
		ProviderKeyID:    keyRecord.KeyID,
		IssuedAtHeight:   s.ctx.BlockHeight(),
		ExpiresAtHeight:  s.ctx.BlockHeight() + 20,
		IssuedAtUnix:     s.ctx.BlockTime().Unix(),
		ExpiresAtUnix:    s.ctx.BlockTime().Add(10 * time.Minute).Unix(),
		SignatureVersion: types.SignatureVersionV1,
	}
	signBytes, err := types.CanonicalUsageSignBytes(record.CanonicalUsagePayload(s.ctx.ChainID()))
	require.NoError(t, err)
	record.ProviderSignature = ed25519.Sign(privateKey, signBytes)
	return record
}

func task84BProviderKeysForKeeper(s *KeeperTestSuite, privateKey ed25519.PrivateKey) (providertypes.ProviderPublicKeyRecord, bool) {
	publicKey := privateKey.Public().(ed25519.PublicKey)
	record := providertypes.NewProviderPublicKeyRecord(publicKey, providertypes.PublicKeyTypeEd25519, s.ctx.BlockHeight())
	record.ActivatedAtUnix = s.ctx.BlockTime().Unix()
	return record, true
}

func (s *KeeperTestSuite) TestAuthenticatedUsageReplayAndSequencePolicy() {
	providerPrivate, _ := configureTask84BAuthentication(s.T(), s)
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "auth-order", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "auth-lease", s.provider))

	start := s.ctx.BlockTime().Add(-2 * time.Hour)
	first := task84BSignedUsage(s.T(), s, providerPrivate, "auth-order", "auth-lease", "allocation-1", "cpu", 1, start, s.ctx.BlockTime().Add(-time.Hour))
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, first))
	s.Require().True(first.SignatureVerified)
	s.Require().Len(first.UsageDigest, types.DigestSize)
	firstID := first.UsageID
	firstEventCount := len(s.ctx.EventManager().Events())

	retry := *first
	retry.UsageID = ""
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, &retry))
	s.Require().Equal(firstID, retry.UsageID)
	s.Require().Equal(firstEventCount, len(s.ctx.EventManager().Events()), "exact retry must emit no event")
	s.Require().Len(s.keeper.GetUsageRecordsByOrder(s.ctx, "auth-order"), 1)

	conflict := *first
	conflict.UsageID = ""
	conflict.Metrics.CPUMilliSeconds += 3_600_000
	conflict.UsageUnits++
	conflict.TotalCost = nil
	signBytes, err := types.CanonicalUsageSignBytes(conflict.CanonicalUsagePayload(s.ctx.ChainID()))
	s.Require().NoError(err)
	conflict.ProviderSignature = ed25519.Sign(providerPrivate, signBytes)
	err = s.keeper.RecordUsage(s.ctx, &conflict)
	s.Require().ErrorIs(err, types.ErrUsageReplayConflict)

	gap := task84BSignedUsage(s.T(), s, providerPrivate, "auth-order", "auth-lease", "allocation-1", "cpu", 3, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	err = s.keeper.RecordUsage(s.ctx, gap)
	s.Require().ErrorIs(err, types.ErrUsageSequenceGap)

	second := task84BSignedUsage(s.T(), s, providerPrivate, "auth-order", "auth-lease", "allocation-1", "cpu", 2, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, second))
	s.Require().Empty(s.keeper.ValidateUsageReplayIndexes(s.ctx))

	overlap := task84BSignedUsage(s.T(), s, providerPrivate, "auth-order", "auth-lease", "allocation-1", "cpu", 3, s.ctx.BlockTime().Add(-30*time.Minute), s.ctx.BlockTime())
	err = s.keeper.RecordUsage(s.ctx, overlap)
	s.Require().ErrorIs(err, types.ErrUsagePeriodOverlap)
}

func (s *KeeperTestSuite) TestAuthenticatedUsageRejectsWrongSignatureAndExpiredProof() {
	providerPrivate, _ := configureTask84BAuthentication(s.T(), s)
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "auth-negative", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "auth-negative-lease", s.provider))

	record := task84BSignedUsage(s.T(), s, providerPrivate, "auth-negative", "auth-negative-lease", "", "cpu", 1, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	record.ProviderSignature[0] ^= 0xff
	err = s.keeper.RecordUsage(s.ctx, record)
	s.Require().ErrorIs(err, types.ErrInvalidSignature)

	record = task84BSignedUsage(s.T(), s, providerPrivate, "auth-negative", "auth-negative-lease", "", "cpu", 1, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	record.ExpiresAtHeight = s.ctx.BlockHeight() - 1
	record.ExpiresAtUnix = s.ctx.BlockTime().Add(-time.Second).Unix()
	err = s.keeper.RecordUsage(s.ctx, record)
	s.Require().ErrorIs(err, types.ErrUsageProofExpired)
}

func (s *KeeperTestSuite) TestAuthenticatedUsageRejectsBoundFieldTampering() {
	providerPrivate, _ := configureTask84BAuthentication(s.T(), s)
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "tamper-order", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "tamper-lease", s.provider))

	base := task84BSignedUsage(s.T(), s, providerPrivate, "tamper-order", "tamper-lease", "allocation-tamper", "cpu", 1, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	tests := []struct {
		name   string
		mutate func(*types.UsageRecord)
		target error
	}{
		{"chain", func(r *types.UsageRecord) { r.ChainID = "other-chain" }, types.ErrInvalidSignature},
		{"provider", func(r *types.UsageRecord) { r.Provider = s.depositor.String() }, types.ErrProviderSigningKeyNotFound},
		{"customer", func(r *types.UsageRecord) { r.Customer = s.provider.String() }, types.ErrInvalidSignature},
		{"order", func(r *types.UsageRecord) { r.OrderID = "other-order" }, types.ErrInvalidSignature},
		{"lease", func(r *types.UsageRecord) { r.LeaseID = "other-lease" }, types.ErrInvalidSignature},
		{"allocation", func(r *types.UsageRecord) { r.AllocationID = "other-allocation" }, types.ErrInvalidSignature},
		{"metrics", func(r *types.UsageRecord) { r.Metrics.CPUMilliSeconds++ }, types.ErrInvalidUsageRecord},
		{"price", func(r *types.UsageRecord) { r.UnitPrice = sdk.NewDecCoinFromDec("uve", sdkmath.LegacyNewDec(3)) }, types.ErrInvalidSignature},
		{"total cost override", func(r *types.UsageRecord) { r.TotalCost = sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(999))) }, types.ErrInvalidUsageRecord},
		{"formula", func(r *types.UsageRecord) { r.FormulaVersion = 2 }, types.ErrUsagePricingVersion},
		{"key epoch", func(r *types.UsageRecord) { r.ProviderKeyEpoch++ }, types.ErrProviderSigningKeyNotFound},
		{"future proof", func(r *types.UsageRecord) { r.IssuedAtHeight = s.ctx.BlockHeight() + types.MaxProofFutureBlocks + 1 }, types.ErrUsageProofExpired},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			record := *base
			record.Nonce = append([]byte(nil), base.Nonce...)
			record.IdempotencyKey = append([]byte(nil), base.IdempotencyKey...)
			record.ProviderSignature = append([]byte(nil), base.ProviderSignature...)
			tc.mutate(&record)
			err := s.keeper.RecordUsage(s.ctx, &record)
			s.Require().ErrorIs(err, tc.target)
		})
	}
}

func (s *KeeperTestSuite) TestAuthenticatedUsageCannotBeCountedTwiceInSettlement() {
	providerPrivate, _ := configureTask84BAuthentication(s.T(), s)
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "settle-duplicate-order", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "settle-duplicate-lease", s.provider))
	record := task84BSignedUsage(s.T(), s, providerPrivate, "settle-duplicate-order", "settle-duplicate-lease", "allocation-settle", "cpu", 1, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, record))

	_, err = s.keeper.SettleOrder(s.ctx, record.OrderID, []string{record.UsageID, record.UsageID}, false)
	s.Require().ErrorIs(err, types.ErrUsageReplayConflict)
	stored, found := s.keeper.GetUsageRecord(s.ctx, record.UsageID)
	s.Require().True(found)
	s.Require().False(stored.Settled)
	s.Require().Empty(s.keeper.GetSettlementsByOrder(s.ctx, record.OrderID))
}

func (s *KeeperTestSuite) TestCustomerAcknowledgmentVerifiedAndIdempotent() {
	providerPrivate, customerPrivate := configureTask84BAuthentication(s.T(), s)
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "ack-auth-order", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "ack-auth-lease", s.provider))

	record := task84BSignedUsage(s.T(), s, providerPrivate, "ack-auth-order", "ack-auth-lease", "allocation-ack", "cpu", 1, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, record))

	replayKey, err := types.DeriveReplayKey("task84b-ack", record.UsageID)
	s.Require().NoError(err)
	proof := types.UsageAcknowledgmentProof{
		UsageDigest:      append([]byte(nil), record.UsageDigest...),
		ReplayKey:        replayKey,
		IssuedAtHeight:   s.ctx.BlockHeight(),
		ExpiresAtHeight:  s.ctx.BlockHeight() + 10,
		IssuedAtUnix:     s.ctx.BlockTime().Unix(),
		ExpiresAtUnix:    s.ctx.BlockTime().Add(5 * time.Minute).Unix(),
		SignatureVersion: types.SignatureVersionV1,
	}
	ackPayload := proof.CanonicalPayload(s.ctx.ChainID(), record.Customer, record.UsageID)
	ackBytes, err := types.CanonicalAcknowledgmentSignBytes(ackPayload)
	s.Require().NoError(err)
	proof.Signature, err = customerPrivate.Sign(ackBytes)
	s.Require().NoError(err)
	badProof := proof
	badProof.Signature = append([]byte(nil), proof.Signature...)
	badProof.Signature[0] ^= 0xff
	err = s.keeper.AcknowledgeUsageAuthenticated(s.ctx, record.UsageID, badProof)
	s.Require().ErrorIs(err, types.ErrInvalidSignature)

	s.Require().NoError(s.keeper.AcknowledgeUsageAuthenticated(s.ctx, record.UsageID, proof))
	events := len(s.ctx.EventManager().Events())
	s.Require().NoError(s.keeper.AcknowledgeUsageAuthenticated(s.ctx, record.UsageID, proof))
	s.Require().Equal(events, len(s.ctx.EventManager().Events()))

	conflict := proof
	conflict.ReplayKey = bytes.Repeat([]byte{0x44}, types.ReplayKeySize)
	err = s.keeper.AcknowledgeUsageAuthenticated(s.ctx, record.UsageID, conflict)
	s.Require().ErrorIs(err, types.ErrUsageReplayConflict)
}

func (s *KeeperTestSuite) TestCustomerSecp256k1AcknowledgmentRejectsHighS() {
	providerPrivate, _ := configureTask84BAuthentication(s.T(), s)
	customerPrivate := cosmossecp256k1.GenPrivKey()
	account := authtypes.NewBaseAccountWithAddress(s.depositor)
	s.Require().NoError(account.SetPubKey(customerPrivate.PubKey()))
	s.keeper.SetUsageAuthenticationKeepers(
		task84BProviderKeysForPrivate(s, providerPrivate),
		task84BAccounts{account: account},
	)

	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "ack-secp-order", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "ack-secp-lease", s.provider))
	record := task84BSignedUsage(s.T(), s, providerPrivate, "ack-secp-order", "ack-secp-lease", "allocation-secp", "cpu", 1, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, record))

	replayKey, err := types.DeriveReplayKey("task84b-ack-secp", record.UsageID)
	s.Require().NoError(err)
	proof := types.UsageAcknowledgmentProof{
		UsageDigest:      append([]byte(nil), record.UsageDigest...),
		ReplayKey:        replayKey,
		IssuedAtHeight:   s.ctx.BlockHeight(),
		ExpiresAtHeight:  s.ctx.BlockHeight() + 10,
		IssuedAtUnix:     s.ctx.BlockTime().Unix(),
		ExpiresAtUnix:    s.ctx.BlockTime().Add(5 * time.Minute).Unix(),
		SignatureVersion: types.SignatureVersionV1,
	}
	ackBytes, err := types.CanonicalAcknowledgmentSignBytes(proof.CanonicalPayload(s.ctx.ChainID(), record.Customer, record.UsageID))
	s.Require().NoError(err)
	proof.Signature, err = customerPrivate.Sign(ackBytes)
	s.Require().NoError(err)

	highS := append([]byte(nil), proof.Signature...)
	sValue := new(big.Int).SetBytes(highS[32:])
	sValue.Sub(decredsecp256k1.S256().N, sValue)
	sBytes := sValue.Bytes()
	for i := 32; i < 64; i++ {
		highS[i] = 0
	}
	copy(highS[64-len(sBytes):], sBytes)
	highProof := proof
	highProof.Signature = highS
	err = s.keeper.AcknowledgeUsageAuthenticated(s.ctx, record.UsageID, highProof)
	s.Require().ErrorIs(err, types.ErrInvalidSignature)

	s.Require().NoError(s.keeper.AcknowledgeUsageAuthenticated(s.ctx, record.UsageID, proof))
}

func (s *KeeperTestSuite) TestProviderSecp256k1UsageAcceptsLowSAndRejectsHighS() {
	s.ctx = s.ctx.WithChainID("virtengine-test-1").WithBlockHeight(100).WithBlockTime(time.Unix(1_700_000_000, 0).UTC())
	providerPrivate := cosmossecp256k1.GenPrivKey()
	providerRecord := providertypes.NewProviderPublicKeyRecord(providerPrivate.PubKey().Bytes(), providertypes.PublicKeyTypeSecp256k1, s.ctx.BlockHeight())
	providerRecord.ActivatedAtUnix = s.ctx.BlockTime().Unix()
	s.keeper.SetUsageAuthenticationKeepers(
		task84BProviderKeys{owner: s.provider.String(), record: providerRecord},
		task84BAccounts{},
	)
	s.Require().NoError(s.keeper.ActivateUsageAuthentication(s.ctx))

	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "provider-secp-order", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "provider-secp-lease", s.provider))

	newRecord := func(sequence uint64, start, end time.Time) *types.UsageRecord {
		sequenceText := strconv.FormatUint(sequence, 10)
		nonce, deriveErr := types.DeriveReplayKey("task84b-provider-secp-nonce", sequenceText)
		s.Require().NoError(deriveErr)
		idempotencyKey, deriveErr := types.DeriveReplayKey("task84b-provider-secp-idempotency", sequenceText)
		s.Require().NoError(deriveErr)
		return &types.UsageRecord{
			OrderID:          "provider-secp-order",
			LeaseID:          "provider-secp-lease",
			AllocationID:     "provider-secp-allocation",
			Provider:         s.provider.String(),
			Customer:         s.depositor.String(),
			UsageUnits:       1,
			UsageType:        "cpu",
			PeriodStart:      start,
			PeriodEnd:        end,
			UnitPrice:        sdk.NewDecCoinFromDec("uve", sdkmath.LegacyOneDec()),
			Metrics:          types.RawUsageMetrics{CPUMilliSeconds: 3_600_000},
			PricingVersion:   1,
			FormulaVersion:   1,
			ModelVersion:     1,
			Sequence:         sequence,
			Nonce:            nonce,
			IdempotencyKey:   idempotencyKey,
			ProviderKeyEpoch: providerRecord.Epoch,
			ProviderKeyID:    providerRecord.KeyID,
			IssuedAtHeight:   s.ctx.BlockHeight(),
			ExpiresAtHeight:  s.ctx.BlockHeight() + 10,
			IssuedAtUnix:     end.Unix(),
			ExpiresAtUnix:    end.Add(5 * time.Minute).Unix(),
			SignatureVersion: types.SignatureVersionV1,
		}
	}

	lowS := newRecord(1, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	signBytes, err := types.CanonicalUsageSignBytes(lowS.CanonicalUsagePayload(s.ctx.ChainID()))
	s.Require().NoError(err)
	lowS.ProviderSignature, err = providerPrivate.Sign(signBytes)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, lowS))

	highSEnd := s.ctx.BlockTime().Add(time.Hour)
	highS := newRecord(2, lowS.PeriodEnd, highSEnd)
	signBytes, err = types.CanonicalUsageSignBytes(highS.CanonicalUsagePayload(s.ctx.ChainID()))
	s.Require().NoError(err)
	highS.ProviderSignature, err = providerPrivate.Sign(signBytes)
	s.Require().NoError(err)
	sValue := new(big.Int).SetBytes(highS.ProviderSignature[32:])
	sValue.Sub(decredsecp256k1.S256().N, sValue)
	highSBytes := sValue.Bytes()
	for i := 32; i < 64; i++ {
		highS.ProviderSignature[i] = 0
	}
	copy(highS.ProviderSignature[64-len(highSBytes):], highSBytes)
	err = s.keeper.RecordUsage(s.ctx.WithBlockTime(highSEnd), highS)
	s.Require().ErrorIs(err, types.ErrInvalidSignature)
}

func (s *KeeperTestSuite) TestAuthenticatedUsageStateIsImmutableAfterVerification() {
	providerPrivate, _ := configureTask84BAuthentication(s.T(), s)
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "immutable-order", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "immutable-lease", s.provider))

	record := task84BSignedUsage(s.T(), s, providerPrivate, "immutable-order", "immutable-lease", "immutable-allocation", "cpu", 1, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, record))

	mutated := *record
	mutated.UsageUnits++
	mutated.Metrics.CPUMilliSeconds += 3_600_000
	err = s.keeper.SetUsageRecord(s.ctx, mutated)
	s.Require().ErrorIs(err, types.ErrInvalidUsageRecord)

	metadataMutated := *record
	metadataMutated.Metadata = map[string]string{"tampered": "true"}
	err = s.keeper.SetUsageRecord(s.ctx, metadataMutated)
	s.Require().ErrorIs(err, types.ErrInvalidUsageRecord)

	stored, found := s.keeper.GetUsageRecord(s.ctx, record.UsageID)
	s.Require().True(found)
	s.Require().Equal(record.UsageUnits, stored.UsageUnits)
	s.Require().Equal(record.UsageDigest, stored.UsageDigest)
}

func (s *KeeperTestSuite) TestReplayIndexDigestCannotAliasDifferentUsageID() {
	providerPrivate, _ := configureTask84BAuthentication(s.T(), s)
	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowID, err := s.keeper.CreateEscrow(s.ctx, "replay-corrupt-order", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowID, "replay-corrupt-lease", s.provider))

	record := task84BSignedUsage(s.T(), s, providerPrivate, "replay-corrupt-order", "replay-corrupt-lease", "replay-corrupt-allocation", "cpu", 1, s.ctx.BlockTime().Add(-time.Hour), s.ctx.BlockTime())
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, record))
	streamID, err := types.UsageStreamID(record.Provider, record.AllocationID, record.OrderID, record.LeaseID)
	s.Require().NoError(err)
	corrupt, err := json.Marshal(map[string]interface{}{
		"usage_id": "missing-usage-id",
		"digest":   record.UsageDigest,
	})
	s.Require().NoError(err)
	s.ctx.KVStore(s.storeKey).Set(types.UsageReplaySequenceKey(streamID, record.Sequence), corrupt)

	retry := *record
	retry.UsageID = ""
	err = s.keeper.RecordUsage(s.ctx, &retry)
	s.Require().ErrorIs(err, types.ErrUsageReplayConflict)
}

func (s *KeeperTestSuite) TestGlobalNonceCollisionAcrossProvidersIsRejected() {
	providerOnePublic, providerOnePrivate, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	providerTwoPublic, providerTwoPrivate, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)
	providerTwo := sdk.AccAddress(bytes.Repeat([]byte{0x7a}, 20))
	keys := task84BMultiProviderKeys{records: map[string]providertypes.ProviderPublicKeyRecord{}}
	for owner, publicKey := range map[string]ed25519.PublicKey{
		s.provider.String():  providerOnePublic,
		providerTwo.String(): providerTwoPublic,
	} {
		keyRecord := providertypes.NewProviderPublicKeyRecord(publicKey, providertypes.PublicKeyTypeEd25519, 100)
		keyRecord.ActivatedAtUnix = 1_700_000_000
		keys.records[owner] = keyRecord
	}
	s.ctx = s.ctx.WithChainID("virtengine-test-1").WithBlockHeight(100).WithBlockTime(time.Unix(1_700_000_000, 0).UTC())
	s.keeper.SetUsageAuthenticationKeepers(keys, task84BAccounts{})
	s.Require().NoError(s.keeper.ActivateUsageAuthentication(s.ctx))

	amount := sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10_000)))
	escrowOne, err := s.keeper.CreateEscrow(s.ctx, "nonce-order-one", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowOne, "nonce-lease-one", s.provider))
	escrowTwo, err := s.keeper.CreateEscrow(s.ctx, "nonce-order-two", s.depositor, amount, 24*time.Hour, nil)
	s.Require().NoError(err)
	s.Require().NoError(s.keeper.ActivateEscrow(s.ctx, escrowTwo, "nonce-lease-two", providerTwo))

	sharedNonce := bytes.Repeat([]byte{0x5a}, types.ReplayKeySize)
	makeRecord := func(provider sdk.AccAddress, privateKey ed25519.PrivateKey, orderID, leaseID, allocationID string) *types.UsageRecord {
		keyRecord := keys.records[provider.String()]
		idempotencyKey, deriveErr := types.DeriveReplayKey("task84b-global-nonce", orderID)
		s.Require().NoError(deriveErr)
		record := &types.UsageRecord{
			OrderID: orderID, LeaseID: leaseID, AllocationID: allocationID,
			Provider: provider.String(), Customer: s.depositor.String(),
			UsageUnits: 1, UsageType: "cpu",
			PeriodStart: s.ctx.BlockTime().Add(-time.Hour), PeriodEnd: s.ctx.BlockTime(),
			UnitPrice:      sdk.NewDecCoinFromDec("uve", sdkmath.LegacyOneDec()),
			Metrics:        types.RawUsageMetrics{CPUMilliSeconds: 3_600_000},
			PricingVersion: 1, FormulaVersion: 1, ModelVersion: 1, Sequence: 1,
			Nonce: append([]byte(nil), sharedNonce...), IdempotencyKey: idempotencyKey,
			ProviderKeyEpoch: keyRecord.Epoch, ProviderKeyID: keyRecord.KeyID,
			IssuedAtHeight: 100, ExpiresAtHeight: 110,
			IssuedAtUnix: s.ctx.BlockTime().Unix(), ExpiresAtUnix: s.ctx.BlockTime().Add(5 * time.Minute).Unix(),
			SignatureVersion: types.SignatureVersionV1,
		}
		signBytes, signErr := types.CanonicalUsageSignBytes(record.CanonicalUsagePayload(s.ctx.ChainID()))
		s.Require().NoError(signErr)
		record.ProviderSignature = ed25519.Sign(privateKey, signBytes)
		return record
	}

	first := makeRecord(s.provider, providerOnePrivate, "nonce-order-one", "nonce-lease-one", "nonce-allocation-one")
	s.Require().NoError(s.keeper.RecordUsage(s.ctx, first))
	second := makeRecord(providerTwo, providerTwoPrivate, "nonce-order-two", "nonce-lease-two", "nonce-allocation-two")
	err = s.keeper.RecordUsage(s.ctx, second)
	s.Require().ErrorIs(err, types.ErrUsageReplayConflict)
}

type task84BMultiProviderKeys struct {
	records map[string]providertypes.ProviderPublicKeyRecord
}

func (m task84BMultiProviderKeys) GetProviderSigningKey(_ sdk.Context, owner sdk.AccAddress, keyID string, epoch uint64) (providertypes.ProviderPublicKeyRecord, bool) {
	record, found := m.records[owner.String()]
	if !found || record.KeyID != keyID || record.Epoch != epoch {
		return providertypes.ProviderPublicKeyRecord{}, false
	}
	return record, true
}

func task84BProviderKeysForPrivate(s *KeeperTestSuite, privateKey ed25519.PrivateKey) task84BProviderKeys {
	record, _ := task84BProviderKeysForKeeper(s, privateKey)
	return task84BProviderKeys{owner: s.provider.String(), record: record}
}

var _ settlementkeeper.ProviderSigningKeyKeeper = task84BProviderKeys{}
var _ settlementkeeper.ProviderSigningKeyKeeper = task84BMultiProviderKeys{}
var _ settlementkeeper.AccountKeeper = task84BAccounts{}
