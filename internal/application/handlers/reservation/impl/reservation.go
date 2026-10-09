package impl

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muharik28/inventory-reservation-backend/internal/application/handlers/reservation"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/request"
	usecase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/reservation"
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

func (h *ReservationHandlerImpl) Reserve(c *gin.Context) {
	var req request.ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	res, err := h.useCase.ReserveStock(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domainInv.ErrInventoryNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}
		if errors.Is(err, domainInv.ErrInsufficientStock) || errors.Is(err, domainInv.ErrInvalidQuantity) {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  err.Error(),
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

func (h *ReservationHandlerImpl) Confirm(c *gin.Context) {
	var req request.ConfirmReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	res, err := h.useCase.ConfirmReservation(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domainRes.ErrReservationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}
		if errors.Is(err, domainRes.ErrReservationAlreadyHandled) || errors.Is(err, domainRes.ErrReservationExpired) {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  err.Error(),
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
