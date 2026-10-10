package impl

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/muharik28/inventory-reservation-backend/internal/application/repositories/inventory"
	domain "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
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

func (r *InventoryRepositoryImpl) GetByIDForUpdate(ctx context.Context, db bun.IDB, itemID string) (*domain.Inventory, error) {
	targetDB := db
	if targetDB == nil {
		targetDB = r.db
	}

	model := new(InventoryModel)
	err := targetDB.NewSelect().
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

func (r *InventoryRepositoryImpl) Save(ctx context.Context, db bun.IDB, item *domain.Inventory) error {
	targetDB := db
	if targetDB == nil {
		targetDB = r.db
	}

	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	item.UpdatedAt = time.Now()

	model := ToInventoryModel(item)
	_, err := targetDB.NewInsert().
		Model(model).
		On("CONFLICT (item_id) DO UPDATE").
		Set("total_stock = EXCLUDED.total_stock").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)

	return err
}

func (r *InventoryRepositoryImpl) UpdateStock(ctx context.Context, db bun.IDB, item *domain.Inventory) error {
	targetDB := db
	if targetDB == nil {
		targetDB = r.db
	}

	item.UpdatedAt = time.Now()
	model := ToInventoryModel(item)

	_, err := targetDB.NewUpdate().
		Model(model).
		Column("total_stock", "updated_at").
		Where("item_id = ?", item.ItemID).
		Exec(ctx)

	return err
}
