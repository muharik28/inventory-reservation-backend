package impl

import (
	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
)

func ToInventoryDomain(m *InventoryModel) *inventory.Inventory {
	if m == nil {
		return nil
	}
	return &inventory.Inventory{
		ItemID:     m.ItemID,
		TotalStock: m.TotalStock,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func ToInventoryModel(d *inventory.Inventory) *InventoryModel {
	if d == nil {
		return nil
	}
	return &InventoryModel{
		ItemID:     d.ItemID,
		TotalStock: d.TotalStock,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
}
