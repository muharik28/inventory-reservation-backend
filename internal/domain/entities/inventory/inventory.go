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
	ItemID     string    `json:"item_id"`
	TotalStock int       `json:"total_stock"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AvailableStock computes the sellable stock given the currently active
// reserved quantity (active PENDING reservations with expires_at in the future).
func (i *Inventory) AvailableStock(reservedQty int) int {
	available := i.TotalStock - reservedQty
	if available < 0 {
		return 0
	}
	return available
}

func (i *Inventory) CanReserve(qty, reservedQty int) bool {
	if qty <= 0 {
		return false
	}
	return i.AvailableStock(reservedQty) >= qty
}
