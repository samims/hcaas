package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/samims/hcaas/services/url/internal/model"
)

type PostgresStorage struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) HealthCheckStorage {
	return &PostgresStorage{pool}
}

func (ps *PostgresStorage) Ping(ctx context.Context) error {
	return ps.db.Ping(ctx)
}

func (ps *PostgresStorage) FindByID(id string) (model.URL, error) {
	ctx := context.Background()

	const query = `
		SELECT id, address, status, checked_at
		FROM urls
		WHERE id = $1
	`

	var url model.URL
	err := ps.db.QueryRow(ctx, query, id).Scan(
		&url.ID, &url.Address, &url.Status, &url.CheckedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.URL{}, fmt.Errorf("url not found: %w", err)
		}
		return model.URL{}, fmt.Errorf("find by id failed: %w", err)
	}

	return url, nil
}

func (ps *PostgresStorage) FindAll() ([]model.URL, error) {
	ctx := context.Background()

	const query = `
		SELECT id, address, status, checked_at
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
		if err := rows.Scan(&url.ID, &url.Address, &url.Status, &url.CheckedAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return urls, nil
}

func (ps *PostgresStorage) Save(url model.URL) error {
	ctx := context.Background()

	const queryStr = `
		INSERT INTO urls(id, address, status, checked_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := ps.db.Exec(ctx, queryStr, url.ID, url.Address, url.Status)

	if err != nil {
		return fmt.Errorf("failed to save url: %w", err)
	}

	return nil
}

func (ps *PostgresStorage) UpdateStatus(id string, status string, checkedAt time.Time) error {
	ctx := context.Background()
	const query = `
		UPDATE urls
		SET status = $1, checked_at $2
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
