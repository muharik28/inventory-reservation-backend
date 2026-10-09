package response

import "time"

type ConfirmReservationResponse struct {
	Status        string    `json:"status"`
	ReservationID string    `json:"reservation_id"`
	ConfirmedAt   time.Time `json:"confirmed_at"`
}
