package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

// FiatConversionMigrationReport is the deterministic v1.8.0 reconciliation summary.
type FiatConversionMigrationReport struct {
	Scanned           uint64 `json:"scanned"`
	TerminalPreserved uint64 `json:"terminal_preserved"`
	ActiveQuarantined uint64 `json:"active_quarantined"`
	AlreadyMigrated   uint64 `json:"already_migrated"`
	Malformed         uint64 `json:"malformed"`
	Digest            string `json:"digest"`
}

func (k Keeper) MigrateFiatConversions(ctx sdk.Context) (FiatConversionMigrationReport, error) {
	cacheCtx, write := ctx.CacheContext()
	report, err := k.migrateFiatConversions(cacheCtx)
	if err != nil {
		return report, err
	}
	write()
	return report, nil
}

// RebuildFiatConversionState reconstructs derived indexes and accounting for
// already protocol-migrated genesis records. It never moves coins or infers a
// custody effect that is not explicitly certified by conversion and payout.
func (k Keeper) RebuildFiatConversionState(ctx sdk.Context) error {
	cacheCtx, write := ctx.CacheContext()
	store := cacheCtx.KVStore(k.skey)
	clearPrefixes(store,
		types.PrefixFiatConversionByInvoice, types.PrefixFiatConversionBySettlement,
		types.PrefixFiatConversionByPayout, types.PrefixFiatConversionByProvider,
		types.PrefixFiatConversionByState, types.PrefixFiatConversionIdempotency,
		types.PrefixFiatObservationReplay, types.PrefixFiatObservationSequence,
		types.PrefixFiatConversionRequestDigest, types.PrefixFiatDailyTotals,
		types.PrefixFiatCustodyEffect,
	)
	records := make([]types.FiatConversionRecord, 0)
	k.WithFiatConversions(cacheCtx, func(record types.FiatConversionRecord) bool {
		records = append(records, record)
		return false
	})
	sort.Slice(records, func(i, j int) bool { return records[i].ConversionID < records[j].ConversionID })
	for _, record := range records {
		if record.ProtocolVersion == 0 || len(record.RequestDigest) != sha256.Size || record.DailyBucket == "" {
			return types.ErrFiatConversionQuarantined.Wrapf("conversion %s is not protocol-migrated", record.ConversionID)
		}
		if err := k.setFiatConversion(cacheCtx, record, true); err != nil {
			return err
		}
		store.Set(types.FiatConversionRequestDigestKey(record.IdempotencyKey), record.RequestDigest)
		if record.DailyQuotaReserved {
			totalKey := types.FiatDailyTotalKey(record.Provider, record.DailyBucket)
			current, err := decodeDailyTotal(store.Get(totalKey))
			if err != nil {
				return err
			}
			store.Set(totalKey, encodeDailyTotal(current.Add(record.CryptoAmount.Amount)))
		}
		for _, observation := range record.Observations {
			if len(observation.IdempotencyKey) != sha256.Size || len(observation.ObservationDigest) != sha256.Size {
				return types.ErrFiatObservationEvidence.Wrapf("malformed observation lineage in %s", record.ConversionID)
			}
			store.Set(types.FiatObservationReplayKey(record.ConversionID, observation.IdempotencyKey), observation.ObservationDigest)
			store.Set(types.FiatObservationSequenceKey(record.ConversionID, observation.Sequence), observation.ObservationDigest)
		}
		if record.ValueMovementApplied && !record.LegacyQuarantined {
			payout, found := k.GetPayout(cacheCtx, record.PayoutID)
			if !found || payout.State != types.PayoutStateCompleted || !payout.ValueMovementApplied ||
				!bytes.Equal(payout.ValueMovementEffectHash, record.CustodySinkEffectHash) ||
				!bytes.Equal(fiatCustodyEffectHash(record, payout, record.PayoutFinalityHash), record.CustodySinkEffectHash) {
				return types.ErrInvalidSettlement.Wrapf("conversion %s custody effect is not certified", record.ConversionID)
			}
			store.Set(types.FiatCustodyEffectKey(record.ConversionID), record.CustodySinkEffectHash)
		}
	}
	if broken := k.ValidateFiatConversionInvariants(cacheCtx); len(broken) != 0 {
		return types.ErrInvalidSettlement.Wrapf("fiat conversion genesis invariant: %v", broken)
	}
	write()
	return nil
}

