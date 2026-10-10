package request

type ConfirmReservationRequest struct {
	ReservationID string `json:"reservation_id" binding:"required" example:"res-abc-123"`
}

