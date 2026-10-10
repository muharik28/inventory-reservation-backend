package routes

import (
	"github.com/gin-gonic/gin"
	_ "github.com/muharik28/inventory-reservation-backend/docs/api"
	infoHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/info"
	invHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/inventory"
	resHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/reservation"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type RouterParams struct {
	InfoHandler        infoHandler.InfoHandler
	InventoryHandler   invHandler.InventoryHandler
	ReservationHandler resHandler.ReservationHandler
}

func SetupRouter(r *gin.Engine, p RouterParams) {
	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Base health checks
	r.GET("/health", p.InfoHandler.HealthCheck)
	r.GET("/info", p.InfoHandler.GetInfo)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Info
		v1.GET("/info", p.InfoHandler.GetInfo)

		// Inventory
		inventory := v1.Group("/inventory")
		{
			inventory.GET("/stock", p.InventoryHandler.GetStock)
			inventory.POST("/reserve", p.ReservationHandler.Reserve)
			inventory.POST("/confirm", p.ReservationHandler.Confirm)
			inventory.POST("/items", p.InventoryHandler.CreateItem)
		}
	}
}

