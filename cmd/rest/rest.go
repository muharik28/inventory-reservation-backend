package rest

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/muharik28/inventory-reservation-backend/api/rest/api"
	"github.com/muharik28/inventory-reservation-backend/cmd"
	"github.com/muharik28/inventory-reservation-backend/configs"
	infoHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/info/impl"
	invHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/inventory/impl"
	resHandler "github.com/muharik28/inventory-reservation-backend/internal/application/handlers/reservation/impl"
	infoRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/info/impl"
	invRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/inventory/impl"
	resRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/reservation/impl"
	infoUseCase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/info/impl"
	invUseCase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/inventory/impl"
	resUseCase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/reservation/impl"
	"github.com/muharik28/inventory-reservation-backend/internal/pkg/database/sql"
	"github.com/muharik28/inventory-reservation-backend/internal/pkg/logger"
)

// @title Inventory Reservation API
// @version 1.0
// @description Backend REST API for Inventory Reservation System.
// @host localhost:8080
// @BasePath /
// @schemes http https
var RestCmd = &cobra.Command{
	Use:   "rest",
	Short: "Start the REST API server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := configs.LoadConfig()

		db, err := sql.NewBunDB(cfg.Database)
		if err != nil {
			logger.Warn("Database initial connection warning", "error", err.Error())
		} else {
			// Auto migrate if database is accessible
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_ = sql.AutoMigrate(ctx, db)
			cancel()
		}

		// Repositories
		iRepo := infoRepo.NewInfoRepository(db, cfg)
		inventoryRepo := invRepo.NewInventoryRepository(db)
		reservationRepo := resRepo.NewReservationRepository(db)

		// Use Cases
		iUseCase := infoUseCase.NewInfoUseCase(iRepo)
		inventoryUseCase := invUseCase.NewInventoryUseCase(inventoryRepo)
		reservationUseCase := resUseCase.NewReservationUseCase(db, inventoryRepo, reservationRepo, cfg)

		// Handlers
		iHandler := infoHandler.NewInfoHandler(iUseCase)
		inventoryHandler := invHandler.NewInventoryHandler(inventoryUseCase)
		reservationHandler := resHandler.NewReservationHandler(reservationUseCase)

		server := api.NewServer(cfg, iHandler, inventoryHandler, reservationHandler)

		go func() {
			if err := server.Start(); err != nil && err != http.ErrServerClosed {
				logger.Error("HTTP server failed to start", "error", err.Error())
				os.Exit(1)
			}
		}()

		// Wait for interrupt signal
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logger.Info("Shutting down REST server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Server forced to shutdown", "error", err.Error())
		}

		logger.Info("Server stopped cleanly")
	},
}

func init() {
	cmd.AddCommand(RestCmd)
}
