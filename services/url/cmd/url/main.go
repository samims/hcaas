package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"

	"github.com/samims/hcaas/internal/checker"
	"github.com/samims/hcaas/internal/handler"
	"github.com/samims/hcaas/internal/logger"
	"github.com/samims/hcaas/internal/metrics"
	"github.com/samims/hcaas/internal/router"
	"github.com/samims/hcaas/internal/service"
	"github.com/samims/hcaas/internal/storage"
)

func main() {
	l := logger.NewJSONLogger()
	slog.SetDefault(l)

	metrics.Init()

	if err := godotenv.Load(); err != nil {
		l.Error("Error loading .env file", "err", err)
	}

	ctx := context.Background()
	dbPool, err := storage.NewPostgresPool(ctx)
	if err != nil {
		l.Error("Failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Initialize layers
	ps := storage.NewPostgresStorage(dbPool)
	urlSvc := service.NewURLService(ps, l)
	healthSvc := service.NewHealthService(ps, l)

	httpClient := &http.Client{Timeout: 5 * time.Second}
	chkr := checker.NewURLChecker(urlSvc, l, httpClient, 1*time.Minute)
	go chkr.Start(ctx)

	urlHandler := handler.NewURLHandler(urlSvc, l)
	healthHandler := handler.NewHealthHandler(healthSvc, l)

	// Setup router and server
	port := ":8080"
	r := router.NewRouter(urlHandler, healthHandler)

	server := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		l.Info("Server started", "addr", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("Failed to start server", "err", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	l.Info("Shutting down server...")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxTimeout); err != nil {
		l.Error("Shutdown failed", "err", err)
	} else {
		l.Info("Server exited cleanly")
	}
}
