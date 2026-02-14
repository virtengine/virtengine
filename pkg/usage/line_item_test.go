package usage

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand" //nolint:gosec // G404: test-only deterministic randomness for reproducibility
	"sort"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func baseLineItem() *LineItem {
	return &LineItem{
		LineItemID:    "li-1",
		Source:        SourceHPC,
		OrderID:       "order-1",
		LeaseID:       "lease-1",
		UsageRecordID: "usage-1",
		ResourceType:  ResourceCPU,
		Quantity:      sdkmath.LegacyNewDec(1500),
		Unit:          "milli-cpu",
		UnitPrice:     sdk.NewDecCoinFromDec("uvt", sdkmath.LegacyNewDec(2)),
		TotalCost:     sdk.NewInt64Coin("uvt", 3000),
		PeriodStart:   time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		PeriodEnd:     time.Date(2024, 1, 2, 4, 4, 5, 0, time.UTC),
		CreatedAt:     time.Date(2024, 1, 2, 5, 4, 5, 0, time.UTC),
		Metadata: map[string]string{
			"node":   "alpha",
			"region": "us-east-1",
		},
	}
}

func cloneLineItem(item *LineItem) *LineItem {
	if item == nil {
		return nil
	}
	metadata := make(map[string]string, len(item.Metadata))
	for key, value := range item.Metadata {
		metadata[key] = value
	}
	return &LineItem{
		LineItemID:    item.LineItemID,
		Source:        item.Source,
		OrderID:       item.OrderID,
		LeaseID:       item.LeaseID,
		UsageRecordID: item.UsageRecordID,
		ResourceType:  item.ResourceType,
		Quantity:      item.Quantity,
		Unit:          item.Unit,
		UnitPrice:     item.UnitPrice,
		TotalCost:     item.TotalCost,
		PeriodStart:   item.PeriodStart,
		PeriodEnd:     item.PeriodEnd,
		CreatedAt:     item.CreatedAt,
		Metadata:      metadata,
	}
}

func TestLineItemHashDeterminism(t *testing.T) {
	item := baseLineItem()
	first := item.Hash()
	second := item.Hash()

	if len(first) == 0 {
		t.Fatalf("expected hash to be non-empty")
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("expected hash to be deterministic")
	}

	clone := cloneLineItem(item)
	clone.LineItemID = "li-2"
	if !bytes.Equal(first, clone.Hash()) {
		t.Fatalf("expected LineItemID to be excluded from hash")
	}

	clone.OrderID = "order-2"
	if bytes.Equal(first, clone.Hash()) {
		t.Fatalf("expected OrderID change to alter hash")
	}
}

func TestLineItemHashHex(t *testing.T) {
	item := baseLineItem()
	expected := hex.EncodeToString(item.Hash())
	if item.HashHex() != expected {
		t.Fatalf("expected hash hex to match encoded hash")
	}

	var nilItem *LineItem
	if nilItem.HashHex() != "" {
		t.Fatalf("expected nil line item hash hex to be empty")
	}
}

func TestLineItemCanonicalID(t *testing.T) {
	item := baseLineItem()
	id := item.CanonicalID("")
	if !strings.HasPrefix(id, "li-") {
		t.Fatalf("expected default prefix to be applied")
	}
	if len(id) != len("li-")+16 {
		t.Fatalf("expected canonical id length to be prefix + 16 chars")
	}

	custom := item.CanonicalID("usage")
	if !strings.HasPrefix(custom, "usage-") {
		t.Fatalf("expected custom prefix to be applied")
	}

	var nilItem *LineItem
	if nilItem.CanonicalID("x") != "" {
		t.Fatalf("expected empty canonical id for nil line item")
	}
}

func TestNormalizeLineItemsSorting(t *testing.T) {
	const (
		usageIDOne = "usage-1"
		usageIDTwo = "usage-2"
	)
	itemA := baseLineItem()
	itemA.UsageRecordID = usageIDTwo
	itemB := baseLineItem()
	itemB.UsageRecordID = usageIDOne
	itemC := baseLineItem()
	itemC.UsageRecordID = usageIDOne
	itemC.LineItemID = "li-3"

	input := []*LineItem{itemA, itemB, itemC}
	normalized := NormalizeLineItems(input)

	if normalized[0] != itemB || normalized[1] != itemC || normalized[2] != itemA {
		t.Fatalf("expected items to be sorted by canonical key")
	}
}

func TestNormalizeLineItemsEdgeCases(t *testing.T) {
	var nilItems []*LineItem
	if NormalizeLineItems(nilItems) != nil {
		t.Fatalf("expected nil slice to remain nil")
	}

	item := baseLineItem()
	duplicate := cloneLineItem(item)
	duplicate.LineItemID = "li-dup"
	withNil := []*LineItem{item, nil, duplicate}

	normalized := NormalizeLineItems(withNil)
	if len(normalized) != len(withNil) {
		t.Fatalf("expected duplicates to be preserved")
	}
	if normalized[0] != nil {
		t.Fatalf("expected nil line items to sort first")
	}
	if normalized[1] != item || normalized[2] != duplicate {
		t.Fatalf("expected stable ordering for duplicates")
	}
}

