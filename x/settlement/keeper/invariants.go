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