func (k Keeper) migrateFiatConversions(ctx sdk.Context) (FiatConversionMigrationReport, error) {
	store := ctx.KVStore(k.skey)
	clearPrefixes(store,
		types.PrefixFiatConversionByInvoice, types.PrefixFiatConversionBySettlement,
		types.PrefixFiatConversionByPayout, types.PrefixFiatConversionByProvider,
		types.PrefixFiatConversionByState, types.PrefixFiatConversionIdempotency,
		types.PrefixFiatObservationReplay, types.PrefixFiatObservationSequence,
		types.PrefixFiatConversionRequestDigest, types.PrefixFiatDailyTotals,
	)
	var report FiatConversionMigrationReport
	var records []types.FiatConversionRecord
	k.WithFiatConversions(ctx, func(record types.FiatConversionRecord) bool {
		records = append(records, record)
		return false
	})
	sort.Slice(records, func(i, j int) bool { return records[i].ConversionID < records[j].ConversionID })
	digestHasher := sha256.New()
	for i := range records {
		record := &records[i]
		report.Scanned++
		if record.ProtocolVersion >= fiatConversionProtocolVersion {
			report.AlreadyMigrated++
			if record.State == types.FiatConversionStatePayoutCompleted {
				payout, payoutFound := k.GetPayout(ctx, record.PayoutID)
				validExplicitEffect := payoutFound && payout.State == types.PayoutStateCompleted && payout.ValueMovementApplied && record.ValueMovementApplied &&
					record.CustodySinkAmount.IsEqual(record.CryptoAmount) && len(record.CustodySinkEffectHash) == sha256.Size &&
					bytes.Equal(payout.ValueMovementEffectHash, record.CustodySinkEffectHash) &&
					bytes.Equal(fiatCustodyEffectHash(*record, payout, record.PayoutFinalityHash), record.CustodySinkEffectHash)
				if validExplicitEffect {
					// Genesis export stores the immutable effect hash in both owners;
					// rebuild the derived lookup marker without moving coins again.
					store.Set(types.FiatCustodyEffectKey(record.ConversionID), record.CustodySinkEffectHash)
				} else {
					record.LegacyQuarantined = true
					record.QuarantineReason = "completed_conversion_missing_deterministic_custody_sink_effect"
					record.ValueMovementApplied = false
					record.CustodySinkAmount = sdk.Coin{}
					record.CustodySinkEffectHash = nil
					if payoutFound && (payout.ValueMovementApplied || len(payout.ValueMovementEffectHash) != 0) {
						payout.ValueMovementApplied = false
						payout.ValueMovementEffectHash = nil
						if err := k.SetPayout(ctx, payout); err != nil {
							return report, err
						}
					}
					store.Delete(types.FiatCustodyEffectKey(record.ConversionID))
				}
			}
		} else if record.State.IsTerminal() {
			record.ProtocolVersion = fiatConversionProtocolVersion
			record.LegacyQuarantined = true
			record.QuarantineReason = "legacy_terminal_conversion_preserved_without_authenticated_observation_certification"
			if record.PayoutID == "" && record.InvoiceID != "" {
				if payout, found := k.GetPayoutByInvoice(ctx, record.InvoiceID); found && payout.SettlementID == record.SettlementID && payout.Provider == record.Provider && payout.Customer == record.Customer {
					record.PayoutID = payout.PayoutID
				}
			}
			if record.PayoutID != "" {
				if payout, found := k.GetPayout(ctx, record.PayoutID); found {
					payout.ValueMovementApplied = false
					payout.ValueMovementEffectHash = nil
					if payout.FiatConversionID == "" {
						payout.FiatConversionID = record.ConversionID
						if err := k.SetPayout(ctx, payout); err != nil {
							return report, err
						}
					}
					if record.State == types.FiatConversionStatePayoutCompleted && payout.State == types.PayoutStateCompleted {
						if len(record.PayoutFinalityHash) == sha256.Size {
							payout.TxHash = ""
							payout.ExternalFinalityHash = append([]byte(nil), record.PayoutFinalityHash...)
							if err := k.SetPayout(ctx, payout); err != nil {
								return report, err
							}
						}
						record.ValueMovementApplied = false
						record.CustodySinkAmount = sdk.Coin{}
						record.CustodySinkEffectHash = nil
					}
				}
			}
			report.TerminalPreserved++
		} else {
			record.ProtocolVersion = fiatConversionProtocolVersion
			record.LegacyQuarantined = true
			record.QuarantineReason = "legacy_active_conversion_missing_authenticated_observation_lineage"
			record.TerminalPolicy = terminalPolicyManualReview
			report.ActiveQuarantined++
		}
		if record.IdempotencyKey == "" {
			record.IdempotencyKey = record.DefaultIdempotencyKey()
		}
		if len(record.RequestDigest) != 32 {
			legacyDigest := sha256.Sum256([]byte(strings.Join([]string{record.ConversionID, record.IdempotencyKey, record.Provider, record.Customer, record.CryptoAmount.String()}, "\x00")))
			record.RequestDigest = legacyDigest[:]
		}
		if record.DailyBucket == "" {
			record.DailyBucket = dailyFiatBucket(record.RequestedAt)
		}
		if record.ProtocolVersion == fiatConversionProtocolVersion && record.SlippageToleranceExact == "" {
			record.SlippageToleranceExact = fmt.Sprintf("%.18f", record.SlippageTolerance)
		}
		if record.ProtocolVersion == fiatConversionProtocolVersion && record.State != types.FiatConversionStateCancelled && !record.LegacyQuarantined && !record.DailyQuotaReserved {
			record.DailyQuotaReserved = true
		}
		if err := k.setFiatConversion(ctx, *record, true); err != nil {
			report.Malformed++
			return report, err
		}
		store.Set(types.FiatConversionRequestDigestKey(record.IdempotencyKey), record.RequestDigest)
		if record.DailyQuotaReserved {
			totalKey := types.FiatDailyTotalKey(record.Provider, record.DailyBucket)
			current, decodeErr := decodeDailyTotal(store.Get(totalKey))
			if decodeErr != nil {
				return report, decodeErr
			}
			store.Set(totalKey, encodeDailyTotal(current.Add(record.CryptoAmount.Amount)))
		}
		for _, observation := range record.Observations {
			if len(observation.IdempotencyKey) != 32 || len(observation.ObservationDigest) != 32 {
				report.Malformed++
				return report, types.ErrFiatObservationEvidence.Wrapf("malformed observation lineage in %s", record.ConversionID)
			}
			store.Set(types.FiatObservationReplayKey(record.ConversionID, observation.IdempotencyKey), observation.ObservationDigest)
			store.Set(types.FiatObservationSequenceKey(record.ConversionID, observation.Sequence), observation.ObservationDigest)
		}
		_, _ = digestHasher.Write([]byte(record.ConversionID))
		_, _ = digestHasher.Write(record.RequestDigest)
	}
	report.Digest = hex.EncodeToString(digestHasher.Sum(nil))
	bz, err := json.Marshal(report)
	if err != nil {
		return report, err
	}
	store.Set(types.FiatConversionMigrationAuditKey(), bz)
	return report, nil
}

