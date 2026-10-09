package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muharik28/inventory-reservation-backend/api/rest/api/routes"
	"github.com/muharik28/inventory-reservation-backend/configs"
	infoHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/info"
	invHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/inventory"
	resHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/reservation"
	"github.com/muharik28/inventory-reservation-backend/internal/pkg/logger"
)

type Server struct {
	cfg        *configs.Config
	router     *gin.Engine
	httpServer *http.Server
}

func NewServer(
	cfg *configs.Config,
	infoH infoHandler.InfoHandler,
	invH invHandler.InventoryHandler,
	resH resHandler.ReservationHandler,
) *Server {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	// Simple CORS middleware
	engine.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	routes.SetupRouter(engine, routes.RouterParams{
		InfoHandler:        infoH,
		InventoryHandler:   invH,
		ReservationHandler: resH,
	})

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Server{
		cfg:        cfg,
		router:     engine,
		httpServer: srv,
	}
}

func (s *Server) Start() error {
	logger.Info("Starting HTTP server", "port", s.cfg.AppPort, "env", s.cfg.AppEnv)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) GetEngine() *gin.Engine {
	return s.router
}
