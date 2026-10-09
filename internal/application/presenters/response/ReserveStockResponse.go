package response

import "time"

type ReserveStockResponse struct {
	Status        string    `json:"status"`
	ReservationID string    `json:"reservation_id"`
	ItemID        string    `json:"item_id"`
	Quantity      int       `json:"quantity"`
	ExpiresAt     time.Time `json:"expires_at"`
}