func clearPrefixes(store storetypes.KVStore, prefixes ...[]byte) {
	for _, prefix := range prefixes {
		iter := storetypes.KVStorePrefixIterator(store, prefix)
		var keys [][]byte
		for ; iter.Valid(); iter.Next() {
			keys = append(keys, append([]byte(nil), iter.Key()...))
		}
		iter.Close()
		for _, key := range keys {
			store.Delete(key)
		}
	}
}

// ValidateFiatConversionInvariants checks records, all replay/reference
// indexes, profile commitments, daily accounting, payout linkage, and terminal
// contradictions without consulting wall clock or external systems.
func (k Keeper) ValidateFiatConversionInvariants(ctx sdk.Context) []string {
	store := ctx.KVStore(k.skey)
	params := k.GetParams(ctx)
	var broken []string
	payouts := make(map[string]types.PayoutRecord)
	expectedCustody := sdk.NewCoins()
	k.WithPayouts(ctx, func(payout types.PayoutRecord) bool {
		payouts[payout.PayoutID] = payout
		if err := payout.Validate(); err != nil {
			broken = append(broken, payout.PayoutID+": malformed payout")
			return false
		}
		for _, index := range []struct {
			name, value string
			key         []byte
		}{
			{"invoice", payout.InvoiceID, types.PayoutByInvoiceKey(payout.InvoiceID)},
			{"settlement", payout.SettlementID, types.PayoutBySettlementKey(payout.SettlementID)},
			{"idempotency", payout.IdempotencyKey, types.PayoutIdempotencyKey(payout.IdempotencyKey)},
		} {
			if index.value != "" && !bytes.Equal(store.Get(index.key), []byte(payout.PayoutID)) {
				broken = append(broken, payout.PayoutID+": payout "+index.name+" index mismatch")
			}
		}
		if !store.Has(types.PayoutByProviderKey(payout.Provider, payout.PayoutID)) || !store.Has(types.PayoutByStateKey(payout.State, payout.PayoutID)) {
			broken = append(broken, payout.PayoutID+": payout provider/state index mismatch")
		}
		return false
	})
	expectedDaily := make(map[string]sdkmath.Int)
	k.WithFiatConversions(ctx, func(record types.FiatConversionRecord) bool {
		if err := record.Validate(); err != nil {
			broken = append(broken, record.ConversionID+": malformed record")
			return false
		}
		if record.ProtocolVersion > 0 {
			if len(record.RequestDigest) != 32 || record.DailyBucket == "" {
				broken = append(broken, record.ConversionID+": request digest or daily bucket missing")
			}
			if owner := store.Get(types.FiatConversionRequestDigestKey(record.IdempotencyKey)); !bytes.Equal(owner, record.RequestDigest) {
				broken = append(broken, record.ConversionID+": request digest index mismatch")
			}
		}
		for _, index := range []struct {
			name, value string
			key         []byte
		}{
			{"invoice", record.InvoiceID, types.FiatConversionByInvoiceKey(record.InvoiceID)}, {"settlement", record.SettlementID, types.FiatConversionBySettlementKey(record.SettlementID)},
			{"payout", record.PayoutID, types.FiatConversionByPayoutKey(record.PayoutID)}, {"idempotency", record.IdempotencyKey, types.FiatConversionIdempotencyKey(record.IdempotencyKey)},
		} {
			if index.value != "" && !bytes.Equal(store.Get(index.key), []byte(record.ConversionID)) {
				broken = append(broken, record.ConversionID+": "+index.name+" index mismatch")
			}
		}
		if record.PayoutID != "" {
			payout, found := k.GetPayout(ctx, record.PayoutID)
			if !found || payout.FiatConversionID != record.ConversionID || payout.Provider != record.Provider || payout.Customer != record.Customer ||
				payout.InvoiceID != record.InvoiceID || payout.SettlementID != record.SettlementID || len(payout.NetAmount) != 1 || !payout.NetAmount[0].IsEqual(record.CryptoAmount) {
				broken = append(broken, record.ConversionID+": orphan payout linkage")
			} else if record.State == types.FiatConversionStatePayoutCompleted && !record.LegacyQuarantined && (payout.State != types.PayoutStateCompleted || !record.ValueMovementApplied || !payout.ValueMovementApplied || len(record.CustodySinkEffectHash) != sha256.Size || !record.CustodySinkAmount.IsEqual(record.CryptoAmount) || !bytes.Equal(payout.ValueMovementEffectHash, record.CustodySinkEffectHash) || !bytes.Equal(store.Get(types.FiatCustodyEffectKey(record.ConversionID)), record.CustodySinkEffectHash) || len(payout.ExternalFinalityHash) != sha256.Size || payout.TxHash != "" || !bytes.Equal(payout.ExternalFinalityHash, record.PayoutFinalityHash)) {
				broken = append(broken, record.ConversionID+": terminal completion contradiction")
			} else if (record.State == types.FiatConversionStateFailed || record.State == types.FiatConversionStateCancelled) && payout.State == types.PayoutStateCompleted {
				broken = append(broken, record.ConversionID+": failed conversion has completed payout")
			}
		}
		if record.ValueMovementApplied && !record.LegacyQuarantined {
			expectedCustody = expectedCustody.Add(record.CustodySinkAmount)
		}
		if record.ProtocolVersion > 0 && !record.LegacyQuarantined && !record.State.IsTerminal() {
			if record.DEXProfileID == "" || record.PayoutProfileID == "" || len(record.DEXProfileDigest) != 32 || len(record.PayoutProfileDigest) != 32 {
				broken = append(broken, record.ConversionID+": profile commitments malformed")
			}
		}
		if record.ProtocolVersion > 0 && record.SlippageToleranceExact == "" {
			broken = append(broken, record.ConversionID+": exact slippage missing")
		}
		if record.DailyQuotaReserved && (record.DailyBucket == "" || record.State == types.FiatConversionStateCancelled) {
			broken = append(broken, record.ConversionID+": daily quota reservation contradiction")
		}
		if record.ObservationSequence != uint64(len(record.Observations)) {
			broken = append(broken, record.ConversionID+": observation sequence/history mismatch")
		}
		var previous []byte
		seenObservationKeys := make(map[string]struct{}, len(record.Observations))
		for index, observation := range record.Observations {
			expectedSequence := uint64(index + 1) //nolint:gosec // observation history is bounded to 64 entries
			observationKey := string(observation.IdempotencyKey)
			_, duplicatedKey := seenObservationKeys[observationKey]
			seenObservationKeys[observationKey] = struct{}{}
			if len(observation.IdempotencyKey) != sha256.Size || len(observation.ObservationDigest) != sha256.Size || duplicatedKey || observation.Sequence != expectedSequence || !bytes.Equal(store.Get(types.FiatObservationReplayKey(record.ConversionID, observation.IdempotencyKey)), observation.ObservationDigest) || !bytes.Equal(store.Get(types.FiatObservationSequenceKey(record.ConversionID, observation.Sequence)), observation.ObservationDigest) {
				broken = append(broken, record.ConversionID+": observation replay index mismatch")
				break
			}
			lineage := sha256.Sum256(append(append([]byte(nil), previous...), observation.ObservationDigest...))
			if !bytes.Equal(lineage[:], observation.LineageDigest) {
				broken = append(broken, record.ConversionID+": observation lineage digest mismatch")
				break
			}
			previous = observation.ObservationDigest
		}
		if !bytes.Equal(previous, record.LastObservationDigest) && len(record.Observations) > 0 {
			broken = append(broken, record.ConversionID+": last observation digest mismatch")
		}
		if record.DailyQuotaReserved {
			dailyMapKey := record.Provider + "\x00" + record.DailyBucket
			current, exists := expectedDaily[dailyMapKey]
			if !exists {
				current = sdkmath.ZeroInt()
			}
			expectedDaily[dailyMapKey] = current.Add(record.CryptoAmount.Amount)
		}
		return false
	})
	dailyKeys := make([]string, 0, len(expectedDaily))
	for key := range expectedDaily {
		dailyKeys = append(dailyKeys, key)
	}
	sort.Strings(dailyKeys)
	for _, key := range dailyKeys {
		expected := expectedDaily[key]
		parts := strings.SplitN(key, "\x00", 2)
		actual, decodeErr := sdkmath.ZeroInt(), error(nil)
		if len(parts) == 2 {
			actual, decodeErr = decodeDailyTotal(store.Get(types.FiatDailyTotalKey(parts[0], parts[1])))
		}
		if len(parts) != 2 || decodeErr != nil || !actual.Equal(expected) {
			broken = append(broken, fmt.Sprintf("daily accounting mismatch for %q", key))
		}
	}
	if params.FiatConversionEnabled {
		if err := params.Validate(); err != nil {
			broken = append(broken, "enabled profile parameters invalid")
		}
	}
	for _, scan := range []struct {
		name   string
		prefix []byte
	}{
		{"invoice", types.PrefixFiatConversionByInvoice}, {"settlement", types.PrefixFiatConversionBySettlement},
		{"payout", types.PrefixFiatConversionByPayout}, {"idempotency", types.PrefixFiatConversionIdempotency},
		{fiatIndexProvider, types.PrefixFiatConversionByProvider}, {fiatIndexState, types.PrefixFiatConversionByState},
	} {
		iter := storetypes.KVStorePrefixIterator(store, scan.prefix)
		for ; iter.Valid(); iter.Next() {
			conversionID := string(iter.Value())
			if conversionID == "" && (scan.name == fiatIndexProvider || scan.name == fiatIndexState) {
				key := iter.Key()[len(scan.prefix):]
				separator := bytes.IndexByte(key, '/')
				if separator <= 0 || separator == len(key)-1 || bytes.IndexByte(key[separator+1:], '/') >= 0 {
					broken = append(broken, fmt.Sprintf("malformed %s index %x", scan.name, iter.Key()))
					continue
				}
				conversionID = string(key[separator+1:])
			}
			record, found := k.GetFiatConversion(ctx, conversionID)
			if !found {
				broken = append(broken, fmt.Sprintf("orphan %s index %x", scan.name, iter.Key()))
			} else if scan.name == fiatIndexProvider && !bytes.Equal(iter.Key(), types.FiatConversionByProviderKey(record.Provider, record.ConversionID)) {
				broken = append(broken, fmt.Sprintf("contradictory provider index %x", iter.Key()))
			} else if scan.name == fiatIndexState && !bytes.Equal(iter.Key(), types.FiatConversionByStateKey(record.State, record.ConversionID)) {
				broken = append(broken, fmt.Sprintf("contradictory state index %x", iter.Key()))
			}
		}
		iter.Close()
	}
	for _, scan := range []struct {
		name   string
		prefix []byte
	}{
		{"request digest", types.PrefixFiatConversionRequestDigest}, {"observation replay", types.PrefixFiatObservationReplay},
		{"observation sequence", types.PrefixFiatObservationSequence},
	} {
		iter := storetypes.KVStorePrefixIterator(store, scan.prefix)
		for ; iter.Valid(); iter.Next() {
			if len(iter.Value()) != 32 {
				broken = append(broken, fmt.Sprintf("malformed %s index %x", scan.name, iter.Key()))
			}
		}
		iter.Close()
	}
	dailyIter := storetypes.KVStorePrefixIterator(store, types.PrefixFiatDailyTotals)
	for ; dailyIter.Valid(); dailyIter.Next() {
		key := dailyIter.Key()[len(types.PrefixFiatDailyTotals):]
		separator := bytes.LastIndexByte(key, '/')
		if separator <= 0 || len(key)-separator-1 != 8 || bytes.IndexByte(key[:separator], '/') >= 0 {
			broken = append(broken, fmt.Sprintf("malformed daily accounting owner %x", dailyIter.Key()))
		}
		if total, ok := sdkmath.NewIntFromString(string(dailyIter.Value())); !ok || total.IsNegative() {
			broken = append(broken, fmt.Sprintf("malformed daily accounting index %x", dailyIter.Key()))
		}
	}
	dailyIter.Close()
	requestIter := storetypes.KVStorePrefixIterator(store, types.PrefixFiatConversionRequestDigest)
	for ; requestIter.Valid(); requestIter.Next() {
		key := requestIter.Key()[len(types.PrefixFiatConversionRequestDigest):]
		if len(key) < 2 || key[len(key)-1] != 0 || bytes.IndexByte(key[:len(key)-1], 0) >= 0 {
			broken = append(broken, fmt.Sprintf("malformed request digest owner %x", requestIter.Key()))
			continue
		}
		idempotency := string(key[:len(key)-1])
		owner := store.Get(types.FiatConversionIdempotencyKey(idempotency))
		record, found := k.GetFiatConversion(ctx, string(owner))
		if !found || !bytes.Equal(record.RequestDigest, requestIter.Value()) {
			broken = append(broken, fmt.Sprintf("orphan request digest index %x", requestIter.Key()))
		}
	}
	requestIter.Close()
	for _, scan := range []struct {
		name   string
		prefix []byte
	}{
		{"observation replay", types.PrefixFiatObservationReplay}, {"observation sequence", types.PrefixFiatObservationSequence},
	} {
		iter := storetypes.KVStorePrefixIterator(store, scan.prefix)
		for ; iter.Valid(); iter.Next() {
			key := iter.Key()[len(scan.prefix):]
			separator := bytes.IndexByte(key, 0)
			if separator <= 0 || separator == len(key)-1 {
				broken = append(broken, fmt.Sprintf("malformed %s owner %x", scan.name, iter.Key()))
				continue
			}
			record, found := k.GetFiatConversion(ctx, string(key[:separator]))
			if !found {
				broken = append(broken, fmt.Sprintf("orphan %s index %x", scan.name, iter.Key()))
				continue
			}
			matched := false
			for _, observation := range record.Observations {
				expected := types.FiatObservationReplayKey(record.ConversionID, observation.IdempotencyKey)
				if scan.name == "observation sequence" {
					expected = types.FiatObservationSequenceKey(record.ConversionID, observation.Sequence)
				}
				if bytes.Equal(expected, iter.Key()) && bytes.Equal(observation.ObservationDigest, iter.Value()) {
					matched = true
					break
				}
			}
			if !matched {
				broken = append(broken, fmt.Sprintf("unowned %s index %x", scan.name, iter.Key()))
			}
		}
		iter.Close()
	}
	dailyOwnerIter := storetypes.KVStorePrefixIterator(store, types.PrefixFiatDailyTotals)
	for ; dailyOwnerIter.Valid(); dailyOwnerIter.Next() {
		matched := false
		for _, key := range dailyKeys {
			parts := strings.SplitN(key, "\x00", 2)
			if len(parts) == 2 && bytes.Equal(dailyOwnerIter.Key(), types.FiatDailyTotalKey(parts[0], parts[1])) {
				matched = true
				break
			}
		}
		if !matched {
			broken = append(broken, fmt.Sprintf("orphan daily accounting index %x", dailyOwnerIter.Key()))
		}
	}
	dailyOwnerIter.Close()
	for _, scan := range []struct {
		name   string
		prefix []byte
	}{
		{"payout provider", types.PrefixPayoutByProvider}, {"payout state", types.PrefixPayoutByState},
	} {
		iter := storetypes.KVStorePrefixIterator(store, scan.prefix)
		for ; iter.Valid(); iter.Next() {
			key := iter.Key()[len(scan.prefix):]
			separator := bytes.IndexByte(key, '/')
			if separator <= 0 || separator == len(key)-1 || bytes.IndexByte(key[separator+1:], '/') >= 0 {
				broken = append(broken, fmt.Sprintf("malformed %s index %x", scan.name, iter.Key()))
				continue
			}
			payout, found := payouts[string(key[separator+1:])]
			if !found {
				broken = append(broken, fmt.Sprintf("orphan %s index %x", scan.name, iter.Key()))
				continue
			}
			expected := types.PayoutByProviderKey(payout.Provider, payout.PayoutID)
			if scan.name == "payout state" {
				expected = types.PayoutByStateKey(payout.State, payout.PayoutID)
			}
			if !bytes.Equal(expected, iter.Key()) {
				broken = append(broken, fmt.Sprintf("contradictory %s index %x", scan.name, iter.Key()))
			}
		}
		iter.Close()
	}
	ledgerIter := storetypes.KVStorePrefixIterator(store, types.PrefixPayoutLedgerByPayout)
	for ; ledgerIter.Valid(); ledgerIter.Next() {
		key := ledgerIter.Key()[len(types.PrefixPayoutLedgerByPayout):]
		separator := bytes.IndexByte(key, '/')
		entryID := string(ledgerIter.Value())
		if separator <= 0 || separator == len(key)-1 || string(key[separator+1:]) != entryID {
			broken = append(broken, fmt.Sprintf("malformed payout ledger index %x", ledgerIter.Key()))
			continue
		}
		if _, found := payouts[string(key[:separator])]; !found || !store.Has(types.PayoutLedgerEntryKey(entryID)) {
			broken = append(broken, fmt.Sprintf("orphan payout ledger index %x", ledgerIter.Key()))
		}
	}
	ledgerIter.Close()
	entryIter := storetypes.KVStorePrefixIterator(store, types.PrefixPayoutLedgerEntry)
	for ; entryIter.Valid(); entryIter.Next() {
		var entry types.PayoutLedgerEntry
		if err := json.Unmarshal(entryIter.Value(), &entry); err != nil || entry.EntryID == "" || entry.PayoutID == "" || entry.EntryID != string(entryIter.Key()[len(types.PrefixPayoutLedgerEntry):]) {
			broken = append(broken, fmt.Sprintf("malformed payout ledger entry %x", entryIter.Key()))
			continue
		}
		if _, found := payouts[entry.PayoutID]; !found || !store.Has(types.PayoutLedgerByPayoutKey(entry.PayoutID, entry.EntryID)) {
			broken = append(broken, fmt.Sprintf("orphan payout ledger entry %x", entryIter.Key()))
		}
	}
	entryIter.Close()
	broken = append(broken, k.validateCompletedFiatTreasuryAccounting(ctx, payouts)...)
	actualCustody := k.GetFiatConversionCustodyBalance(ctx)
	if !actualCustody.Equal(expectedCustody) {
		broken = append(broken, fmt.Sprintf("fiat custody sink balance mismatch: expected=%s actual=%s", expectedCustody, actualCustody))
	}
	custodyIter := storetypes.KVStorePrefixIterator(store, types.PrefixFiatCustodyEffect)
	for ; custodyIter.Valid(); custodyIter.Next() {
		key := custodyIter.Key()[len(types.PrefixFiatCustodyEffect):]
		if len(key) < 2 || key[len(key)-1] != 0 || len(custodyIter.Value()) != sha256.Size {
			broken = append(broken, fmt.Sprintf("malformed fiat custody effect %x", custodyIter.Key()))
			continue
		}
		record, found := k.GetFiatConversion(ctx, string(key[:len(key)-1]))
		if !found || !record.ValueMovementApplied || record.LegacyQuarantined || !bytes.Equal(record.CustodySinkEffectHash, custodyIter.Value()) {
			broken = append(broken, fmt.Sprintf("orphan fiat custody effect %x", custodyIter.Key()))
		}
	}
	custodyIter.Close()
	return broken
}

