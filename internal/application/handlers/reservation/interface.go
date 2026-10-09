package reservation

import "github.com/gin-gonic/gin"

type ReservationHandler interface {
	Reserve(c *gin.Context)
	Confirm(c *gin.Context)
}
