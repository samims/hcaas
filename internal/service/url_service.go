package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	store  storage.HealthCheckStorage
	logger *slog.Logger
}

func NewURLService(store storage.HealthCheckStorage, logger *slog.Logger) URLService {
	l := logger.With("layer", "service", "component", "urlService")
	return &urlService{store: store, logger: l}
}

func (s *urlService) GetAll(_ context.Context) ([]model.URL, error) {
	s.logger.Info("GetAll called")
	urls, err := s.store.FindAll()
	if err != nil {
		s.logger.Error("failed to fetch URLs", slog.String("error", err.Error()))
		return nil, appErr.NewInternal("failed to fetch URLs: %v", err)
	}
	s.logger.Info("GetAll succeeded", slog.Int("count", len(urls)))
	return urls, nil
}

func (s *urlService) GetByID(_ context.Context, id string) (*model.URL, error) {
	s.logger.Info("GetByID called", slog.String("id", id))
	url, err := s.store.FindByID(id)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			s.logger.Warn("URL not found", slog.String("id", id))
			return nil, appErr.NewNotFound(fmt.Sprintf("URL with ID %s not found", id))
		}
		s.logger.Error("failed to fetch URL by ID", slog.String("id", id))
		return nil, appErr.NewInternal("failed to fetch URL by ID: %v", err)
	}

	s.logger.Info("GetByID succeeded", slog.String("id", id))
	return &url, nil
}

func (s *urlService) Add(_ context.Context, url model.URL) error {
	s.logger.Info("Add called")
	if err := s.store.Save(url); err != nil {
		if errors.Is(err, appErr.ErrConflict) {
			s.logger.Warn("URL already exists", slog.String("id", url.ID))
			return appErr.NewInternal("URL with ID %s already exists", url.ID)
		}
		s.logger.Error("failed to add URL", slog.String("id", url.ID), slog.String("error", err.Error()))

		return appErr.NewInternal("failed to add URL: %v", err)
	}
	s.logger.Info("GetByID succeeded", slog.String("id", url.ID))

	return nil
}

func (s *urlService) UpdateStatus(_ context.Context, id string, status string) error {
	s.logger.Info("UpdateStatus called", slog.String("id", id), slog.String("status", status))
	if err := s.store.UpdateStatus(id, status, time.Now()); err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			s.logger.Warn("URL not found for update", slog.String("id", id))
			return appErr.NewNotFound(fmt.Sprintf("cannot update: URL with ID %s not found", id))
		}
		s.logger.Error("failed to update status", slog.String("id", id), slog.String("error", err.Error()))
		return appErr.NewInternal("failed to update URL status: %v", err)
	}
	s.logger.Info("UpdateStatus succeeded", slog.String("id", id), slog.String("status", status))
	return nil
}
