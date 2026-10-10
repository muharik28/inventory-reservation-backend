package request

type StockQueryRequest struct {
	ItemID string `form:"item_id" binding:"required" example:"ITEM-001"`
}

type CreateItemRequest struct {
	ItemID     string `json:"item_id" binding:"required" example:"ITEM-001"`
	TotalStock int    `json:"total_stock" binding:"required,gte=0" example:"100"`
}

