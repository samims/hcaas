package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"

	"github.com/samims/hcaas/internal/handler"
	"github.com/samims/hcaas/internal/logger"
	"github.com/samims/hcaas/internal/router"
	"github.com/samims/hcaas/internal/service"
	"github.com/samims/hcaas/internal/storage"
)

func main() {
	l := logger.NewJSONLogger()
	slog.SetDefault(l)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ctx := context.Background()
	dbPool, err := storage.NewPostgresPool(ctx)

	if err != nil {
		defer dbPool.Close()
		log.Fatalf("Failed to connect to databse: %v", err)
	}

	ps := storage.NewPostgresStorage(dbPool)
	svc := service.NewURLService(ps, l)
	h := handler.NewURLHandler(svc, l)
	port := ":8080"
	// Setup router
	r := router.NewRouter(h)
	server := &http.Server{
		Addr:    port,
		Handler: r,
	}

	go func() {
		log.Println("Server Started on: ", port)
	}()

	// GracefulShutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")
	ctxTimeOut, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	if err := server.Shutdown(ctxTimeOut); err != nil {
		log.Fatalf("Shutdown failed: %v", err)

	}

	log.Println("Server exited cleanly")
}