func TestLineItemHashEdgeCases(t *testing.T) {
	cases := []struct {
		name string
		item *LineItem
	}{
		{
			name: "zero-values",
			item: &LineItem{
				OrderID:       "order-zero",
				LeaseID:       "lease-zero",
				UsageRecordID: "usage-zero",
				ResourceType:  ResourceMemory,
				Quantity:      sdkmath.LegacyZeroDec(),
				Unit:          "gb",
				UnitPrice:     sdk.NewDecCoinFromDec("uvt", sdkmath.LegacyZeroDec()),
				TotalCost:     sdk.NewInt64Coin("uvt", 0),
				PeriodStart:   time.Unix(0, 0).UTC(),
				PeriodEnd:     time.Unix(0, 0).UTC(),
				Metadata:      map[string]string{},
			},
		},
		{
			name: "max-values",
			item: &LineItem{
				OrderID:       "order-max",
				LeaseID:       "lease-max",
				UsageRecordID: "usage-max",
				ResourceType:  ResourceStorage,
				Quantity:      sdkmath.LegacyNewDec(1_000_000_000),
				Unit:          "tb",
				UnitPrice:     sdk.NewDecCoinFromDec("uvt", sdkmath.LegacyNewDec(9_999_999)),
				TotalCost:     sdk.NewInt64Coin("uvt", 9_999_999_999),
				PeriodStart:   time.Unix(1_000_000, 0).UTC(),
				PeriodEnd:     time.Unix(2_000_000, 0).UTC(),
				Metadata: map[string]string{
					"region": "us-west-2",
				},
			},
		},
		{
			name: "unicode-values",
			item: &LineItem{
				OrderID:       "order-unicode",
				LeaseID:       "lease-unicode",
				UsageRecordID: "usage-unicode",
				ResourceType:  ResourceOther,
				Quantity:      sdkmath.LegacyNewDec(42),
				Unit:          "核/秒",
				UnitPrice:     sdk.NewDecCoinFromDec("uvt", sdkmath.LegacyNewDec(7)),
				TotalCost:     sdk.NewInt64Coin("uvt", 294),
				PeriodStart:   time.Unix(3_000_000, 0).UTC(),
				PeriodEnd:     time.Unix(4_000_000, 0).UTC(),
				Metadata: map[string]string{
					"描述": "高性能",
				},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			hash := testCase.item.Hash()
			if len(hash) == 0 {
				t.Fatalf("expected hash to be computed for %s", testCase.name)
			}
			if !bytes.Equal(hash, testCase.item.Hash()) {
				t.Fatalf("expected hash to be deterministic for %s", testCase.name)
			}
		})
	}
}

func TestLineItemHashFieldSensitivity(t *testing.T) {
	base := baseLineItem()
	baseHash := base.Hash()

	tests := []struct {
		name  string
		apply func(item *LineItem)
		same  bool
	}{
		{
			name: "order-id",
			apply: func(item *LineItem) {
				item.OrderID = "order-2"
			},
		},
		{
			name: "lease-id",
			apply: func(item *LineItem) {
				item.LeaseID = "lease-2"
			},
		},
		{
			name: "usage-record-id",
			apply: func(item *LineItem) {
				item.UsageRecordID = "usage-2"
			},
		},
		{
			name: "resource-type",
			apply: func(item *LineItem) {
				item.ResourceType = ResourceGPU
			},
		},
		{
			name: "quantity",
			apply: func(item *LineItem) {
				item.Quantity = sdkmath.LegacyNewDec(999)
			},
		},
		{
			name: "unit",
			apply: func(item *LineItem) {
				item.Unit = "gpu-hour"
			},
		},
		{
			name: "unit-price",
			apply: func(item *LineItem) {
				item.UnitPrice = sdk.NewDecCoinFromDec("uvt", sdkmath.LegacyNewDec(9))
			},
		},
		{
			name: "period-start",
			apply: func(item *LineItem) {
				item.PeriodStart = item.PeriodStart.Add(time.Hour)
			},
		},
		{
			name: "period-end",
			apply: func(item *LineItem) {
				item.PeriodEnd = item.PeriodEnd.Add(2 * time.Hour)
			},
		},
		{
			name: "metadata",
			apply: func(item *LineItem) {
				item.Metadata = map[string]string{
					"node":   "alpha",
					"region": "eu-central-1",
				}
			},
		},
		{
			name: "line-item-id-excluded",
			apply: func(item *LineItem) {
				item.LineItemID = "li-2"
			},
			same: true,
		},
		{
			name: "source-excluded",
			apply: func(item *LineItem) {
				item.Source = SourceWaldur
			},
			same: true,
		},
		{
			name: "total-cost-excluded",
			apply: func(item *LineItem) {
				item.TotalCost = sdk.NewInt64Coin("uvt", 999)
			},
			same: true,
		},
		{
			name: "created-at-excluded",
			apply: func(item *LineItem) {
				item.CreatedAt = item.CreatedAt.Add(24 * time.Hour)
			},
			same: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			item := cloneLineItem(base)
			testCase.apply(item)
			if testCase.same {
				if !bytes.Equal(baseHash, item.Hash()) {
					t.Fatalf("expected hash to remain stable for %s", testCase.name)
				}
				return
			}
			if bytes.Equal(baseHash, item.Hash()) {
				t.Fatalf("expected hash to change for %s", testCase.name)
			}
		})
	}
}