// validateCompletedFiatTreasuryAccounting is intentionally payout-scoped, not
// a fabricated global module liability equation. The settlement store has no
// complete historical liability model for unrelated escrow/reward balances.
// For every authenticated completed fiat conversion it can, however, prove per
// denomination that the custody effect plus retained components equals gross
// exposure and that each retained component has exactly one treasury entry.
func (k Keeper) validateCompletedFiatTreasuryAccounting(ctx sdk.Context, payouts map[string]types.PayoutRecord) []string {
	type retainedRecords map[types.TreasuryRecordType][]types.TreasuryRecord
	recordsByPayout := make(map[string]retainedRecords)
	completedByPayout := make(map[string]types.FiatConversionRecord)
	var broken []string
	k.WithFiatConversions(ctx, func(conversion types.FiatConversionRecord) bool {
		if !conversion.LegacyQuarantined && conversion.State == types.FiatConversionStatePayoutCompleted && conversion.ValueMovementApplied {
			completedByPayout[conversion.PayoutID] = conversion
		}
		return false
	})
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.skey), types.PrefixTreasuryRecord)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var record types.TreasuryRecord
		if err := json.Unmarshal(iterator.Value(), &record); err != nil {
			// Historical treasury data is outside this scoped liability proof.
			// A malformed expected deterministic key is detected as missing below.
			continue
		}
		if _, scoped := completedByPayout[record.PayoutID]; !scoped {
			continue
		}
		if record.RecordID == "" || !bytes.Equal(iterator.Key(), types.TreasuryRecordKey(record.RecordID)) || !record.Amount.IsValid() || record.Amount.IsZero() {
			broken = append(broken, fmt.Sprintf("malformed scoped treasury record %x", iterator.Key()))
			continue
		}
		payout, found := payouts[record.PayoutID]
		if !found || payout.SettlementID != record.SettlementID {
			broken = append(broken, fmt.Sprintf("orphan scoped treasury record %x", iterator.Key()))
			continue
		}
		switch record.RecordType {
		case types.TreasuryRecordPlatformFee, types.TreasuryRecordValidatorFee, types.TreasuryRecordHoldback:
			if record.RecordID != treasuryPayoutRecordID(record.PayoutID, record.RecordType) {
				broken = append(broken, fmt.Sprintf("noncanonical scoped treasury record %x", iterator.Key()))
				continue
			}
			if recordsByPayout[record.PayoutID] == nil {
				recordsByPayout[record.PayoutID] = make(retainedRecords)
			}
			recordsByPayout[record.PayoutID][record.RecordType] = append(recordsByPayout[record.PayoutID][record.RecordType], record)
		case types.TreasuryRecordRefund, types.TreasuryRecordWithdrawal:
		default:
			broken = append(broken, fmt.Sprintf("unknown treasury record type %x", iterator.Key()))
		}
	}

	completedPayoutIDs := make([]string, 0, len(completedByPayout))
	for payoutID := range completedByPayout {
		completedPayoutIDs = append(completedPayoutIDs, payoutID)
	}
	sort.Strings(completedPayoutIDs)
	for _, payoutID := range completedPayoutIDs {
		conversion := completedByPayout[payoutID]
		payout, found := payouts[conversion.PayoutID]
		if !found {
			continue // reported by the primary payout linkage invariant
		}
		retained := payout.PlatformFee.Add(payout.ValidatorFee...).Add(payout.HoldbackAmount...)
		reconciled := sdk.NewCoins(conversion.CustodySinkAmount).Add(retained...)
		if !reconciled.Equal(payout.GrossAmount) {
			broken = append(broken, conversion.ConversionID+": custody plus retained components do not equal gross exposure")
		}
		for _, expected := range []struct {
			recordType types.TreasuryRecordType
			amount     sdk.Coins
		}{
			{types.TreasuryRecordPlatformFee, payout.PlatformFee},
			{types.TreasuryRecordValidatorFee, payout.ValidatorFee},
			{types.TreasuryRecordHoldback, payout.HoldbackAmount},
		} {
			records := recordsByPayout[payout.PayoutID][expected.recordType]
			if expected.amount.IsZero() {
				if len(records) != 0 {
					broken = append(broken, conversion.ConversionID+": zero retained component has treasury entry")
				}
				continue
			}
			if len(records) != 1 || !records[0].Amount.Equal(expected.amount) {
				broken = append(broken, conversion.ConversionID+": retained treasury component missing, duplicated, or mismatched")
			}
		}
	}
	return broken
}
