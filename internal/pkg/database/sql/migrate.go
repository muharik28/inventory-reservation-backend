package sql

import (
	"context"
	"time"

	"github.com/muharik28/inventory-reservation-backend/internal/pkg/logger"
	"github.com/uptrace/bun"
)

func AutoMigrate(ctx context.Context, db *bun.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS inventories (
			item_id VARCHAR(100) PRIMARY KEY,
			total_stock INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT ck_inventories_total_stock_nonneg CHECK (total_stock >= 0)
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
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT ck_reservations_quantity_positive CHECK (quantity > 0)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_reservations_status_expires_at ON reservations(status, expires_at);`,
		`CREATE INDEX IF NOT EXISTS idx_reservations_item_id ON reservations(item_id);`,
		// Guard constraints for databases created before the checks existed.
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_inventories_total_stock_nonneg') THEN
				ALTER TABLE inventories ADD CONSTRAINT ck_inventories_total_stock_nonneg CHECK (total_stock >= 0);
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_reservations_quantity_positive') THEN
				ALTER TABLE reservations ADD CONSTRAINT ck_reservations_quantity_positive CHECK (quantity > 0);
			END IF;
		END $$;`,
	}

	for _, q := range queries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			logger.Error("Failed to run auto migration query", "query", q, "error", err.Error())
			return err
		}
	}

	// Ensure seed for item_4021 exists if table is empty or item missing
	_, _ = db.ExecContext(ctx, `
		INSERT INTO inventories (item_id, total_stock, created_at, updated_at)
		VALUES ('item_4021', 100, NOW(), NOW())
		ON CONFLICT (item_id) DO NOTHING;
	`)

	// Seed 15 units of active reservations so stock reads show
	// total=100, reserved=15, available=85 for the demo item.
	_, _ = db.ExecContext(ctx, `
		INSERT INTO reservations (reservation_id, user_id, item_id, quantity, status, expires_at, created_at, updated_at)
		SELECT 'seed_res_' || g, 'demo_user', 'item_4021', 1, 'PENDING', NOW() + INTERVAL '2 hours', NOW(), NOW()
		FROM generate_series(1, 15) AS g
		ON CONFLICT (reservation_id) DO NOTHING;
	`)

	logger.Info("Database tables migration verified successfully", "timestamp", time.Now())
	return nil
}
