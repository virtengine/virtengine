package waldur

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/virtengine/virtengine/pkg/usage"
)

func TestLineItemsFromUsageReport_RejectsInvalidAmounts(t *testing.T) {
	_, err := LineItemsFromUsageReport(&ResourceUsageReport{
		ResourceUUID: "resource-1",
		PeriodStart:  time.Now().Add(-time.Hour),
		PeriodEnd:    time.Now(),
		Components: []ComponentUsage{
			{Type: "cpu_hours", Amount: -1},
		},
	}, "uvirt")
	if err == nil {
		t.Fatal("LineItemsFromUsageReport() expected error for negative amount")
	}
}

func TestUsageReportFromLineItems_SortsComponents(t *testing.T) {
	items := []*usage.LineItem{
		{
			ResourceType: usage.ResourceStorage,
			Quantity:     sdkmath.LegacyNewDec(5),
		},
		{
			ResourceType: usage.ResourceCPU,
			Quantity:     sdkmath.LegacyNewDec(2),
		},
	}

	report, err := UsageReportFromLineItems(
		"resource-1",
		time.Now().Add(-time.Hour),
		time.Now(),
		items,
		"backend-1",
		map[string]string{"currency": "uvirt"},
	)
	if err != nil {
		t.Fatalf("UsageReportFromLineItems() unexpected error: %v", err)
	}
	if len(report.Components) != 2 {
		t.Fatalf("UsageReportFromLineItems() returned %d components, want 2", len(report.Components))
	}
	if report.Components[0].Type != "cpu_hours" || report.Components[1].Type != "storage_gb_hours" {
		t.Fatalf("UsageReportFromLineItems() component order = %+v", report.Components)
	}
}
