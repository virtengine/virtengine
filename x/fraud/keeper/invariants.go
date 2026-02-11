package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
)

// RegisterInvariants registers fraud module invariants.
//
//nolint:staticcheck // sdk.InvariantRegistry is required by the module interface.
func RegisterInvariants(ir sdk.InvariantRegistry, k IKeeper) {
	ir.RegisterRoute(types.ModuleName, "report-state-consistency", ReportStateConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "blacklist-integrity", BlacklistIntegrityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "evidence-hash-verification", EvidenceHashVerificationInvariant(k))
	ir.RegisterRoute(types.ModuleName, "reporter-exists", ReporterExistsInvariant(k))
}

// ReportStateConsistencyInvariant checks report status with queue consistency.
//
//nolint:staticcheck // sdk.Invariant is required by the module interface.
func ReportStateConsistencyInvariant(k IKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var broken []string

		k.WithFraudReports(ctx, func(report types.FraudReport) bool {
			if !report.Status.IsValid() {
				broken = append(broken, fmt.Sprintf("report=%s invalid-status=%s", report.ID, report.Status.String()))
				return false
			}

			queueEntry, inQueue := k.GetModeratorQueueEntry(ctx, report.ID)
			if report.Status.IsPending() {
				if !inQueue {
					broken = append(broken, fmt.Sprintf("report=%s pending-without-queue", report.ID))
				} else if queueEntry.Category != report.Category {
					broken = append(broken, fmt.Sprintf("report=%s queue-category-mismatch=%s report-category=%s", report.ID, queueEntry.Category.String(), report.Category.String()))
				}
				if report.Status == types.FraudReportStatusReviewing && report.AssignedModerator == "" {
					broken = append(broken, fmt.Sprintf("report=%s reviewing-without-assigned-moderator", report.ID))
				}
			} else if inQueue {
				broken = append(broken, fmt.Sprintf("report=%s terminal-with-queue-entry", report.ID))
			}

			if report.Status.IsTerminal() && report.ResolvedAt == nil {
				broken = append(broken, fmt.Sprintf("report=%s terminal-without-resolved-at", report.ID))
			}
			if !report.Status.IsTerminal() && report.ResolvedAt != nil {
				broken = append(broken, fmt.Sprintf("report=%s nonterminal-with-resolved-at", report.ID))
			}

			return false
		})

		if len(broken) > 0 {
			return fmt.Sprintf("report-state-consistency invariant broken: %s", strings.Join(broken, "; ")), true
		}
		return "report-state-consistency: ok", false
	}
}

// BlacklistIntegrityInvariant validates reported-party index integrity.
//
//nolint:staticcheck // sdk.Invariant is required by the module interface.
func BlacklistIntegrityInvariant(k IKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.StoreKey())
		indexStore := prefix.NewStore(store, types.ReportedPartyIndexPrefix)
		iterator := storetypes.KVStorePrefixIterator(indexStore, nil)
		defer iterator.Close()

		var broken []string

		for ; iterator.Valid(); iterator.Next() {
			key := string(iterator.Key())
			parts := strings.SplitN(key, "/", 2)
			if len(parts) != 2 {
				broken = append(broken, fmt.Sprintf("reported-party-index-invalid-key=%s", key))
				continue
			}
			reportedParty := parts[0]
			reportID := parts[1]
			value := string(iterator.Value())
			if value != reportID {
				broken = append(broken, fmt.Sprintf("reported-party-index-value-mismatch key=%s value=%s", reportID, value))
				continue
			}
			report, found := k.GetFraudReport(ctx, reportID)
			if !found {
				broken = append(broken, fmt.Sprintf("reported-party-index-missing-report=%s", reportID))
				continue
			}
			if report.ReportedParty != reportedParty {
				broken = append(broken, fmt.Sprintf("reported-party-index-report-mismatch report=%s expected=%s got=%s", report.ID, reportedParty, report.ReportedParty))
			}
		}

		if len(broken) > 0 {
			return fmt.Sprintf("blacklist-integrity invariant broken: %s", strings.Join(broken, "; ")), true
		}
		return "blacklist-integrity: ok", false
	}
}

// EvidenceHashVerificationInvariant verifies report content and evidence hashes.
//
//nolint:staticcheck // sdk.Invariant is required by the module interface.
func EvidenceHashVerificationInvariant(k IKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var broken []string

		k.WithFraudReports(ctx, func(report types.FraudReport) bool {
			expectedContentHash := report.ComputeContentHash()
			if report.ContentHash == "" || report.ContentHash != expectedContentHash {
				broken = append(broken, fmt.Sprintf("report=%s content-hash-mismatch", report.ID))
			}

			for idx, evidence := range report.Evidence {
				if evidence.EvidenceHash == "" {
					broken = append(broken, fmt.Sprintf("report=%s evidence[%d]-missing-hash", report.ID, idx))
					continue
				}
				normalized, ok := normalizeEvidenceHash(evidence.EvidenceHash)
				if !ok {
					continue
				}
				computed := sha256.Sum256(evidence.Ciphertext)
				computedHex := hex.EncodeToString(computed[:])
				if computedHex != normalized {
					broken = append(broken, fmt.Sprintf("report=%s evidence[%d]-hash-mismatch", report.ID, idx))
				}
			}

			return false
		})

		if len(broken) > 0 {
			return fmt.Sprintf("evidence-hash-verification invariant broken: %s", strings.Join(broken, "; ")), true
		}
		return "evidence-hash-verification: ok", false
	}
}

// ReporterExistsInvariant ensures reporters are valid provider addresses.
//
//nolint:staticcheck // sdk.Invariant is required by the module interface.
func ReporterExistsInvariant(k IKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var broken []string

		k.WithFraudReports(ctx, func(report types.FraudReport) bool {
			addr, err := sdk.AccAddressFromBech32(report.Reporter)
			if err != nil {
				broken = append(broken, fmt.Sprintf("report=%s invalid-reporter=%s", report.ID, report.Reporter))
				return false
			}
			if !k.IsProvider(ctx, addr) {
				broken = append(broken, fmt.Sprintf("report=%s reporter-not-provider=%s", report.ID, report.Reporter))
			}
			return false
		})

		if len(broken) > 0 {
			return fmt.Sprintf("reporter-exists invariant broken: %s", strings.Join(broken, "; ")), true
		}
		return "reporter-exists: ok", false
	}
}

func normalizeEvidenceHash(hash string) (string, bool) {
	trimmed := strings.TrimSpace(strings.ToLower(hash))
	trimmed = strings.TrimPrefix(trimmed, "sha256:")
	if len(trimmed) != 64 {
		return "", false
	}
	if _, err := hex.DecodeString(trimmed); err != nil {
		return "", false
	}
	return trimmed, true
}
