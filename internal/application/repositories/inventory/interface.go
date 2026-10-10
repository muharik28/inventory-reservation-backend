package inventory

import (
	"context"

	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
)

type InventoryRepository interface {
	GetByID(ctx context.Context, itemID string) (*inventory.Inventory, error)
	GetByIDForUpdate(ctx context.Context, itemID string) (*inventory.Inventory, error)
	Save(ctx context.Context, item *inventory.Inventory) error
	UpdateStock(ctx context.Context, item *inventory.Inventory) error
}
