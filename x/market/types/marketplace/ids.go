// Package marketplace provides types for the marketplace on-chain module.
package marketplace

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseOrderID parses an OrderID from its string form ("customer/sequence").
func ParseOrderID(value string) (OrderID, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return OrderID{}, fmt.Errorf("invalid order id format: %s", value)
	}

	seq, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return OrderID{}, fmt.Errorf("invalid order sequence: %w", err)
	}

	id := OrderID{
		CustomerAddress: parts[0],
		Sequence:        seq,
	}
	if err := id.Validate(); err != nil {
		return OrderID{}, err
	}
	return id, nil
}

// ParseAllocationID parses an AllocationID from its string form ("customer/orderSeq/allocationSeq").
func ParseAllocationID(value string) (AllocationID, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 3 {
		return AllocationID{}, fmt.Errorf("invalid allocation id format: %s", value)
	}

	orderSeq, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return AllocationID{}, fmt.Errorf("invalid order sequence: %w", err)
	}

	allocSeq, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return AllocationID{}, fmt.Errorf("invalid allocation sequence: %w", err)
	}

	id := AllocationID{
		OrderID: OrderID{
			CustomerAddress: parts[0],
			Sequence:        orderSeq,
		},
		Sequence: allocSeq,
	}
	if err := id.Validate(); err != nil {
		return AllocationID{}, err
	}
	return id, nil
}
