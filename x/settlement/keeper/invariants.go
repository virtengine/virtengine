package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/settlement/types"
)

// RegisterInvariants registers settlement invariants.
//
//nolint:staticcheck // sdk.InvariantRegistry is required by the module interface.
func RegisterInvariants(ir sdk.InvariantRegistry, k IKeeper) {
	ir.RegisterRoute(types.ModuleName, "escrow-settlement-reconciliation", EscrowSettlementReconciliationInvariant(k))
	ir.RegisterRoute(types.ModuleName, "authenticated-usage-replay-indexes", AuthenticatedUsageReplayInvariant(k))
	ir.RegisterRoute(types.ModuleName, "canonical-financial-cases", FinancialCaseInvariant(k))
	ir.RegisterRoute(types.ModuleName, "authenticated-fiat-conversions", FiatConversionInvariant(k))
}

// FiatConversionInvariant detects malformed records, orphan/replay/index/daily
// corruption, profile drift, and terminal payout contradictions.
//
//nolint:staticcheck // sdk.Invariant is required by the module interface.
func FiatConversionInvariant(k IKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		broken := k.ValidateFiatConversionInvariants(ctx)
		if len(broken) > 0 {
			return fmt.Sprintf("authenticated fiat conversions broken: %s", strings.Join(broken, "; ")), true
		}
		return "authenticated fiat conversions: ok", false
	}
}

// FinancialCaseInvariant fails closed on malformed cases, orphan holds, broken
// indexes, unconserved allocations, or incomplete terminal effects.
//
//nolint:staticcheck // sdk.Invariant is required by the module interface.
func FinancialCaseInvariant(k IKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		broken := k.ValidateFinancialCaseInvariants(ctx)
		if len(broken) > 0 {
			return fmt.Sprintf("canonical financial cases broken: %s", strings.Join(broken, "; ")), true
		}
		return "canonical financial cases: ok", false
	}
}

// AuthenticatedUsageReplayInvariant ensures exact-once indexes remain bound to
// the authenticated record that can trigger financial side effects.
//
//nolint:staticcheck // sdk.Invariant is required by the module interface.
func AuthenticatedUsageReplayInvariant(k IKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		broken := k.ValidateUsageReplayIndexes(ctx)
		if len(broken) > 0 {
			return fmt.Sprintf("authenticated usage replay indexes broken: %s", strings.Join(broken, "; ")), true
		}
		return "authenticated usage replay indexes: ok", false
	}
}

// EscrowSettlementReconciliationInvariant ensures escrow debits equal settlement totals per order.
//
//nolint:staticcheck // sdk.Invariant is required by the module interface.
func EscrowSettlementReconciliationInvariant(k IKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var broken []string

		k.WithEscrows(ctx, func(escrow types.EscrowAccount) bool {
			settlements := k.GetSettlementsByOrder(ctx, escrow.OrderID)
			total := sdk.NewCoins()
			for _, settlement := range settlements {
				if settlement.EscrowID != escrow.EscrowID {
					continue
				}
				total = total.Add(settlement.TotalAmount...)
			}

			if !total.IsAllGTE(escrow.TotalSettled) || !escrow.TotalSettled.IsAllGTE(total) {
				broken = append(broken, fmt.Sprintf("order=%s escrow=%s settled=%s expected=%s", escrow.OrderID, escrow.EscrowID, escrow.TotalSettled.String(), total.String()))
			}

			return false
		})

		if len(broken) > 0 {
			return fmt.Sprintf("escrow settlement reconciliation broken: %s", strings.Join(broken, "; ")), true
		}

		return "escrow settlement reconciliation: ok", false
	}
}
