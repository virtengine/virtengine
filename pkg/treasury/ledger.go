package treasury

import (
	"fmt"
	"sort"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

type LedgerEntryType string

const (
	LedgerEntryConversion LedgerEntryType = "conversion"
	LedgerEntryFee        LedgerEntryType = "fee"
	LedgerEntryAdjustment LedgerEntryType = "adjustment"
)

type LedgerEntry struct {
	ID           string
	Type         LedgerEntryType
	FromAsset    string
	ToAsset      string
	InputAmount  sdkmath.Int
	OutputAmount sdkmath.Int
	FeeAmount    sdkmath.Int
	FeeAsset     string
	Timestamp    time.Time
	Reference    string
}

type Ledger struct {
	mu       sync.Mutex
	balances map[string]sdkmath.Int
	entries  []LedgerEntry
}

func NewLedger() *Ledger {
	return &Ledger{balances: make(map[string]sdkmath.Int)}
}

func (l *Ledger) RecordConversion(entry LedgerEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("ledger-%d", time.Now().UnixNano())
	}
	entry.Type = LedgerEntryConversion
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	l.applyBalanceDelta(entry.FromAsset, entry.InputAmount.Neg())
	l.applyBalanceDelta(entry.ToAsset, entry.OutputAmount)

	if !entry.FeeAmount.IsZero() {
		switch entry.FeeAsset {
		case "", entry.ToAsset:
			l.applyBalanceDelta(entry.ToAsset, entry.FeeAmount.Neg())
		case entry.FromAsset:
			l.applyBalanceDelta(entry.FromAsset, entry.FeeAmount.Neg())
		default:
			l.applyBalanceDelta(entry.FeeAsset, entry.FeeAmount.Neg())
		}
	}

	l.entries = append(l.entries, entry)
}

func (l *Ledger) RecordFee(asset string, amount sdkmath.Int, reference string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LedgerEntry{
		ID:        fmt.Sprintf("fee-%d", time.Now().UnixNano()),
		Type:      LedgerEntryFee,
		FromAsset: asset,
		FeeAsset:  asset,
		FeeAmount: amount,
		Timestamp: time.Now().UTC(),
		Reference: reference,
	}

	l.applyBalanceDelta(asset, amount.Neg())
	l.entries = append(l.entries, entry)
}

func (l *Ledger) Balance(asset string) sdkmath.Int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.balances[asset]
}

func (l *Ledger) Balances() map[string]sdkmath.Int {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make(map[string]sdkmath.Int, len(l.balances))
	for k, v := range l.balances {
		out[k] = v
	}
	return out
}

func (l *Ledger) Entries() []LedgerEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	entries := make([]LedgerEntry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

func (l *Ledger) applyBalanceDelta(asset string, delta sdkmath.Int) {
	if asset == "" {
		return
	}
	current, ok := l.balances[asset]
	if !ok || current.IsNil() {
		current = sdkmath.ZeroInt()
	}
	l.balances[asset] = current.Add(delta)
}

type ReconciliationReport struct {
	DriftByAsset map[string]sdkmath.Int
	Balanced     bool
}

func (l *Ledger) Reconcile(actual map[string]sdkmath.Int) ReconciliationReport {
	l.mu.Lock()
	defer l.mu.Unlock()

	drift := make(map[string]sdkmath.Int)
	balanced := true

	assets := make(map[string]struct{})
	for asset := range l.balances {
		assets[asset] = struct{}{}
	}
	for asset := range actual {
		assets[asset] = struct{}{}
	}

	keys := make([]string, 0, len(assets))
	for asset := range assets {
		keys = append(keys, asset)
	}
	sort.Strings(keys)

	for _, asset := range keys {
		expected := l.balances[asset]
		act := actual[asset]
		delta := act.Sub(expected)
		drift[asset] = delta
		if !delta.IsZero() {
			balanced = false
		}
	}

	return ReconciliationReport{DriftByAsset: drift, Balanced: balanced}
}
