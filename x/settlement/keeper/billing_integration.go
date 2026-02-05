package keeper

import (
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
	"github.com/virtengine/virtengine/x/settlement/types"
)

// BillingKeeper defines the subset of escrow billing functionality used by settlement.
type BillingKeeper interface {
	SaveUsageRecord(ctx sdk.Context, record *billing.UsageRecord) error
	GetUsageRecord(ctx sdk.Context, recordID string) (*billing.UsageRecord, error)
	CreateInvoice(ctx sdk.Context, invoice *billing.Invoice, artifactCID string) (*billing.InvoiceLedgerRecord, error)
	UpdateInvoiceStatus(ctx sdk.Context, invoiceID string, newStatus billing.InvoiceStatus, initiator string) (*billing.InvoiceLedgerEntry, error)
	RecordPayment(ctx sdk.Context, invoiceID string, amount sdk.Coins, initiator string) (*billing.InvoiceLedgerEntry, error)
	GetInvoiceSequence(ctx sdk.Context) uint64
}

func (k Keeper) billingEnabled() bool {
	return k.billingKeeper != nil
}

func (k Keeper) recordBillingUsage(ctx sdk.Context, usage types.UsageRecord) error {
	if k.billingKeeper == nil {
		return nil
	}

	record := billing.UsageRecord{
		RecordID:     usage.UsageID,
		LeaseID:      usage.LeaseID,
		Provider:     usage.Provider,
		Customer:     usage.Customer,
		StartTime:    usage.PeriodStart,
		EndTime:      usage.PeriodEnd,
		ResourceType: mapSettlementUsageType(usage.UsageType),
		UsageAmount:  sdkmath.LegacyNewDec(usageUnitsToInt64(usage.UsageUnits)),
		UnitPrice:    usage.UnitPrice,
		TotalAmount:  usage.TotalCost,
		Status:       billing.UsageRecordStatusPending,
	}

	return k.billingKeeper.SaveUsageRecord(ctx, &record)
}

func (k Keeper) updateBillingUsageRecordsForSettlement(
	ctx sdk.Context,
	settlement *types.SettlementRecord,
	usageRecords []types.UsageRecord,
	invoiceID string,
) error {
	if k.billingKeeper == nil {
		return nil
	}

	now := ctx.BlockTime()
	for _, usage := range usageRecords {
		record := billing.UsageRecord{
			RecordID:     usage.UsageID,
			LeaseID:      usage.LeaseID,
			Provider:     usage.Provider,
			Customer:     usage.Customer,
			StartTime:    usage.PeriodStart,
			EndTime:      usage.PeriodEnd,
			ResourceType: mapSettlementUsageType(usage.UsageType),
			UsageAmount:  sdkmath.LegacyNewDec(usageUnitsToInt64(usage.UsageUnits)),
			UnitPrice:    usage.UnitPrice,
			TotalAmount:  usage.TotalCost,
			InvoiceID:    invoiceID,
			Status:       billing.UsageRecordStatusSettled,
			BlockHeight:  settlement.BlockHeight,
			CreatedAt:    settlement.SettledAt,
			UpdatedAt:    now,
		}

		if err := k.billingKeeper.SaveUsageRecord(ctx, &record); err != nil {
			return err
		}
	}

	if len(usageRecords) == 0 {
		for _, coin := range settlement.TotalAmount {
			record := billing.UsageRecord{
				RecordID:     fmt.Sprintf("settlement-%s-%s", settlement.SettlementID, coin.Denom),
				LeaseID:      settlement.LeaseID,
				Provider:     settlement.Provider,
				Customer:     settlement.Customer,
				StartTime:    settlement.PeriodStart,
				EndTime:      settlement.PeriodEnd,
				ResourceType: billing.UsageTypeFixed,
				UsageAmount:  sdkmath.LegacyOneDec(),
				UnitPrice:    sdk.NewDecCoinFromCoin(coin),
				TotalAmount:  sdk.NewCoins(coin),
				InvoiceID:    invoiceID,
				Status:       billing.UsageRecordStatusSettled,
				BlockHeight:  settlement.BlockHeight,
				CreatedAt:    settlement.SettledAt,
				UpdatedAt:    now,
			}

			if err := k.billingKeeper.SaveUsageRecord(ctx, &record); err != nil {
				return err
			}
		}
	}

	return nil
}

