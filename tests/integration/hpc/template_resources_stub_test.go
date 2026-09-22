//go:build e2e.integration

package hpc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
)

// templateReservationStub is a minimal ResourcesKeeper for the workload-template
// harness.
//
// The keeper rejects job submission when no resources keeper is wired
// (x/hpc/keeper/keeper.go submitJob), and production always wires one
// (app/types/app.go:605). The template suite exercises that submit path, so it
// needs a keeper. Reserve/ActivateReservation echo the requested capacity back
// rather than modelling inventory, which is all this suite requires.
//
// It is deliberately NOT installed by setupIntegrationKeeper: the
// job-lifecycle suite posts jobs via SetJob and depends on the reservation
// requirement staying inactive (resourcesKeeper == nil), so wiring it globally
// would change that test's behavior.
type templateReservationStub struct {
	reservations map[string]resourcesv1.Reservation
	reserved     int
	released     int
}

func newTemplateReservationStub() *templateReservationStub {
	return &templateReservationStub{reservations: make(map[string]resourcesv1.Reservation)}
}

// IsCanonicalReservationsActive reports canonical reservations as active so the
// keeper enforces (and this stub satisfies) the reservation requirement.
func (*templateReservationStub) IsCanonicalReservationsActive(sdk.Context) bool { return true }

func (s *templateReservationStub) Reserve(_ sdk.Context, request resourcesv1.ReservationRequest) (*resourcesv1.Reservation, error) {
	s.reserved++
	id := "reservation-" + request.ConsumerId
	reservation := resourcesv1.Reservation{
		ReservationId:    id,
		IdempotencyKey:   request.IdempotencyKey,
		RequestId:        request.RequestId,
		RequesterAddress: request.RequesterAddress,
		ProviderAddress:  request.ProviderAddress,
		ResourceClass:    request.ResourceClass,
		Capacity:         request.Capacity,
		State:            resourcesv1.ReservationState_RESERVATION_STATE_ACTIVE,
		ConsumerType:     request.ConsumerType,
		ConsumerId:       request.ConsumerId,
		MarketOrderId:    request.MarketOrderId,
		MarketBidId:      request.MarketBidId,
		MarketLeaseId:    request.MarketLeaseId,
		HpcJobId:         request.HpcJobId,
		EscrowId:         request.EscrowId,
		Version:          request.Version,
	}
	s.reservations[id] = reservation
	out := reservation
	return &out, nil
}

func (s *templateReservationStub) ActivateReservation(_ sdk.Context, reservationID string, _ resourcesv1.ReservationLink) (*resourcesv1.Reservation, error) {
	return s.mutate(reservationID, resourcesv1.ReservationState_RESERVATION_STATE_ACTIVE, nil)
}

func (s *templateReservationStub) ReleaseReservation(_ sdk.Context, reservationID, _ string) (*resourcesv1.Reservation, error) {
	s.released++
	return s.mutate(reservationID, resourcesv1.ReservationState_RESERVATION_STATE_RELEASED, nil)
}

func (s *templateReservationStub) QuarantineReservation(_ sdk.Context, reservationID, _ string) (*resourcesv1.Reservation, error) {
	return s.mutate(reservationID, resourcesv1.ReservationState_RESERVATION_STATE_QUARANTINED, nil)
}

func (s *templateReservationStub) mutate(reservationID string, state resourcesv1.ReservationState, apply func(*resourcesv1.Reservation)) (*resourcesv1.Reservation, error) {
	reservation, ok := s.reservations[reservationID]
	if !ok {
		return nil, fmt.Errorf("reservation %q not found", reservationID)
	}
	reservation.State = state
	if apply != nil {
		apply(&reservation)
	}
	s.reservations[reservationID] = reservation
	out := reservation
	return &out, nil
}

func (s *templateReservationStub) GetReservation(_ sdk.Context, reservationID string) (resourcesv1.Reservation, bool) {
	reservation, ok := s.reservations[reservationID]
	return reservation, ok
}
