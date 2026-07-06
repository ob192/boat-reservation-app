package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
)

// @title           Harbour & Wave API
// @version         1.0
// @description     Boat reservation backend
// @BasePath        /api
func main() {
	// 1. Load .env BEFORE Wire invokes config.Load()
	_ = godotenv.Load()

	// 2. Execute Wire-generated DI constructor
	server, err := InitializeServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize server: %v\n", err)
		os.Exit(1)
	}

	// 3. Database auto-migrations — run on a dedicated, non-pooled connection.
	// See platform.NewMigrationDB for why this can't share the pooled runtime connection.
	migrationDB, err := platform.NewMigrationDB(server.Config.DirectDatabaseURL)
	if err != nil {
		server.Log.Error("open migration db failed", "err", err)
		os.Exit(1)
	}
	migrateErr := migrationDB.AutoMigrate(
		&model.Booking{},
		&model.Slot{},
		&model.DateBlock{},
		&model.SystemSettings{},
		&model.Admin{},
		&model.Promocode{},
	)
	if sqlDB, dbErr := migrationDB.DB(); dbErr == nil {
		_ = sqlDB.Close()
	}
	if migrateErr != nil {
		server.Log.Error("auto-migrate failed", "err", migrateErr)
		os.Exit(1)
	}

	// 4. Ensure system_settings singleton row exists at boot
	if _, err := server.SystemRepo.Get(context.Background()); err != nil {
		server.Log.Error("system settings init failed", "err", err)
		os.Exit(1)
	}

	// 5. Background expiry sweeper
	expiryCtx, expiryCancel := context.WithCancel(context.Background())
	defer expiryCancel()
	go runExpirySweeper(expiryCtx, server.BookingSvc, server.Log)

	// 6. HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:              ":" + server.Config.Port,
		Handler:           server.Router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		server.Log.Info("listening", "port", server.Config.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			server.Log.Error("listen failed", "err", err)
			os.Exit(1)
		}
	}()

	// Wait for SIGINT / SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	server.Log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		server.Log.Error("server shutdown error", "err", err)
	}
	expiryCancel()
	server.Log.Info("bye")
}

// runExpirySweeper periodically expires stale pending bookings.
func runExpirySweeper(ctx context.Context, bookingSvc service.BookingService, log *slog.Logger) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	sweep(ctx, bookingSvc, log)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sweep(ctx, bookingSvc, log)
		}
	}
}

func sweep(ctx context.Context, bookingSvc service.BookingService, log *slog.Logger) {
	n, err := bookingSvc.ExpireStale(ctx)
	if err != nil {
		log.Error("expire-stale failed", "err", err)
		return
	}
	if n > 0 {
		log.Info("expired pending bookings", "count", n)
	}
}
