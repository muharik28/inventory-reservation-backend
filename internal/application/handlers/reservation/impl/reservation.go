package impl

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muharik28/inventory-reservation-backend/internal/application/handlers/reservation"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
	usecase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/reservation"
	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities"
	domainInv "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/inventory"
	domainRes "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/reservation"
)

type ReservationHandlerImpl struct {
	useCase usecase.ReservationUseCase
}

func NewReservationHandler(useCase usecase.ReservationUseCase) reservation.ReservationHandler {
	return &ReservationHandlerImpl{
		useCase: useCase,
	}
}

// Reserve godoc
// @Summary Reserve inventory stock
// @Description Temporarily reserve a quantity of stock for a specific user with a TTL expiration
// @Tags Reservation
// @Accept json
// @Produce json
// @Param request body request.ReserveStockRequest true "Reserve stock request payload"
// @Success 200 {object} response.ReserveStockResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 404 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /api/v1/inventory/reserve [post]
func (h *ReservationHandlerImpl) Reserve(c *gin.Context) {
	var req request.ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Status: "error",
			Code:   "INVALID_REQUEST",
			Error:  err.Error(),
		})
		return
	}

	var res *response.ReserveStockResponse
	res, err := h.useCase.ReserveStock(c.Request.Context(), req)
	if err != nil {
		status, code := http.StatusInternalServerError, "INTERNAL_ERROR"
		switch {
		case errors.Is(err, domainInv.ErrInventoryNotFound):
			status, code = http.StatusNotFound, "ITEM_NOT_FOUND"
		case errors.Is(err, domainInv.ErrInvalidQuantity):
			status, code = http.StatusBadRequest, "INVALID_QUANTITY"
		case errors.Is(err, domainInv.ErrInsufficientStock):
			status, code = http.StatusBadRequest, "INSUFFICIENT_STOCK"
		}
		c.JSON(status, entities.ErrorResponse{
			Status: "error",
			Code:   code,
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Confirm godoc
// @Summary Confirm reservation
// @Description Confirm an active stock reservation before it expires
// @Tags Reservation
// @Accept json
// @Produce json
// @Param request body request.ConfirmReservationRequest true "Confirm reservation request payload"
// @Success 200 {object} response.ConfirmReservationResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 404 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /api/v1/inventory/confirm [post]
func (h *ReservationHandlerImpl) Confirm(c *gin.Context) {
	var req request.ConfirmReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entities.ErrorResponse{
			Status: "error",
			Code:   "INVALID_REQUEST",
			Error:  err.Error(),
		})
		return
	}

	var res *response.ConfirmReservationResponse
	res, err := h.useCase.ConfirmReservation(c.Request.Context(), req)
	if err != nil {
		status, code := http.StatusInternalServerError, "INTERNAL_ERROR"
		switch {
		case errors.Is(err, domainRes.ErrReservationNotFound):
			status, code = http.StatusNotFound, "RESERVATION_NOT_FOUND"
		case errors.Is(err, domainRes.ErrReservationAlreadyHandled):
			status, code = http.StatusBadRequest, "RESERVATION_ALREADY_HANDLED"
		case errors.Is(err, domainRes.ErrReservationExpired):
			status, code = http.StatusBadRequest, "RESERVATION_EXPIRED"
		}
		c.JSON(status, entities.ErrorResponse{
			Status: "error",
			Code:   code,
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}
