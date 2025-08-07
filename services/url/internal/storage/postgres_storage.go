package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	appErr "github.com/samims/hcaas/services/url/internal/errors"
	"github.com/samims/hcaas/services/url/internal/model"
)

type Storage interface {
	Ping(ctx context.Context) error
	Save(url *model.URL) error
	FindAll() ([]model.URL, error)
	FindAllByUserID(ctx context.Context, userID string) ([]model.URL, error)
	FindByID(id string) (model.URL, error)
	FindByAddress(address string) (model.URL, error)
	UpdateStatus(id, status string, checkedAt time.Time) error
}

type postgresStorage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) Storage {
	return &postgresStorage{pool}
}

func (ps *postgresStorage) Ping(ctx context.Context) error {
	return ps.db.Ping(ctx)
}

func (ps *postgresStorage) FindByID(id string) (model.URL, error) {
	ctx := context.Background()

	const query = `
		SELECT id, user_id, address, status, checked_at, updated_at, created_at
		FROM urls
		WHERE id = $1
	`

	var url model.URL
	err := ps.db.QueryRow(ctx, query, id).Scan(
		&url.ID, &url.UserID, &url.Address, &url.Status, &url.CheckedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.URL{}, fmt.Errorf("url not found: %w", err)
		}
		return model.URL{}, fmt.Errorf("find by id failed: %w", err)
	}

	return url, nil
}

func (ps *postgresStorage) FindAll() ([]model.URL, error) {
	ctx := context.Background()

	const query = `
		SELECT id, user_id, address, status, checked_at
		FROM urls
	`

	rows, err := ps.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var urls []model.URL

	for rows.Next() {
		var url model.URL
		if err := rows.Scan(&url.ID, &url.UserID, &url.Address, &url.Status, &url.CheckedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return urls, nil
}

func (ps *postgresStorage) FindAllByUserID(ctx context.Context, userID string) ([]model.URL, error) {
	const query = `
		SELECT id, user_id, address, status, checked_at
		from urls
		where user_id = $1
	`
	rows, err := ps.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query failed %w", err)
	}
	defer rows.Close()

	var urls []model.URL

	for rows.Next() {
		var url model.URL
		if err := rows.Scan(&url.ID, &url.UserID, &url.Address, &url.Status, &url.CheckedAt); err != nil {
			return nil, fmt.Errorf("scan failed %w", err)
		}
		urls = append(urls, url)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed %w", err)

	}
	return urls, nil

}
func (ps *postgresStorage) Save(url *model.URL) error {
	ctx := context.Background()

	const queryStr = `
		INSERT INTO urls(id, user_id, address, status, checked_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := ps.db.QueryRow(ctx, queryStr, url.ID, url.UserID, url.Address, url.Status, url.CheckedAt).Scan(&url.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return appErr.ErrConflict
			}
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}

	return nil
}

func (ps *postgresStorage) UpdateStatus(id string, status string, checkedAt time.Time) error {
	ctx := context.Background()
	const query = `
		UPDATE urls
		SET status = $1, checked_at = $2
		WHERE id = $3
	`

	cmdTags, err := ps.db.Exec(ctx, query, status, checkedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	if cmdTags.RowsAffected() == 0 {
		return fmt.Errorf("no record found to update with id %s", id)
	}
	return nil
}

func (ps *postgresStorage) FindByAddress(address string) (model.URL, error) {
	ctx := context.Background()

	const query = `
		SELECT id, user_id, address, status, checked_at
		FROM urls
		WHERE address = $1
	`

	var url model.URL
	err := ps.db.QueryRow(ctx, query, address).Scan(
		&url.ID, &url.UserID, &url.Address, &url.Status, &url.CheckedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.URL{}, appErr.ErrNotFound
		}
		return model.URL{}, fmt.Errorf("find by address failed: %w", err)
	}

	return url, nil
}
