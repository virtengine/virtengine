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

// ParseOfferingID parses an OfferingID from its string form ("provider/sequence").
func ParseOfferingID(value string) (OfferingID, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return OfferingID{}, fmt.Errorf("invalid offering id format: %s", value)
	}

	seq, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return OfferingID{}, fmt.Errorf("invalid offering sequence: %w", err)
	}

	id := OfferingID{
		ProviderAddress: parts[0],
		Sequence:        seq,
	}
	if err := id.Validate(); err != nil {
		return OfferingID{}, err
	}
	return id, nil
}

// ParseBidID parses a BidID from its string form ("customer/orderSeq/provider/bidSeq").
func ParseBidID(value string) (BidID, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 4 {
		return BidID{}, fmt.Errorf("invalid bid id format: %s", value)
	}

	orderSeq, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return BidID{}, fmt.Errorf("invalid order sequence: %w", err)
	}

	bidSeq, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return BidID{}, fmt.Errorf("invalid bid sequence: %w", err)
	}

	id := BidID{
		OrderID: OrderID{
			CustomerAddress: parts[0],
			Sequence:        orderSeq,
		},
		ProviderAddress: parts[2],
		Sequence:        bidSeq,
	}
	if err := id.Validate(); err != nil {
		return BidID{}, err
	}
	return id, nil
}