func TestLineItemHashMetadataOrderIndependence(t *testing.T) {
	itemA := baseLineItem()
	itemA.Metadata = map[string]string{
		"region": "us-east-1",
		"node":   "alpha",
	}
	itemB := baseLineItem()
	itemB.Metadata = map[string]string{
		"node":   "alpha",
		"region": "us-east-1",
	}

	if !bytes.Equal(itemA.Hash(), itemB.Hash()) {
		t.Fatalf("expected metadata order not to affect hash")
	}
}

func TestHashDeterminismProperty(t *testing.T) {
	rng := newDeterministicRand(46)
	for i := 0; i < 50; i++ {
		item := randomLineItem(rng, i)
		first := item.Hash()
		second := item.Hash()
		if !bytes.Equal(first, second) {
			t.Fatalf("expected deterministic hash at index %d", i)
		}
		clone := cloneLineItem(item)
		if !bytes.Equal(first, clone.Hash()) {
			t.Fatalf("expected cloned item hash to match at index %d", i)
		}
	}
}

func TestNormalizeLineItemsDeterminismProperty(t *testing.T) {
	rng := newDeterministicRand(99)
	items := make([]*LineItem, 0, 40)
	for i := 0; i < 40; i++ {
		items = append(items, randomLineItem(rng, i))
	}

	first := NormalizeLineItems(items)
	second := NormalizeLineItems(items)

	if len(first) != len(second) {
		t.Fatalf("expected normalized slices to have same length")
	}
	for i := range first {
		if canonicalKey(first[i]) != canonicalKey(second[i]) {
			t.Fatalf("expected normalized order to be deterministic at index %d", i)
		}
	}

	hashes := make(map[string]struct{}, len(first))
	for _, item := range first {
		hashes[item.HashHex()] = struct{}{}
	}
	if len(hashes) != len(first) {
		t.Fatalf("expected no hash collisions for generated items")
	}
}

func randomLineItem(rng *rand.Rand, index int) *LineItem {
	resourceTypes := []ResourceType{
		ResourceCPU,
		ResourceMemory,
		ResourceStorage,
		ResourceGPU,
		ResourceNetwork,
		ResourceOther,
	}
	unitChoices := []string{"cpu", "gb", "tb", "gpu-hour", "mbps"}

	metadata := map[string]string{
		"region": fmt.Sprintf("region-%d", rng.Intn(5)),
	}
	if rng.Intn(2) == 0 {
		metadata["node"] = fmt.Sprintf("node-%d", rng.Intn(10))
	}

	start := time.Unix(int64(1_700_000_000+rng.Intn(10_000)), 0).UTC()
	end := start.Add(time.Duration(60+rng.Intn(10_000)) * time.Second)

	return &LineItem{
		LineItemID:    fmt.Sprintf("li-%d", index),
		Source:        SourceHPC,
		OrderID:       fmt.Sprintf("order-%d", rng.Intn(1000)),
		LeaseID:       fmt.Sprintf("lease-%d", rng.Intn(1000)),
		UsageRecordID: fmt.Sprintf("usage-%d", rng.Intn(1000)),
		ResourceType:  resourceTypes[rng.Intn(len(resourceTypes))],
		Quantity:      sdkmath.LegacyNewDec(int64(1 + rng.Intn(5000))),
		Unit:          unitChoices[rng.Intn(len(unitChoices))],
		UnitPrice:     sdk.NewDecCoinFromDec("uvt", sdkmath.LegacyNewDec(int64(1+rng.Intn(20)))),
		TotalCost:     sdk.NewInt64Coin("uvt", int64(1+rng.Intn(10_000))),
		PeriodStart:   start,
		PeriodEnd:     end,
		CreatedAt:     start.Add(time.Duration(rng.Intn(300)) * time.Second),
		Metadata:      metadata,
	}
}

func newDeterministicRand(seed int64) *rand.Rand {
	//nolint:gosec // Deterministic math/rand is required for reproducible property tests.
	return rand.New(rand.NewSource(seed))
}

func TestCanonicalKeySorting(t *testing.T) {
	itemA := baseLineItem()
	itemA.UsageRecordID = "usage-2"
	itemB := baseLineItem()
	itemB.UsageRecordID = "usage-1"
	itemC := baseLineItem()
	itemC.UsageRecordID = "usage-3"

	items := []*LineItem{itemA, itemB, itemC}
	sort.SliceStable(items, func(i, j int) bool {
		return canonicalKey(items[i]) < canonicalKey(items[j])
	})

	if items[0] != itemB || items[1] != itemA || items[2] != itemC {
		t.Fatalf("expected canonical key ordering to sort by usage record id")
	}
}
