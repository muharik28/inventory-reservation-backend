package impl

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muharik28/inventory-reservation-backend/internal/application/handlers/inventory"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	usecase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/inventory"
	domain "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
)

type InventoryHandlerImpl struct {
	useCase usecase.InventoryUseCase
}

func NewInventoryHandler(useCase usecase.InventoryUseCase) inventory.InventoryHandler {
	return &InventoryHandlerImpl{
		useCase: useCase,
	}
}

func (h *InventoryHandlerImpl) GetStock(c *gin.Context) {
	var req request.StockQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Query parameter 'item_id' is required",
		})
		return
	}

	res, err := h.useCase.GetStock(c.Request.Context(), req.ItemID)
	if err != nil {
		if errors.Is(err, domain.ErrInventoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  "Item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *InventoryHandlerImpl) CreateItem(c *gin.Context) {
	var req request.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	res, err := h.useCase.CreateOrUpdateItem(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, res)
}
