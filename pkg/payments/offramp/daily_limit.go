package offramp

import (
	"context"
	"sync"

	sdkmath "cosmossdk.io/math"
)

// DailyLimitRepository atomically reserves corridor capacity. Production
// implementations must persist both totals and operation IDs across restarts.
type DailyLimitRepository interface {
	ReserveDailyAmount(ctx context.Context, key string, operationID string, amount sdkmath.LegacyDec, limit sdkmath.LegacyDec) (bool, error)
	ReleaseDailyAmount(ctx context.Context, key string, operationID string) error
	Durable() bool
}

type dailyLimitReservation struct {
	amount sdkmath.LegacyDec
}

// MemoryDailyLimitRepository is suitable only for tests and engineering
// sandboxes. Production adapter construction rejects it.
type MemoryDailyLimitRepository struct {
	mu           sync.Mutex
	totals       map[string]sdkmath.LegacyDec
	reservations map[string]dailyLimitReservation
}

// NewMemoryDailyLimitRepository creates an empty sandbox repository.
func NewMemoryDailyLimitRepository() *MemoryDailyLimitRepository {
	return &MemoryDailyLimitRepository{
		totals:       make(map[string]sdkmath.LegacyDec),
		reservations: make(map[string]dailyLimitReservation),
	}
}

func (r *MemoryDailyLimitRepository) ReserveDailyAmount(_ context.Context, key string, operationID string, amount sdkmath.LegacyDec, limit sdkmath.LegacyDec) (bool, error) {
	if key == "" || operationID == "" || amount.IsNil() || !amount.IsPositive() || limit.IsNil() || !limit.IsPositive() {
		return false, ErrInvalidRequest
	}
	reservationKey := key + "|" + operationID
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.reservations[reservationKey]; ok {
		return existing.amount.Equal(amount), nil
	}
	total := r.totals[key]
	if total.IsNil() {
		total = sdkmath.LegacyZeroDec()
	}
	updated := total.Add(amount)
	if updated.GT(limit) {
		return false, nil
	}
	r.totals[key] = updated
	r.reservations[reservationKey] = dailyLimitReservation{amount: amount}
	return true, nil
}

func (r *MemoryDailyLimitRepository) ReleaseDailyAmount(_ context.Context, key string, operationID string) error {
	reservationKey := key + "|" + operationID
	r.mu.Lock()
	defer r.mu.Unlock()
	reservation, ok := r.reservations[reservationKey]
	if !ok {
		return nil
	}
	delete(r.reservations, reservationKey)
	total := r.totals[key]
	if total.IsNil() || total.LTE(reservation.amount) {
		delete(r.totals, key)
		return nil
	}
	r.totals[key] = total.Sub(reservation.amount)
	return nil
}

func (*MemoryDailyLimitRepository) Durable() bool { return false }

var _ DailyLimitRepository = (*MemoryDailyLimitRepository)(nil)
