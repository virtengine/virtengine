package billing

import (
	"fmt"

	"github.com/virtengine/virtengine/pkg/usage"
)

// UsageInputFromLineItem converts a canonical usage line item into a billing usage input.
func UsageInputFromLineItem(item *usage.LineItem) (UsageInput, error) {
	if item == nil {
		return UsageInput{}, fmt.Errorf("line item is nil")
	}

	usageType, err := UsageTypeFromResource(item.ResourceType)
	if err != nil {
		return UsageInput{}, err
	}

	input := UsageInput{
		UsageRecordID: item.UsageRecordID,
		UsageType:     usageType,
		Quantity:      item.Quantity,
		Unit:          item.Unit,
		UnitPrice:     item.UnitPrice,
		Description:   fmt.Sprintf("%s usage", item.ResourceType),
		PeriodStart:   item.PeriodStart,
		PeriodEnd:     item.PeriodEnd,
		Metadata:      item.Metadata,
	}

	return input, nil
}

// UsageInputsFromLineItems converts a slice of canonical line items into usage inputs.
func UsageInputsFromLineItems(items []*usage.LineItem) ([]UsageInput, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no line items provided")
	}

	inputs := make([]UsageInput, 0, len(items))
	for _, item := range items {
		input, err := UsageInputFromLineItem(item)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}

// UsageTypeFromResource maps canonical usage resource types into billing usage types.
func UsageTypeFromResource(resource usage.ResourceType) (UsageType, error) {
	switch resource {
	case usage.ResourceCPU:
		return UsageTypeCPU, nil
	case usage.ResourceMemory:
		return UsageTypeMemory, nil
	case usage.ResourceStorage:
		return UsageTypeStorage, nil
	case usage.ResourceNetwork:
		return UsageTypeNetwork, nil
	case usage.ResourceGPU:
		return UsageTypeGPU, nil
	case usage.ResourceOther:
		return UsageTypeOther, nil
	default:
		return UsageTypeOther, fmt.Errorf("unknown resource type: %s", resource)
	}
}
