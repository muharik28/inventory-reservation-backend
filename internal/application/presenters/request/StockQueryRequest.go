package request

type StockQueryRequest struct {
	ItemID string `form:"item_id" binding:"required"`
}

type CreateItemRequest struct {
	ItemID     string `json:"item_id" binding:"required"`
	TotalStock int    `json:"total_stock" binding:"required,gte=0"`
}
