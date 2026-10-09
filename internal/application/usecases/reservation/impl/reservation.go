package impl

import (
	"context"
	"database/sql"
	"time"

	"github.com/uptrace/bun"
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

type ReservationUseCaseImpl struct {
	db      *bun.DB
	invRepo invRepo.InventoryRepository
	resRepo resRepo.ReservationRepository
	cfg     *configs.Config
}

func NewReservationUseCase(
	db *bun.DB,
	invRepo invRepo.InventoryRepository,
	resRepo resRepo.ReservationRepository,
	cfg *configs.Config,
) reservation.ReservationUseCase {
	return &ReservationUseCaseImpl{
		db:      db,
		invRepo: invRepo,
		resRepo: resRepo,
		cfg:     cfg,
	}
}

func (u *ReservationUseCaseImpl) ReserveStock(ctx context.Context, req request.ReserveStockRequest) (*response.ReserveStockResponse, error) {
	if req.Quantity <= 0 {
		return nil, domainInv.ErrInvalidQuantity
	}

	reservationID := utils.GenerateNumericID("res")
	now := time.Now().UTC()
	expiresAt := now.Add(u.cfg.ReservationTTL)

	res := &domainRes.Reservation{
		ReservationID: reservationID,
		UserID:        req.UserID,
		ItemID:        req.ItemID,
		Quantity:      req.Quantity,
		Status:        domainRes.StatusPending,
		ExpiresAt:     expiresAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := u.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		item, err := u.invRepo.GetByIDForUpdate(ctx, tx, req.ItemID)
		if err != nil {
			return err
		}

		if !item.CanReserve(req.Quantity) {
			return domainInv.ErrInsufficientStock
		}

		item.ReservedStock += req.Quantity
		if err := u.invRepo.UpdateStock(ctx, tx, item); err != nil {
			return err
		}

		if err := u.resRepo.Create(ctx, tx, res); err != nil {
			return err
		}

		return nil
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
	now := time.Now().UTC()
	var confirmedRes *domainRes.Reservation

	err := u.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		res, err := u.resRepo.GetByIDForUpdate(ctx, tx, req.ReservationID)
		if err != nil {
			return err
		}

		if !res.IsPending() {
			return domainRes.ErrReservationAlreadyHandled
		}

		if res.IsExpired(now) {
			return domainRes.ErrReservationExpired
		}

		item, err := u.invRepo.GetByIDForUpdate(ctx, tx, res.ItemID)
		if err != nil {
			return err
		}

		item.TotalStock -= res.Quantity
		item.ReservedStock -= res.Quantity
		if item.TotalStock < 0 {
			item.TotalStock = 0
		}
		if item.ReservedStock < 0 {
			item.ReservedStock = 0
		}

		if err := u.invRepo.UpdateStock(ctx, tx, item); err != nil {
			return err
		}

		res.Status = domainRes.StatusConfirmed
		res.ConfirmedAt = &now
		if err := u.resRepo.Update(ctx, tx, res); err != nil {
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
	now := time.Now().UTC()
	expiredList, err := u.resRepo.FindExpiredPending(ctx, now, 50)
	if err != nil {
		return 0, err
	}

	releasedCount := 0
	for _, target := range expiredList {
		err := u.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			res, err := u.resRepo.GetByIDForUpdate(ctx, tx, target.ReservationID)
			if err != nil {
				return err
			}

			if !res.IsPending() || !res.IsExpired(now) {
				return nil
			}

			item, err := u.invRepo.GetByIDForUpdate(ctx, tx, res.ItemID)
			if err != nil {
				return err
			}

			item.ReservedStock -= res.Quantity
			if item.ReservedStock < 0 {
				item.ReservedStock = 0
			}

			if err := u.invRepo.UpdateStock(ctx, tx, item); err != nil {
				return err
			}

			res.Status = domainRes.StatusExpired
			return u.resRepo.Update(ctx, tx, res)
		})

		if err != nil {
			logger.Error("Failed to release expired reservation", "reservation_id", target.ReservationID, "error", err.Error())
			continue
		}

		releasedCount++
	}

	return releasedCount, nil
}
