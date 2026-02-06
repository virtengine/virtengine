package usage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ResourceType defines the canonical usage resource type.
type ResourceType string

const (
	ResourceCPU     ResourceType = "cpu"
	ResourceMemory  ResourceType = "memory"
	ResourceStorage ResourceType = "storage"
	ResourceGPU     ResourceType = "gpu"
	ResourceNetwork ResourceType = "network"
	ResourceOther   ResourceType = "other"
)

// Source identifies the origin of the usage line item.
type Source string

const (
	SourceHPC        Source = "hpc"
	SourceWaldur     Source = "waldur"
	SourceSettlement Source = "settlement"
)

// LineItem represents the canonical usage line item shared across systems.
type LineItem struct {
	// LineItemID is the unique identifier for this line item.
	LineItemID string `json:"line_item_id"`

	// Source is the origin system for this line item.
	Source Source `json:"source"`

	// OrderID links to marketplace order.
	OrderID string `json:"order_id"`

	// LeaseID links to marketplace lease.
	LeaseID string `json:"lease_id"`

	// UsageRecordID links to the raw usage record.
	UsageRecordID string `json:"usage_record_id"`

	// ResourceType is the resource type (cpu, memory, storage, gpu, network).
	ResourceType ResourceType `json:"resource_type"`

	// Quantity is the amount consumed.
	Quantity sdkmath.LegacyDec `json:"quantity"`

	// Unit is the measurement unit.
	Unit string `json:"unit"`

	// UnitPrice is the price per unit.
	UnitPrice sdk.DecCoin `json:"unit_price"`

	// TotalCost is the total cost for this line item.
	TotalCost sdk.Coin `json:"total_cost"`

	// PeriodStart is the start of the usage period.
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the usage period.
	PeriodEnd time.Time `json:"period_end"`

	// CreatedAt is when the line item was created.
	CreatedAt time.Time `json:"created_at"`

	// Metadata contains additional details.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Hash returns a deterministic hash of the line item for idempotency checks.
func (l *LineItem) Hash() []byte {
	if l == nil {
		return nil
	}

	canonical := struct {
		OrderID       string       `json:"order_id"`
		LeaseID       string       `json:"lease_id"`
		UsageRecordID string       `json:"usage_record_id"`
		ResourceType  ResourceType `json:"resource_type"`
		Quantity      string       `json:"quantity"`
		Unit          string       `json:"unit"`
		UnitPrice     string       `json:"unit_price"`
		PeriodStart   int64        `json:"period_start"`
		PeriodEnd     int64        `json:"period_end"`
		Metadata      []kvPair     `json:"metadata"`
	}{
		OrderID:       l.OrderID,
		LeaseID:       l.LeaseID,
		UsageRecordID: l.UsageRecordID,
		ResourceType:  l.ResourceType,
		Quantity:      l.Quantity.String(),
		Unit:          l.Unit,
		UnitPrice:     l.UnitPrice.String(),
		PeriodStart:   l.PeriodStart.Unix(),
		PeriodEnd:     l.PeriodEnd.Unix(),
		Metadata:      sortedMetadata(l.Metadata),
	}

	data, err := json.Marshal(canonical)
	if err != nil {
		return nil
	}

	hash := sha256.Sum256(data)
	return hash[:]
}

// HashHex returns the hex-encoded hash of the line item.
func (l *LineItem) HashHex() string {
	hash := l.Hash()
	if len(hash) == 0 {
		return ""
	}
	return hex.EncodeToString(hash)
}

// CanonicalID builds a deterministic line item ID from the hash.
func (l *LineItem) CanonicalID(prefix string) string {
	hashHex := l.HashHex()
	if hashHex == "" {
		return ""
	}
	if prefix == "" {
		prefix = "li"
	}
	return fmt.Sprintf("%s-%s", prefix, hashHex[:16])
}

// NormalizeLineItems sorts line items into deterministic order.
func NormalizeLineItems(items []*LineItem) []*LineItem {
	if len(items) == 0 {
		return items
	}

	sorted := make([]*LineItem, 0, len(items))
	sorted = append(sorted, items...)

	sort.SliceStable(sorted, func(i, j int) bool {
		left := canonicalKey(sorted[i])
		right := canonicalKey(sorted[j])
		return left < right
	})

	return sorted
}

type kvPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func sortedMetadata(metadata map[string]string) []kvPair {
	if len(metadata) == 0 {
		return nil
	}

	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]kvPair, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, kvPair{Key: key, Value: metadata[key]})
	}
	return pairs
}

func canonicalKey(item *LineItem) string {
	if item == nil {
		return ""
	}
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		item.UsageRecordID,
		item.ResourceType,
		item.Unit,
		item.UnitPrice.String(),
		item.Quantity.String(),
		item.PeriodStart.UTC().Format(time.RFC3339Nano),
		item.PeriodEnd.UTC().Format(time.RFC3339Nano),
	)
}
