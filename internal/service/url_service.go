package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	appErr "github.com/samims/hcaas/internal/errors"
	"github.com/samims/hcaas/internal/model"
	"github.com/samims/hcaas/internal/storage"
)

type URLService interface {
	GetAll(ctx context.Context) ([]model.URL, error)
	GetByID(ctx context.Context, id string) (*model.URL, error)
	Add(ctx context.Context, url model.URL) error
	UpdateStatus(ctx context.Context, id string, status string) error
}

type urlService struct {
	store storage.HealthCheckStorage
}

func NewURLService(store storage.HealthCheckStorage) URLService {
	return &urlService{store: store}
}

func (s *urlService) GetAll(ctx context.Context) ([]model.URL, error) {
	urls, err := s.store.FindAll()
	if err != nil {
		return nil, appErr.NewInternal("failed to fetch URLs: %v", err)
	}
	return urls, nil
}

func (s *urlService) GetByID(ctx context.Context, id string) (*model.URL, error) {
	url, err := s.store.FindByID(id)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return nil, appErr.NewNotFound(fmt.Sprintf("URL with ID %s not found", id))
		}
		return nil, appErr.NewInternal("failed to fetch URL by ID: %v", err)
	}
	return &url, nil
}

func (s *urlService) Add(ctx context.Context, url model.URL) error {
	if err := s.store.Save(url); err != nil {
		if errors.Is(err, appErr.ErrConflict) {
			return appErr.NewInternal("URL with ID %s already exists", url.ID)
		}
		return appErr.NewInternal("failed to add URL: %v", err)
	}
	return nil
}

func (s *urlService) UpdateStatus(ctx context.Context, id string, status string) error {
	if err := s.store.UpdateStatus(id, status, time.Now()); err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			return appErr.NewNotFound(fmt.Sprintf("cannot update: URL with ID %s not found", id))
		}
		return appErr.NewInternal("failed to update URL status: %v", err)
	}
	return nil
}
