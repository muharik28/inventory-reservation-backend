package impl

import (
	"context"
	"time"

	"github.com/muharik28/inventory-reservation-backend/configs"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
	invRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/inventory"
	resRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/reservation"
	"github.com/muharik28/inventory-reservation-backend/internal/application/usecases/reservation"
	domainInv "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
	domainRes "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/reservation"
	"github.com/muharik28/inventory-reservation-backend/internal/pkg/logger"
	"github.com/muharik28/inventory-reservation-backend/internal/pkg/utils"
)

// TxManager is the persistence port this use case needs: running a function
// atomically and reading the PostgreSQL clock. Defined here (consumer side)
// so the use case never depends on a concrete ORM type; *sql.BunTxManager
// satisfies it.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
	Now(ctx context.Context) (time.Time, error)
}

type ReservationUseCaseImpl struct {
	txMgr   TxManager
	invRepo invRepo.InventoryRepository
	resRepo resRepo.ReservationRepository
	cfg     *configs.Config
}

func NewReservationUseCase(
	txMgr TxManager,
	invRepo invRepo.InventoryRepository,
	resRepo resRepo.ReservationRepository,
	cfg *configs.Config,
) reservation.ReservationUseCase {
	return &ReservationUseCaseImpl{
		txMgr:   txMgr,
		invRepo: invRepo,
		resRepo: resRepo,
		cfg:     cfg,
	}
}

func (u *ReservationUseCaseImpl) ReserveStock(ctx context.Context, req request.ReserveStockRequest) (*response.ReserveStockResponse, error) {
	if req.Quantity <= 0 {
		return nil, domainInv.ErrInvalidQuantity
	}

	var res *domainRes.Reservation

	err := u.txMgr.WithinTx(ctx, func(ctx context.Context) error {
		// Use the PostgreSQL clock so TTL and expiry checks are consistent
		// across every service instance.
		now, err := u.txMgr.Now(ctx)
		if err != nil {
			return err
		}

		// Serialize concurrent reservations for the same item on the
		// inventory row lock to prevent overselling.
		item, err := u.invRepo.GetByIDForUpdate(ctx, req.ItemID)
		if err != nil {
			return err
		}

		// Reserved stock is calculated from active (PENDING, unexpired)
		// reservations, so expired reservations never reduce availability.
		activeReserved, err := u.resRepo.SumActiveByItem(ctx, req.ItemID, now)
		if err != nil {
			return err
		}

		if !item.CanReserve(req.Quantity, activeReserved) {
			return domainInv.ErrInsufficientStock
		}

		res = &domainRes.Reservation{
			ReservationID: utils.GenerateID("res"),
			UserID:        req.UserID,
			ItemID:        req.ItemID,
			Quantity:      req.Quantity,
			Status:        domainRes.StatusPending,
			ExpiresAt:     now.Add(u.cfg.ReservationTTL),
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		return u.resRepo.Create(ctx, res)
	})

	if err != nil {
		return nil, err
	}

	return &response.ReserveStockResponse{
		Status:        "success",
		ReservationID: res.ReservationID,
		ItemID:        res.ItemID,
		Quantity:      res.Quantity,
		ExpiresAt:     res.ExpiresAt,
	}, nil
}

func (u *ReservationUseCaseImpl) ConfirmReservation(ctx context.Context, req request.ConfirmReservationRequest) (*response.ConfirmReservationResponse, error) {
	var confirmedRes *domainRes.Reservation

	err := u.txMgr.WithinTx(ctx, func(ctx context.Context) error {
		now, err := u.txMgr.Now(ctx)
		if err != nil {
			return err
		}

		res, err := u.resRepo.GetByIDForUpdate(ctx, req.ReservationID)
		if err != nil {
			return err
		}

		if !res.IsPending() {
			return domainRes.ErrReservationAlreadyHandled
		}

		if res.IsExpired(now) {
			return domainRes.ErrReservationExpired
		}

		item, err := u.invRepo.GetByIDForUpdate(ctx, res.ItemID)
		if err != nil {
			return err
		}

		// Physical stock decreases only here; reserved quantity is derived
		// from reservation status, so flipping the status releases it.
		item.TotalStock -= res.Quantity
		if item.TotalStock < 0 {
			return domainInv.ErrInsufficientStock
		}

		if err := u.invRepo.UpdateStock(ctx, item); err != nil {
			return err
		}

		res.Status = domainRes.StatusConfirmed
		res.ConfirmedAt = &now
		if err := u.resRepo.Update(ctx, res); err != nil {
			return err
		}

		confirmedRes = res
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &response.ConfirmReservationResponse{
		Status:        "success",
		ReservationID: confirmedRes.ReservationID,
		ConfirmedAt:   *confirmedRes.ConfirmedAt,
	}, nil
}

func (u *ReservationUseCaseImpl) ReleaseExpiredReservations(ctx context.Context) (int, error) {
	now, err := u.txMgr.Now(ctx)
	if err != nil {
		return 0, err
	}

	expiredList, err := u.resRepo.FindExpiredPending(ctx, now, 50)
	if err != nil {
		return 0, err
	}

	releasedCount := 0
	for _, target := range expiredList {
		err := u.txMgr.WithinTx(ctx, func(ctx context.Context) error {
			// Re-read the clock inside the transaction.
			now, err := u.txMgr.Now(ctx)
			if err != nil {
				return err
			}

			res, err := u.resRepo.GetByIDForUpdate(ctx, target.ReservationID)
			if err != nil {
				return err
			}

			if !res.IsPending() || !res.IsExpired(now) {
				return nil
			}

			// Availability is derived from status + expires_at, so marking
			// the reservation expired is pure bookkeeping cleanup.
			res.Status = domainRes.StatusExpired
			return u.resRepo.Update(ctx, res)
		})

		if err != nil {
			logger.Error("Failed to release expired reservation", "reservation_id", target.ReservationID, "error", err.Error())
			continue
		}

		releasedCount++
	}

	return releasedCount, nil
}
