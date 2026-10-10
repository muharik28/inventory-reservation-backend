-- Migrations: 20261010_init.sql
--
-- Stock model: inventories.total_stock is the physical stock. Reserved stock
-- is never stored. It is derived from active reservations (status PENDING and
-- expires_at > NOW()). available_stock = total_stock - active_reserved.

CREATE TABLE IF NOT EXISTS inventories (
    item_id VARCHAR(100) PRIMARY KEY,
    total_stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_inventories_total_stock_nonneg CHECK (total_stock >= 0)
);

CREATE TABLE IF NOT EXISTS reservations (
    reservation_id VARCHAR(100) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    item_id VARCHAR(100) NOT NULL REFERENCES inventories(item_id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    expires_at TIMESTAMPTZ NOT NULL,
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_reservations_quantity_positive CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_reservations_status_expires_at ON reservations(status, expires_at);
CREATE INDEX IF NOT EXISTS idx_reservations_item_id ON reservations(item_id);

-- Seed initial data for testing matching example in specification
INSERT INTO inventories (item_id, total_stock, created_at, updated_at)
VALUES ('item_4021', 100, NOW(), NOW())
ON CONFLICT (item_id) DO NOTHING;

-- Seed 15 units of active reservations for the demo item so the dashboard
-- immediately shows total=100, reserved=15, available=85.
INSERT INTO reservations (reservation_id, user_id, item_id, quantity, status, expires_at, created_at, updated_at)
SELECT 'seed_res_' || g, 'demo_user', 'item_4021', 1, 'PENDING', NOW() + INTERVAL '2 hours', NOW(), NOW()
FROM generate_series(1, 15) AS g
ON CONFLICT (reservation_id) DO NOTHING;
