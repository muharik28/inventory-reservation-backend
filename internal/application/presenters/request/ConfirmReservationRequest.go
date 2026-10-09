package request

type ConfirmReservationRequest struct {
	ReservationID string `json:"reservation_id" binding:"required"`
}
