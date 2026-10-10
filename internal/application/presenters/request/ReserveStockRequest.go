package request

type ReserveStockRequest struct {
	UserID   string `json:"user_id" binding:"required" example:"USR-123"`
	ItemID   string `json:"item_id" binding:"required" example:"ITEM-001"`
	Quantity int    `json:"quantity" binding:"required,gt=0" example:"2"`
}

