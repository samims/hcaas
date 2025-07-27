package service

import (
	"context"
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"

	appErr "github.com/samims/hcaas/services/auth/internal/errors"
	"github.com/samims/hcaas/services/auth/internal/model"
	"github.com/samims/hcaas/services/auth/internal/storage"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (*model.User, error)
}

type authService struct {
	store  storage.UserStorage
	logger *slog.Logger
}

func NewAuthService(store storage.UserStorage, logger *slog.Logger) AuthService {
	l := logger.With("layer", "service", "component", "authService")
	return &authService{store: store, logger: l}
}

func (s *authService) Register(ctx context.Context, email, password string) (*model.User, error) {
	s.logger.Info("Register called", slog.String("email", email))

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Password hashing failed", slog.Any("error", err))
		return nil, appErr.ErrInternal
	}

	createdUser, err := s.store.CreateUser(ctx, email, string(hashedPass))
	if err != nil {
		if errors.Is(err, appErr.ErrConflict) {
			s.logger.Warn("User already exists", slog.String("email", email))
			return nil, appErr.ErrConflict
		}
		s.logger.Error("User creation failed", slog.Any("error", err))
		return nil, appErr.ErrInternal
	}

	s.logger.Info("Register succeeded", slog.String("email", email))
	return createdUser, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*model.User, error) {
	s.logger.Info("Login called", slog.String("email", email))

	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			s.logger.Warn("User not found", slog.String("email", email))
			return nil, appErr.ErrNotFound
		}
		s.logger.Error("User fetch failed", slog.Any("error", err))
		return nil, appErr.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.Warn("Invalid credentials", slog.String("email", email))
		return nil, appErr.ErrUnauthorized
	}

	s.logger.Info("Login succeeded", slog.String("email", email))
	return user, nil
}
