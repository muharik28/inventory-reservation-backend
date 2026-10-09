package info

import (
	"context"

	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
)

type InfoUseCase interface {
	GetInfo(ctx context.Context) (*response.InfoDetailResponse, error)
}
