//go:build e2e.integration

package waldur

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/usage"
	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
)

func TestWaldurSettlementUsageAndInvoiceFlow(t *testing.T) {
	h := newWaldurHarness(t, func(cfg *mocks.WaldurMockConfig) {
		cfg.EnableUsageReporting = true
	})

	offering := h.createOffering("Settlement Compute", h.providerAddr+"/settlement-compute", "VirtEngine.Compute", []waldur.PricingComponent{
		{Type: "usage", Name: "cpu_hours", MeasuredUnit: "hour", BillingType: "usage", Price: "0.10"},
		{Type: "usage", Name: "ram_gb_hours", MeasuredUnit: "gb-hour", BillingType: "usage", Price: "0.02"},
	})
	resourceUUID := h.registerResource(offering.UUID, "", "alloc-settlement-001", string(waldur.ResourceStateOK), nil)

	report, err := waldur.UsageReportFromLineItems(
		resourceUUID,
		time.Now().Add(-1*time.Hour).UTC(),
		time.Now().UTC(),
		[]*usage.LineItem{
			{
				ResourceType: usage.ResourceMemory,
				Quantity:     sdkmath.LegacyMustNewDecFromStr("8"),
				Unit:         "gb-hour",
			},
			{
				ResourceType: usage.ResourceCPU,
				Quantity:     sdkmath.LegacyMustNewDecFromStr("2"),
				Unit:         "cpu-hour",
			},
		},
		"alloc-settlement-001",
		map[string]string{"provider": h.providerAddr, "currency": sdk.DefaultBondDenom},
	)
	if err != nil {
		t.Fatalf("build usage report from line items: %v", err)
	}
	if len(report.Components) != 2 || report.Components[0].Type != "cpu_hours" || report.Components[1].Type != "ram_gb_hours" {
		t.Fatalf("unexpected normalized component order: %+v", report.Components)
	}

	ctx, cancel := h.context()
	reportResp, err := h.usage.SubmitUsageReport(ctx, report)
	cancel()
	if err != nil {
		t.Fatalf("submit usage report: %v", err)
	}
	if reportResp.State == "" {
		t.Fatalf("expected usage submission state")
	}

	if len(h.mock.GetUsageRecords(resourceUUID)) == 0 {
		t.Fatalf("expected recorded usage entries")
	}

	invoice, err := h.mock.CreateInvoice(
		context.Background(),
		resourceUUID,
		time.Now().Add(-24*time.Hour).UTC(),
		time.Now().UTC(),
		[]mocks.MockWaldurInvoiceLineItem{
			{Name: "cpu_hours", Quantity: "2.0", UnitPrice: "0.10", Total: "0.20", Unit: "hour"},
			{Name: "ram_gb_hours", Quantity: "8.0", UnitPrice: "0.02", Total: "0.16", Unit: "gb-hour"},
		},
		"0.36",
	)
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if invoice.TotalAmount != "0.36" {
		t.Fatalf("invoice total = %s, want 0.36", invoice.TotalAmount)
	}

	if err := h.mock.PayInvoice(context.Background(), invoice.UUID); err != nil {
		t.Fatalf("pay invoice: %v", err)
	}
	paid := h.mock.GetInvoice(invoice.UUID)
	if paid == nil || paid.State != "paid" {
		t.Fatalf("expected paid invoice state")
	}
}
