package scheduler

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/muharik28/inventory-reservation-backend/cmd"
	"github.com/muharik28/inventory-reservation-backend/configs"
	invRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/inventory/impl"
	resRepo "github.com/muharik28/inventory-reservation-backend/internal/application/repositories/reservation/impl"
	resUseCase "github.com/muharik28/inventory-reservation-backend/internal/application/usecases/reservation/impl"
	"github.com/muharik28/inventory-reservation-backend/internal/pkg/database/sql"
	"github.com/muharik28/inventory-reservation-backend/internal/pkg/logger"
	"github.com/spf13/cobra"
)

var SchedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Start the background scheduler to release expired reservations",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := configs.LoadConfig()

		db, err := sql.NewBunDB(cfg.Database)
		if err != nil {
			logger.Error("Database connection failed for scheduler", "error", err.Error())
			os.Exit(1)
		}

		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				logger.Error("Failed to close database connections", "error", closeErr.Error())
			}
		}()

		inventoryRepo := invRepo.NewInventoryRepository(db)
		reservationRepo := resRepo.NewReservationRepository(db)
		reservationUseCase := resUseCase.NewReservationUseCase(db, inventoryRepo, reservationRepo, cfg)

		logger.Info("Starting expired reservations cleanup scheduler", "interval", cfg.SchedulerInterval)

		ticker := time.NewTicker(cfg.SchedulerInterval)
		defer ticker.Stop()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				count, err := reservationUseCase.ReleaseExpiredReservations(ctx)
				cancel()

				if err != nil {
					logger.Error("Error during expired reservation cleanup", "error", err.Error())
				} else if count > 0 {
					logger.Info("Successfully released expired reservations", "count", count)
				}
			case <-quit:
				logger.Info("Stopping scheduler gracefully...")
				return
			}
		}
	},
}

func init() {
	cmd.AddCommand(SchedulerCmd)
}
