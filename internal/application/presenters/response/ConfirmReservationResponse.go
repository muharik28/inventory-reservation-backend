package response

import "time"

type ConfirmReservationResponse struct {
	Status        string    `json:"status" example:"confirmed"`
	ReservationID string    `json:"reservation_id" example:"res-abc-123"`
	ConfirmedAt   time.Time `json:"confirmed_at" example:"2026-10-10T08:31:00Z"`
}

