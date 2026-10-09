package impl

import (
	"time"

	"github.com/uptrace/bun"
)

type InventoryModel struct {
	bun.BaseModel `bun:"table:inventories,alias:inv"`

	ItemID        string    `bun:"item_id,pk"`
	TotalStock    int       `bun:"total_stock,notnull"`
	ReservedStock int       `bun:"reserved_stock,notnull"`
	CreatedAt     time.Time `bun:"created_at,nullzero,notnull"`
	UpdatedAt     time.Time `bun:"updated_at,nullzero,notnull"`
}
