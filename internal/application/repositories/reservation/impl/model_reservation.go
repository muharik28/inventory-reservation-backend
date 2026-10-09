package impl

import (
	"time"

	"github.com/uptrace/bun"
)

type ReservationModel struct {
	bun.BaseModel `bun:"table:reservations,alias:res"`

	ReservationID string     `bun:"reservation_id,pk"`
	UserID        string     `bun:"user_id,notnull"`
	ItemID        string     `bun:"item_id,notnull"`
	Quantity      int        `bun:"quantity,notnull"`
	Status        string     `bun:"status,notnull"`
	ExpiresAt     time.Time  `bun:"expires_at,notnull"`
	ConfirmedAt   *time.Time `bun:"confirmed_at"`
	CreatedAt     time.Time  `bun:"created_at,nullzero,notnull"`
	UpdatedAt     time.Time  `bun:"updated_at,nullzero,notnull"`
}
