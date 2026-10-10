package impl

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muharik28/inventory-reservation-backend/internal/application/handlers/info"
	"github.com/muharik28/inventory-reservation-backend/internal/application/presenters/response"
	usecase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/info"
	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities"
)

type InfoHandlerImpl struct {
	useCase usecase.InfoUseCase
}

func NewInfoHandler(useCase usecase.InfoUseCase) info.InfoHandler {
	return &InfoHandlerImpl{
		useCase: useCase,
	}
}

// GetInfo godoc
// @Summary Get service info
// @Description Retrieve service runtime metadata, environment, and database connectivity status
// @Tags Info
// @Produce json
// @Success 200 {object} response.InfoDetailResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /info [get]
// @Router /api/v1/info [get]
func (h *InfoHandlerImpl) GetInfo(c *gin.Context) {
	res, err := h.useCase.GetInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, entities.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// HealthCheck godoc
// @Summary Health check
// @Description Health check endpoint verifying service and database status
// @Tags Info
// @Produce json
// @Success 200 {object} response.HealthResponse
// @Failure 503 {object} entities.ErrorResponse
// @Router /health [get]
func (h *InfoHandlerImpl) HealthCheck(c *gin.Context) {
	res, err := h.useCase.GetInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, entities.ErrorResponse{
			Status: "unhealthy",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.HealthResponse{
		Status:   "healthy",
		Service:  res.ServiceName,
		DBStatus: res.DBStatus,
	})
}

