package impl

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/muharik28/inventory-reservation-backend/configs"
	"github.com/muharik28/inventory-reservation-backend/internal/application/repositories/info"
	domain "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/info"
)

type InfoRepositoryImpl struct {
	db  *bun.DB
	cfg *configs.Config
}

func NewInfoRepository(db *bun.DB, cfg *configs.Config) info.InfoRepository {
	return &InfoRepositoryImpl{
		db:  db,
		cfg: cfg,
	}
}

func (r *InfoRepositoryImpl) GetInfo(ctx context.Context) (*domain.ServiceInfo, error) {
	dbStatus := "connected"
	if err := r.db.PingContext(ctx); err != nil {
		dbStatus = "disconnected"
	}

	model := &InfoModel{
		ServiceName: "inventory-reservation-backend",
		Version:     "1.0.0",
		Environment: r.cfg.AppEnv,
		Status:      "healthy",
		DBStatus:    dbStatus,
		ServerTime:  time.Now().UTC(),
	}

	return ToInfoDomain(model), nil
}
