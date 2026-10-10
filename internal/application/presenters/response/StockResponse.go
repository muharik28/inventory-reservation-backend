package response

type StockResponse struct {
	ItemID         string `json:"item_id" example:"ITEM-001"`
	TotalStock     int    `json:"total_stock" example:"100"`
	ReservedStock  int    `json:"reserved_stock" example:"10"`
	AvailableStock int    `json:"available_stock" example:"90"`
}

