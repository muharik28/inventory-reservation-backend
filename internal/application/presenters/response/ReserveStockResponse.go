package response

import "time"

type ReserveStockResponse struct {
	Status        string    `json:"status" example:"reserved"`
	ReservationID string    `json:"reservation_id" example:"res-abc-123"`
	ItemID        string    `json:"item_id" example:"ITEM-001"`
	Quantity      int       `json:"quantity" example:"2"`
	ExpiresAt     time.Time `json:"expires_at" example:"2026-10-10T08:35:00Z"`
}

