package impl

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muharik28/inventory-reservation-backend/internal/application/handlers/inventory"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
	usecase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/inventory"
	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities"
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

// GetStock godoc
// @Summary Get stock by item ID
// @Description Retrieve current inventory stock (total, reserved, and available) for a specific item
// @Tags Inventory
// @Produce json
// @Param item_id query string true "Item ID"
// @Success 200 {object} response.StockResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 404 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /api/v1/inventory/stock [get]
func (h *InventoryHandlerImpl) GetStock(c *gin.Context) {
	var req request.StockQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Status: "error",
			Error:  "Query parameter 'item_id' is required",
		})
		return
	}

	var res *response.StockResponse
	res, err := h.useCase.GetStock(c.Request.Context(), req.ItemID)
	if err != nil {
		if errors.Is(err, domain.ErrInventoryNotFound) {
			c.JSON(http.StatusNotFound, entities.ErrorResponse{
				Status: "error",
				Error:  "Item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// CreateItem godoc
// @Summary Create or update item stock
// @Description Create a new inventory item or update total stock for an existing item
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body request.CreateItemRequest true "Create item request payload"
// @Success 201 {object} response.StockResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /api/v1/inventory/items [post]
func (h *InventoryHandlerImpl) CreateItem(c *gin.Context) {
	var req request.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	var res *response.StockResponse
	res, err := h.useCase.CreateOrUpdateItem(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, res)
}

