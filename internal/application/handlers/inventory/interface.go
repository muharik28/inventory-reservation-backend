package inventory

import "github.com/gin-gonic/gin"

type InventoryHandler interface {
	GetStock(c *gin.Context)
	CreateItem(c *gin.Context)
}
