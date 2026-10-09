-- Migrations: 20210101_init.sql

CREATE TABLE IF NOT EXISTS inventories (
    item_id VARCHAR(100) PRIMARY KEY,
    total_stock INT NOT NULL DEFAULT 0,
    reserved_stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reservations_status_expires_at ON reservations(status, expires_at);
CREATE INDEX IF NOT EXISTS idx_reservations_item_id ON reservations(item_id);

-- Seed initial data for testing matching example in specification
INSERT INTO inventories (item_id, total_stock, reserved_stock, created_at, updated_at)
VALUES ('item_4021', 100, 15, NOW(), NOW())
ON CONFLICT (item_id) DO NOTHING;
