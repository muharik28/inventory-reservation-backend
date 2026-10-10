package request

type ReserveStockRequest struct {
	UserID string `json:"user_id" binding:"required" example:"USR-123"`
	ItemID string `json:"item_id" binding:"required" example:"ITEM-001"`
	// Quantity is validated by the domain (> 0) so zero/negative values map
	// to the INVALID_QUANTITY error code instead of a generic bind error.
	Quantity int `json:"quantity" example:"2"`
}