func (k Keeper) generateInvoiceForSettlement(
	ctx sdk.Context,
	settlement *types.SettlementRecord,
	usageRecords []types.UsageRecord,
) (string, error) {
	if k.billingKeeper == nil {
		return "", nil
	}

	usageInputs := make([]billing.UsageInput, 0, len(usageRecords))
	for _, usage := range usageRecords {
		usageType := mapSettlementUsageType(usage.UsageType)
		usageInputs = append(usageInputs, billing.UsageInput{
			UsageRecordID: usage.UsageID,
			UsageType:     usageType,
			Quantity:      sdkmath.LegacyNewDec(usageUnitsToInt64(usage.UsageUnits)),
			Unit:          billing.UnitForUsageType(usageType),
			UnitPrice:     usage.UnitPrice,
			Description:   fmt.Sprintf("%s usage", usage.UsageType),
			PeriodStart:   usage.PeriodStart,
			PeriodEnd:     usage.PeriodEnd,
			Metadata: map[string]string{
				"order_id": usage.OrderID,
				"lease_id": usage.LeaseID,
			},
		})
	}

	if len(usageInputs) == 0 {
		for _, coin := range settlement.TotalAmount {
			usageInputs = append(usageInputs, billing.UsageInput{
				UsageRecordID: fmt.Sprintf("settlement-%s-%s", settlement.SettlementID, coin.Denom),
				UsageType:     billing.UsageTypeFixed,
				Quantity:      sdkmath.LegacyOneDec(),
				Unit:          billing.UnitForUsageType(billing.UsageTypeFixed),
				UnitPrice:     sdk.NewDecCoinFromCoin(coin),
				Description:   fmt.Sprintf("Final settlement for order %s", settlement.OrderID),
				PeriodStart:   settlement.PeriodStart,
				PeriodEnd:     settlement.PeriodEnd,
			})
		}
	}

	config := billing.DefaultInvoiceGeneratorConfig()
	config.RoundingMode = billing.RoundingModeDown

	generator := billing.NewInvoiceGenerator(config)
	now := ctx.BlockTime()

	seq := k.billingKeeper.GetInvoiceSequence(ctx)
	invoiceNumber := billing.NextInvoiceNumber(seq, config.InvoiceNumberPrefix)

	req := billing.InvoiceGenerationRequest{
		EscrowID:    settlement.EscrowID,
		OrderID:     settlement.OrderID,
		LeaseID:     settlement.LeaseID,
		Provider:    settlement.Provider,
		Customer:    settlement.Customer,
		UsageInputs: usageInputs,
		BillingPeriod: billing.BillingPeriod{
			StartTime:       settlement.PeriodStart,
			EndTime:         settlement.PeriodEnd,
			DurationSeconds: int64(settlement.PeriodEnd.Sub(settlement.PeriodStart).Seconds()),
			PeriodType:      billing.BillingPeriodTypeUsageBased,
		},
		Currency: settlementCurrency(settlement.TotalAmount, config.DefaultCurrency),
		Metadata: map[string]string{
			"settlement_id":   settlement.SettlementID,
			"settlement_type": string(settlement.SettlementType),
			"is_final":        fmt.Sprintf("%t", settlement.IsFinal),
		},
	}

	invoice, err := generator.GenerateInvoice(req, ctx.BlockHeight(), now)
	if err != nil {
		return "", err
	}

	reconcileInvoiceTotals(invoice, settlement.TotalAmount, "settlement adjustment")
	recalculateInvoiceTotals(invoice)

	invoice.InvoiceNumber = invoiceNumber
	invoice.SettlementID = settlement.SettlementID

	record, err := k.billingKeeper.CreateInvoice(ctx, invoice, fmt.Sprintf("invoice-%s", invoice.InvoiceID))
	if err != nil {
		return "", err
	}

	if _, err := k.billingKeeper.UpdateInvoiceStatus(ctx, record.InvoiceID, billing.InvoiceStatusPending, "settlement"); err != nil {
		return "", err
	}

	if _, err := k.billingKeeper.RecordPayment(ctx, record.InvoiceID, record.Total, "settlement"); err != nil {
		return "", err
	}

	return record.InvoiceID, nil
}

