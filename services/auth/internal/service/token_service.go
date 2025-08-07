package service

import (
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/samims/hcaas/services/auth/internal/model"
)

type TokenService interface {
	GenerateToken(user *model.User) (string, error)
	ValidateToken(tokenStr string) (string, string, error)
}

type jwtService struct {
	secret     string
	expiryTime time.Duration
	logger     *slog.Logger
}

func NewJWTService(secret string, expiry time.Duration, logger *slog.Logger) TokenService {
	return &jwtService{secret: secret, expiryTime: expiry, logger: logger}
}

func (s *jwtService) GenerateToken(user *model.User) (string, error) {
	if user == nil {
		s.logger.Error("error generating token")
		return "", errors.New("user is nil: cannot generate token")
	}
	s.logger.Info("token expiry time", slog.Duration("time", s.expiryTime))

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(s.expiryTime).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *jwtService) ValidateToken(tokenStr string) (string, string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(s.secret), nil
	})

	if err != nil || !token.Valid {
		s.logger.Error("Invalid token ")
		return "", "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		s.logger.Error("token verification failed malformed")
		return "", "", jwt.ErrTokenMalformed
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		s.logger.Error("token verification failed malformed!")
		return "", "", jwt.ErrTokenMalformed
	}
	email, ok := claims["email"].(string)

	return userID, email, nil
}
