package impl

import (
	"context"

	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
	repo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/info"
	"github.com/muharik28/inventory-reservation-backend/internal/application/usecases/info"
)

type InfoUseCaseImpl struct {
	repo repo.InfoRepository
}

func NewInfoUseCase(repo repo.InfoRepository) info.InfoUseCase {
	return &InfoUseCaseImpl{
		repo: repo,
	}
}

func (u *InfoUseCaseImpl) GetInfo(ctx context.Context) (*response.InfoDetailResponse, error) {
	infoData, err := u.repo.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &response.InfoDetailResponse{
		ServiceName: infoData.ServiceName,
		Version:     infoData.Version,
		Environment: infoData.Environment,
		Status:      infoData.Status,
		DBStatus:    infoData.DBStatus,
		ServerTime:  infoData.ServerTime,
	}, nil
}
