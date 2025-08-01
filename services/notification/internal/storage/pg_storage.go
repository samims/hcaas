package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/samims/hcaas/services/notification/internal/config"
	"github.com/samims/hcaas/services/notification/internal/model"
)

type PostgresStorage struct {
	db *sqlx.DB
}

// NewPostgresStorage creates connection pool
func NewPostgresStorage(cfg config.DBConfig) (*PostgresStorage, error) {
	db, err := sqlx.Connect("postgres", cfg.URL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.MaxOpenConn)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdle)
	return &PostgresStorage{db: db}, nil
}

// Save inserts a new notification with status pending
func (s *PostgresStorage) Save(ctx context.Context, notif *model.Notification) error {
	query := `INSERT INTO notifications
		(url_id, type, message, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`

	row := s.db.QueryRowxContext(ctx, query, notif.UrlId, notif.Type, notif.Message, notif.Status)
	if err := row.Scan(&notif.ID, &notif.CreatedAt, &notif.UpdatedAt); err != nil {
		return err
	}
	return nil
}

// GetPending returns notifications with status pending
func (s *PostgresStorage) GetPending(ctx context.Context) ([]model.Notification, error) {
	var notifs []model.Notification
	query := `SELECT * FROM notifications where staus = $1`
	err := s.db.SelectContext(ctx, &notifs, query, model.StatusPending)
	if err != nil {
		return nil, err
	}
	return notifs, nil
}

// UpdateStatus updates notification status and timestamp
func (s *PostgresStorage) UpdateStatus(ctx context.Context, id int, status string, updatedAt time.Time) error {
	query := "UPDATE notifications SET status=$1, updated_at=$2 where id=$3"
	res, err := s.db.ExecContext(ctx, query, status, updatedAt, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("no rows updated")
	}
	return nil
}
