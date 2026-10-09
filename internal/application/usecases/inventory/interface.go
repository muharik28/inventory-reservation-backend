package inventory

import (
	"context"

	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
)

type InventoryUseCase interface {
	GetStock(ctx context.Context, itemID string) (*response.StockResponse, error)
	CreateOrUpdateItem(ctx context.Context, req request.CreateItemRequest) (*response.StockResponse, error)
}
