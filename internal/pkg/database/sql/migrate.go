package sql

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/muharik28/inventory-reservation-backend/internal/pkg/logger"
)

func AutoMigrate(ctx context.Context, db *bun.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS inventories (
			item_id VARCHAR(100) PRIMARY KEY,
			total_stock INT NOT NULL DEFAULT 0,
			reserved_stock INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS reservations (
			reservation_id VARCHAR(100) PRIMARY KEY,
			user_id VARCHAR(100) NOT NULL,
			item_id VARCHAR(100) NOT NULL,
			quantity INT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
			expires_at TIMESTAMPTZ NOT NULL,
			confirmed_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_reservations_status_expires_at ON reservations(status, expires_at);`,
		`CREATE INDEX IF NOT EXISTS idx_reservations_item_id ON reservations(item_id);`,
	}

	for _, q := range queries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			logger.Error("Failed to run auto migration query", "query", q, "error", err.Error())
			return err
		}
	}

	// Ensure seed for item_4021 exists if table is empty or item missing
	_, _ = db.ExecContext(ctx, `
		INSERT INTO inventories (item_id, total_stock, reserved_stock, created_at, updated_at)
		VALUES ('item_4021', 100, 15, NOW(), NOW())
		ON CONFLICT (item_id) DO NOTHING;
	`)

	logger.Info("Database tables migration verified successfully", "timestamp", time.Now())
	return nil
}
