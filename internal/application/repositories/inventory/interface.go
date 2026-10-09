package inventory

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
)

type InventoryRepository interface {
	GetByID(ctx context.Context, itemID string) (*inventory.Inventory, error)
	GetByIDForUpdate(ctx context.Context, db bun.IDB, itemID string) (*inventory.Inventory, error)
	Save(ctx context.Context, db bun.IDB, item *inventory.Inventory) error
	UpdateStock(ctx context.Context, db bun.IDB, item *inventory.Inventory) error
}