func settlementCurrency(total sdk.Coins, fallback string) string {
	if len(total) > 0 && total[0].Denom != "" {
		return total[0].Denom
	}
	return fallback
}

func mapSettlementUsageType(value string) billing.UsageType {
	switch strings.ToLower(value) {
	case "cpu", "cpu_core_hours", "core-hour", "core_hours":
		return billing.UsageTypeCPU
	case "memory", "mem", "memory_gb_hours", "gb-hour":
		return billing.UsageTypeMemory
	case "storage", "storage_gb_hours", "gb-month", "gb-hour-storage":
		return billing.UsageTypeStorage
	case "network", "bandwidth", "egress", "ingress":
		return billing.UsageTypeNetwork
	case "gpu", "gpu_hours", "gpu-hour":
		return billing.UsageTypeGPU
	case "fixed", "flat":
		return billing.UsageTypeFixed
	case "setup", "one_time":
		return billing.UsageTypeSetup
	default:
		return billing.UsageTypeOther
	}
}

const maxInt64 = int64(^uint64(0) >> 1)

func usageUnitsToInt64(units uint64) int64 {
	if units > uint64(maxInt64) {
		return maxInt64
	}
	return int64(units)
}

func reconcileInvoiceTotals(inv *billing.Invoice, target sdk.Coins, reason string) {
	targetDenoms := make(map[string]struct{}, len(target))
	for _, coin := range target {
		targetDenoms[coin.Denom] = struct{}{}
		current := inv.Total.AmountOf(coin.Denom)
		diff := coin.Amount.Sub(current)
		if diff.IsZero() {
			continue
		}

		if diff.IsPositive() {
			inv.LineItems = append(inv.LineItems, billing.LineItem{
				LineItemID:  fmt.Sprintf("adjustment-%s", coin.Denom),
				Description: reason,
				UsageType:   billing.UsageTypeFixed,
				Quantity:    sdkmath.LegacyOneDec(),
				Unit:        billing.UnitForUsageType(billing.UsageTypeFixed),
				UnitPrice:   sdk.NewDecCoinFromCoin(sdk.NewCoin(coin.Denom, diff)),
				Amount:      sdk.NewCoins(sdk.NewCoin(coin.Denom, diff)),
			})
			continue
		}

		discount := billing.AppliedDiscount{
			DiscountID:  fmt.Sprintf("disc-%s", coin.Denom),
			PolicyID:    "settlement-adjustment",
			Type:        billing.DiscountTypeFixed,
			Description: reason,
			Amount:      sdk.NewCoins(sdk.NewCoin(coin.Denom, diff.Neg())),
			AppliedAt:   time.Now().UTC(),
			AppliedBy:   "settlement",
		}
		inv.Discounts = append(inv.Discounts, discount)
	}

	for _, coin := range inv.Total {
		if _, ok := targetDenoms[coin.Denom]; ok {
			continue
		}

		if coin.Amount.IsZero() {
			continue
		}

		discount := billing.AppliedDiscount{
			DiscountID:  fmt.Sprintf("disc-extra-%s", coin.Denom),
			PolicyID:    "settlement-adjustment",
			Type:        billing.DiscountTypeFixed,
			Description: reason,
			Amount:      sdk.NewCoins(coin),
			AppliedAt:   time.Now().UTC(),
			AppliedBy:   "settlement",
		}
		inv.Discounts = append(inv.Discounts, discount)
	}
}

func recalculateInvoiceTotals(inv *billing.Invoice) {
	subtotal := sdk.NewCoins()
	for _, item := range inv.LineItems {
		subtotal = subtotal.Add(item.Amount...)
	}
	inv.Subtotal = subtotal

	discountTotal := sdk.NewCoins()
	for _, discount := range inv.Discounts {
		discountTotal = discountTotal.Add(discount.Amount...)
	}
	inv.DiscountTotal = discountTotal

	taxTotal := sdk.NewCoins()
	if inv.TaxDetails != nil {
		taxTotal = inv.TaxDetails.TotalTax
	}
	inv.TaxTotal = taxTotal

	total := subtotal.Sub(discountTotal...).Add(taxTotal...)
	inv.Total = total
	inv.AmountDue = total.Sub(inv.AmountPaid...)
}
