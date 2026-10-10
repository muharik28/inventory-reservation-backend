package impl

import (
	"context"
	"time"

	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
	repo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/inventory"
	resRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/reservation"
	"github.com/muharik28/inventory-reservation-backend/internal/application/usecases/inventory"
	domain "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
)

// Clock is the persistence port this use case needs: reading the PostgreSQL
// clock. Defined here (consumer side) so the use case never depends on a
// concrete ORM type; *sql.BunTxManager satisfies it.
type Clock interface {
	Now(ctx context.Context) (time.Time, error)
}

type InventoryUseCaseImpl struct {
	clock   Clock
	invRepo repo.InventoryRepository
	resRepo resRepo.ReservationRepository
}

func NewInventoryUseCase(clock Clock, invRepo repo.InventoryRepository, resRepo resRepo.ReservationRepository) inventory.InventoryUseCase {
	return &InventoryUseCaseImpl{
		clock:   clock,
		invRepo: invRepo,
		resRepo: resRepo,
	}
}

func (u *InventoryUseCaseImpl) GetStock(ctx context.Context, itemID string) (*response.StockResponse, error) {
	item, err := u.invRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	// Reserved stock is derived from active (PENDING, unexpired according to
	// the PostgreSQL clock) reservations instead of a stored counter, so it
	// stays correct even if the expiry worker has not run yet.
	now, err := u.clock.Now(ctx)
	if err != nil {
		return nil, err
	}

	activeReserved, err := u.resRepo.SumActiveByItem(ctx, itemID, now)
	if err != nil {
		return nil, err
	}

	return &response.StockResponse{
		ItemID:         item.ItemID,
		TotalStock:     item.TotalStock,
		ReservedStock:  activeReserved,
		AvailableStock: item.AvailableStock(activeReserved),
	}, nil
}

func (u *InventoryUseCaseImpl) CreateOrUpdateItem(ctx context.Context, req request.CreateItemRequest) (*response.StockResponse, error) {
	item := &domain.Inventory{
		ItemID:     req.ItemID,
		TotalStock: req.TotalStock,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := u.invRepo.Save(ctx, item); err != nil {
		return nil, err
	}

	now, err := u.clock.Now(ctx)
	if err != nil {
		return nil, err
	}

	activeReserved, err := u.resRepo.SumActiveByItem(ctx, req.ItemID, now)
	if err != nil {
		return nil, err
	}

	return &response.StockResponse{
		ItemID:         item.ItemID,
		TotalStock:     item.TotalStock,
		ReservedStock:  activeReserved,
		AvailableStock: item.AvailableStock(activeReserved),
	}, nil
}
