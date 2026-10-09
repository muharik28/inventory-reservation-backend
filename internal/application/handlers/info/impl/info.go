package impl

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muharik28/inventory-reservation-backend/internal/application/handlers/info"
	usecase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/info"
)

type InfoHandlerImpl struct {
	useCase usecase.InfoUseCase
}

func NewInfoHandler(useCase usecase.InfoUseCase) info.InfoHandler {
	return &InfoHandlerImpl{
		useCase: useCase,
	}
}

func (h *InfoHandlerImpl) GetInfo(c *gin.Context) {
	res, err := h.useCase.GetInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *InfoHandlerImpl) HealthCheck(c *gin.Context) {
	res, err := h.useCase.GetInfo(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   res.ServiceName,
		"db_status": res.DBStatus,
	})
}
