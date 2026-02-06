package waldur

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/usage"
)

// LineItemsFromUsageReport converts a Waldur usage report into canonical line items.
func LineItemsFromUsageReport(report *ResourceUsageReport, currency string) ([]*usage.LineItem, error) {
	if report == nil {
		return nil, fmt.Errorf("usage report is required")
	}

	if currency == "" {
		if report.Metadata != nil {
			currency = report.Metadata["currency"]
		}
	}
	if currency == "" {
		currency = "uvirt"
	}

	metadata := report.Metadata

	items := make([]*usage.LineItem, 0, len(report.Components))
	for _, component := range report.Components {
		resourceType := resourceTypeFromComponent(component.Type)
		unit := unitForResourceType(resourceType)
		quantity := sdkmath.LegacyNewDecWithPrec(int64(component.Amount*1000000), 6)
		unitPrice := sdk.NewDecCoinFromDec(currency, sdkmath.LegacyZeroDec())

		line := &usage.LineItem{
			Source:        usage.SourceWaldur,
			OrderID:       report.Metadata["order_id"],
			LeaseID:       report.Metadata["lease_id"],
			UsageRecordID: buildWaldurUsageRecordID(report.ResourceUUID, component.Type, report.PeriodStart),
			ResourceType:  resourceType,
			Quantity:      quantity,
			Unit:          unit,
			UnitPrice:     unitPrice,
			TotalCost:     sdk.NewCoin(currency, sdkmath.NewInt(0)),
			PeriodStart:   report.PeriodStart,
			PeriodEnd:     report.PeriodEnd,
			CreatedAt:     report.SubmittedAt,
			Metadata:      metadata,
		}
		line.LineItemID = line.CanonicalID("waldur")
		items = append(items, line)
	}

	return usage.NormalizeLineItems(items), nil
}

// UsageReportFromLineItems creates a Waldur usage report from canonical line items.
func UsageReportFromLineItems(
	resourceUUID string,
	periodStart time.Time,
	periodEnd time.Time,
	lineItems []*usage.LineItem,
	backendID string,
	metadata map[string]string,
) (*ResourceUsageReport, error) {
	if resourceUUID == "" {
		return nil, fmt.Errorf("resource UUID is required")
	}

	if len(lineItems) == 0 {
		return nil, fmt.Errorf("at least one line item is required")
	}

	components := make(map[string]float64)
	for _, item := range lineItems {
		componentType := componentTypeFromResource(item.ResourceType)
		amount, err := strconv.ParseFloat(item.Quantity.String(), 64)
		if err != nil {
			return nil, fmt.Errorf("parse quantity: %w", err)
		}
		components[componentType] += amount
	}

	reportComponents := make([]ComponentUsage, 0, len(components))
	for componentType, amount := range components {
		reportComponents = append(reportComponents, ComponentUsage{
			Type:   componentType,
			Amount: amount,
		})
	}

	report := &ResourceUsageReport{
		ResourceUUID: resourceUUID,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		Components:   reportComponents,
		BackendID:    backendID,
		Metadata:     metadata,
		SubmittedAt:  time.Now().UTC(),
	}

	return report, nil
}

func resourceTypeFromComponent(componentType string) usage.ResourceType {
	component := strings.ToLower(componentType)
	switch {
	case strings.Contains(component, "cpu"):
		return usage.ResourceCPU
	case strings.Contains(component, "mem"), strings.Contains(component, "ram"):
		return usage.ResourceMemory
	case strings.Contains(component, "storage"):
		return usage.ResourceStorage
	case strings.Contains(component, "gpu"):
		return usage.ResourceGPU
	case strings.Contains(component, "network"), strings.Contains(component, "bandwidth"):
		return usage.ResourceNetwork
	default:
		return usage.ResourceOther
	}
}

func componentTypeFromResource(resource usage.ResourceType) string {
	switch resource {
	case usage.ResourceCPU:
		return "cpu_hours"
	case usage.ResourceMemory:
		return "ram_gb_hours"
	case usage.ResourceStorage:
		return "storage_gb_hours"
	case usage.ResourceGPU:
		return "gpu_hours"
	case usage.ResourceNetwork:
		return "network_gb"
	default:
		return "other"
	}
}

func unitForResourceType(resource usage.ResourceType) string {
	switch resource {
	case usage.ResourceCPU:
		return "cpu-hour"
	case usage.ResourceMemory:
		return "gb-hour"
	case usage.ResourceStorage:
		return "gb-month"
	case usage.ResourceGPU:
		return "gpu-hour"
	case usage.ResourceNetwork:
		return "gb"
	default:
		return "unit"
	}
}

func buildWaldurUsageRecordID(resourceUUID string, componentType string, start time.Time) string {
	return fmt.Sprintf("waldur-%s-%s-%s", resourceUUID, componentType, start.UTC().Format("20060102"))
}
