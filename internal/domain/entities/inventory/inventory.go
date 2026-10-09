package inventory

import (
	"errors"
	"time"
)

var (
	ErrInventoryNotFound = errors.New("inventory item not found")
	ErrInsufficientStock = errors.New("insufficient stock for reservation")
	ErrInvalidQuantity   = errors.New("quantity must be greater than zero")
)

type Inventory struct {
	ItemID        string    `json:"item_id"`
	TotalStock    int       `json:"total_stock"`
	ReservedStock int       `json:"reserved_stock"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (i *Inventory) AvailableStock() int {
	available := i.TotalStock - i.ReservedStock
	if available < 0 {
		return 0
	}
	return available
}

func (i *Inventory) CanReserve(qty int) bool {
	if qty <= 0 {
		return false
	}
	return i.AvailableStock() >= qty
}
