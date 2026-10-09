package info

import (
	"context"

	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities/info"
)

type InfoRepository interface {
	GetInfo(ctx context.Context) (*info.ServiceInfo, error)
}
