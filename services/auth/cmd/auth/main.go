package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/samims/hcaas/services/auth/internal/handler"
	"github.com/samims/hcaas/services/auth/internal/logger"
	customMiddleware "github.com/samims/hcaas/services/auth/internal/middleware"
	"github.com/samims/hcaas/services/auth/internal/service"
	"github.com/samims/hcaas/services/auth/internal/storage"

	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	_ = godotenv.Load()

	l := logger.NewJSONLogger()
	r := chi.NewRouter()

	dbPool, err := storage.NewPostgresPool(ctx)
	if err != nil {
		l.Error("Failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	userStorage := storage.NewUserStorage(dbPool)

	secret := os.Getenv("SECRET_KEY")
	expiry := os.Getenv("AUTH_EXPIRY")
	exp, err := strconv.Atoi(expiry)
	if err != nil {
		l.Error("Error converting expiration duration to int ", err)
		return
	}

	expiryDuration := time.Duration(exp)

	tokenSvc := service.NewJWTService(secret, expiryDuration)
	authSvc := service.NewAuthService(userStorage, l)
	healthSvc := service.NewHealthService(userStorage, l)

	authHandler := handler.NewAuthHandler(authSvc, l)
	healthHandler := handler.NewHealthHandler(healthSvc, l)

	r = chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// public
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// protected
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(tokenSvc))
		r.Get("/me", authHandler.GetUser)
	})

	r.Get("/readyz", healthHandler.Readiness)
	r.Get("/healthz", healthHandler.Liveness)

	port := ":8081"

	server := &http.Server{Addr: port, Handler: r}

	go func() {
		l.Info("Server started", "addr", port)
		if err := server.ListenAndServe(); err != nil {
			l.Error("Failed to start server", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	<-quit
	l.Info("Shutting down server")
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxTimeout); err != nil {
		l.Error("Shutdown failed", "err", err)
	} else {
		l.Info("Server exited cleanly")
	}

}
