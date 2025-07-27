package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/samims/hcaas/services/auth/internal/model"

	"github.com/google/uuid"
)

type UserStorage interface {
	CreateUser(ctx context.Context, email, hashedPass string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

type userStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) UserStorage {
	return &userStorage{db: db}
}

func (s *userStorage) CreateUser(ctx context.Context, email, hashedPass string) (*model.User, error) {
	id := uuid.New().String()
	now := time.Now()
	query := `
		INSERT INTO users (id, email, password, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.db.ExecContext(ctx, query, id, email, hashedPass)

	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        id,
		Email:     email,
		Password:  hashedPass,
		CreatedAt: now,
	}, nil
}

func (s *userStorage) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password, created_at
		FROM users
		WHERE email = $1
	`
	row := s.db.QueryRowContext(ctx, query, email)

	var user model.User
	if err := row.Scan(&user.ID, &user.Email, user.Password, user.CreatedAt); err != nil {
		return nil, err
	}

	return &user, nil
}
