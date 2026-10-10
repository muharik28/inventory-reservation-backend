package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/muharik28/inventory-reservation-backend/configs"
	"github.com/muharik28/inventory-reservation-backend/internal/pkg/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func NewBunDB(cfg configs.DatabaseConfig) (*bun.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	connector := pgdriver.NewConnector(pgdriver.WithDSN(dsn))
	sqldb := sql.OpenDB(connector)

	sqldb.SetMaxOpenConns(25)
	sqldb.SetMaxIdleConns(10)
	sqldb.SetConnMaxLifetime(5 * time.Minute)

	db := bun.NewDB(sqldb, pgdialect.New())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Warn("Failed to ping PostgreSQL database (service can still start for offline dev)", "error", err.Error())
		return db, err
	}

	logger.Info("Successfully connected to PostgreSQL via Bun ORM")
	return db, nil
}
