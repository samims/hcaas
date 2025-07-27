package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/samims/hcaas/services/auth/internal/model"
)

type TokenService interface {
	GenerateToken(user *model.User) (string, error)
}

type jwtService struct {
	secret     string
	expiryTime time.Duration
}

func NewJWTService(secret string, expiry time.Duration) TokenService {
	return &jwtService{secret: secret, expiryTime: expiry}
}

func (s *jwtService) GenerateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(s.expiryTime).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}
