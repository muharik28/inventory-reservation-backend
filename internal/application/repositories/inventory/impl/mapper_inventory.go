package impl

import (
	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
)

func ToInventoryDomain(m *InventoryModel) *inventory.Inventory {
	if m == nil {
		return nil
	}
	return &inventory.Inventory{
		ItemID:        m.ItemID,
		TotalStock:    m.TotalStock,
		ReservedStock: m.ReservedStock,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func ToInventoryModel(d *inventory.Inventory) *InventoryModel {
	if d == nil {
		return nil
	}
	return &InventoryModel{
		ItemID:        d.ItemID,
		TotalStock:    d.TotalStock,
		ReservedStock: d.ReservedStock,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}
