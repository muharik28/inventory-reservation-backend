package reservation

import (
	"errors"
	"time"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusConfirmed Status = "CONFIRMED"
	StatusCancelled Status = "CANCELLED"
	StatusExpired   Status = "EXPIRED"
)

var (
	ErrReservationNotFound       = errors.New("reservation not found")
	ErrReservationAlreadyHandled = errors.New("reservation is already confirmed, cancelled, or expired")
	ErrReservationExpired        = errors.New("reservation has expired")
)

type Reservation struct {
	ReservationID string     `json:"reservation_id"`
	UserID        string     `json:"user_id"`
	ItemID        string     `json:"item_id"`
	Quantity      int        `json:"quantity"`
	Status        Status     `json:"status"`
	ExpiresAt     time.Time  `json:"expires_at"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (r *Reservation) IsPending() bool {
	return r.Status == StatusPending
}

func (r *Reservation) IsExpired(now time.Time) bool {
	return r.Status == StatusPending && now.After(r.ExpiresAt)
}
