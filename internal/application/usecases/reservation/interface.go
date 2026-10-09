package reservation

import (
	"context"

	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
)

type ReservationUseCase interface {
	ReserveStock(ctx context.Context, req request.ReserveStockRequest) (*response.ReserveStockResponse, error)
	ConfirmReservation(ctx context.Context, req request.ConfirmReservationRequest) (*response.ConfirmReservationResponse, error)
	ReleaseExpiredReservations(ctx context.Context) (int, error)
}
