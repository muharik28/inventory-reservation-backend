package impl

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/muharik28/inventory-reservation-backend/internal/application/repositories/inventory"
	domain "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
	dbsql "github.com/muharik28/inventory-reservation-backend/internal/pkg/database/sql"
	"github.com/uptrace/bun"
)

type InventoryRepositoryImpl struct {
	db *bun.DB
}

func NewInventoryRepository(db *bun.DB) inventory.InventoryRepository {
	return &InventoryRepositoryImpl{
		db: db,
	}
}

func (r *InventoryRepositoryImpl) GetByID(ctx context.Context, itemID string) (*domain.Inventory, error) {
	model := new(InventoryModel)
	err := r.db.NewSelect().
		Model(model).
		Where("item_id = ?", itemID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInventoryNotFound
		}
		return nil, err
	}

	return ToInventoryDomain(model), nil
}

// GetByIDForUpdate locks the inventory row (SELECT ... FOR UPDATE) on the
// executor carried by ctx, so it participates in the ambient transaction.
func (r *InventoryRepositoryImpl) GetByIDForUpdate(ctx context.Context, itemID string) (*domain.Inventory, error) {
	model := new(InventoryModel)
	err := dbsql.TxOrDB(ctx, r.db).NewSelect().
		Model(model).
		Where("item_id = ?", itemID).
		For("UPDATE").
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInventoryNotFound
		}
		return nil, err
	}

	return ToInventoryDomain(model), nil
}

func (r *InventoryRepositoryImpl) Save(ctx context.Context, item *domain.Inventory) error {
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	item.UpdatedAt = time.Now()

	model := ToInventoryModel(item)
	_, err := dbsql.TxOrDB(ctx, r.db).NewInsert().
		Model(model).
		On("CONFLICT (item_id) DO UPDATE").
		Set("total_stock = EXCLUDED.total_stock").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)

	return err
}

func (r *InventoryRepositoryImpl) UpdateStock(ctx context.Context, item *domain.Inventory) error {
	item.UpdatedAt = time.Now()
	model := ToInventoryModel(item)

	_, err := dbsql.TxOrDB(ctx, r.db).NewUpdate().
		Model(model).
		Column("total_stock", "updated_at").
		Where("item_id = ?", item.ItemID).
		Exec(ctx)

	return err
}
