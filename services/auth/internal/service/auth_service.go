package service

import (
	"context"
	"errors"
	"log/slog"
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"

	appErr "github.com/samims/hcaas/services/auth/internal/errors"
	"github.com/samims/hcaas/services/auth/internal/model"
	"github.com/samims/hcaas/services/auth/internal/storage"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (*model.User, string, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	ValidateToken(token string) (string, error)
}

type authService struct {
	store    storage.UserStorage
	logger   *slog.Logger
	tokenSvc TokenService
}

func NewAuthService(store storage.UserStorage, logger *slog.Logger, tokenSvc TokenService) AuthService {
	l := logger.With("layer", "service", "component", "authService")
	return &authService{store: store, logger: l, tokenSvc: tokenSvc}
}

func (s *authService) Register(ctx context.Context, email, password string) (*model.User, error) {
	s.logger.Info("Register called", slog.String("email", email))

	if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(email) {
		s.logger.Error("Invalid email")
		return nil, appErr.ErrInvalidEmail
	}

	if len(password) == 0 {
		s.logger.Error("Invalid password")
		return nil, appErr.ErrInvalidInput
	}

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

func (s *authService) Login(ctx context.Context, email, password string) (*model.User, string, error) {
	s.logger.Info("Login called", slog.String("email", email))

	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn("User not found", slog.String("email", email))
			return nil, "", appErr.ErrUnauthorized
		}
		s.logger.Error("Failed to fetch user by email", slog.String("email", email), slog.Any("error", err))
		return nil, "", appErr.ErrInternal
	}
	s.logger.Info("Log in user found", slog.String("email", email))

	// Compare the provided password with the stored hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.Warn("Invalid password", slog.String("email", email))
		return nil, "", appErr.ErrUnauthorized
	}
	token, err := s.tokenSvc.GenerateToken(user)
	if err != nil {
		s.logger.Error("Token generation failed ", slog.String("email", email))
		return nil, "", appErr.ErrTokenGeneration
	}

	s.logger.Info("Token Generated successfully", slog.String("email", email))
	return user, token, nil
}

func (s *authService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error(
			"Failed to fetch user by email ",
			slog.String("email", email),
			slog.String("error", err.Error()),
		)
		return nil, appErr.ErrInternal
	}

	return user, nil
}

func (s *authService) ValidateToken(token string) (string, error) {
	s.logger.Info("ValidateToken called")
	userID, err := s.tokenSvc.ValidateToken(token)
	if err != nil {
		s.logger.Info("Token validation failed", slog.String("error", err.Error()))
		return "", err
	}
	return userID, nil

}
