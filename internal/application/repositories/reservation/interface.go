package reservation

import (
	"context"
	"time"

	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities/reservation"
)

type ReservationRepository interface {
	Create(ctx context.Context, res *reservation.Reservation) error
	GetByID(ctx context.Context, reservationID string) (*reservation.Reservation, error)
	GetByIDForUpdate(ctx context.Context, reservationID string) (*reservation.Reservation, error)
	Update(ctx context.Context, res *reservation.Reservation) error
	FindExpiredPending(ctx context.Context, now time.Time, limit int) ([]*reservation.Reservation, error)
	// SumActiveByItem returns the total quantity of active (PENDING, unexpired
	// as of now) reservations for the given item.
	SumActiveByItem(ctx context.Context, itemID string, now time.Time) (int, error)
}
