package info

import "github.com/gin-gonic/gin"

type InfoHandler interface {
	GetInfo(c *gin.Context)
	HealthCheck(c *gin.Context)
}
