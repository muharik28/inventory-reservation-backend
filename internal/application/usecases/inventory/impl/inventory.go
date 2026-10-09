package impl

import (
	"context"
	"time"

	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
	repo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/inventory"
	"github.com/muharik28/inventory-reservation-backend/internal/application/usecases/inventory"
	domain "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
)

type InventoryUseCaseImpl struct {
	invRepo repo.InventoryRepository
}

func NewInventoryUseCase(invRepo repo.InventoryRepository) inventory.InventoryUseCase {
	return &InventoryUseCaseImpl{
		invRepo: invRepo,
	}
}

func (u *InventoryUseCaseImpl) GetStock(ctx context.Context, itemID string) (*response.StockResponse, error) {
	item, err := u.invRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	return &response.StockResponse{
		ItemID:         item.ItemID,
		TotalStock:     item.TotalStock,
		ReservedStock:  item.ReservedStock,
		AvailableStock: item.AvailableStock(),
	}, nil
}

func (u *InventoryUseCaseImpl) CreateOrUpdateItem(ctx context.Context, req request.CreateItemRequest) (*response.StockResponse, error) {
	existing, err := u.invRepo.GetByID(ctx, req.ItemID)
	if err != nil && err != domain.ErrInventoryNotFound {
		return nil, err
	}

	reservedStock := 0
	if existing != nil {
		reservedStock = existing.ReservedStock
	}

	item := &domain.Inventory{
		ItemID:        req.ItemID,
		TotalStock:    req.TotalStock,
		ReservedStock: reservedStock,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := u.invRepo.Save(ctx, nil, item); err != nil {
		return nil, err
	}

	return &response.StockResponse{
		ItemID:         item.ItemID,
		TotalStock:     item.TotalStock,
		ReservedStock:  item.ReservedStock,
		AvailableStock: item.AvailableStock(),
	}, nil
}
