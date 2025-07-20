package service

import (
	"context"
	"fmt"
	"time"

	"github.com/samims/hcaas/internal/model"
	"github.com/samims/hcaas/internal/storage"
)

type URLService interface {
	GetAllURLs(ctx context.Context) ([]model.URL, error)
	GetURLByID(ctx context.Context, id string) (model.URL, error)
	AddURL(ctx context.Context, url model.URL) error
	UpdateStatus(ctx context.Context, id string, status string) error
}

type DefaultURLService struct {
	storage storage.HealthCheckStorage
}

func NewURLService(s storage.HealthCheckStorage) URLService {
	return &DefaultURLService{storage: s}
}

func (s *DefaultURLService) GetAllURLs(ctx context.Context) ([]model.URL, error) {
	return s.storage.FindAll()
}

func (s *DefaultURLService) GetURLByID(ctx context.Context, id string) (model.URL, error) {
	return s.storage.FindByID(id)
}

func (s *DefaultURLService) AddURL(ctx context.Context, url model.URL) error {
	if url.ID == "" || url.Address == "" {
		return fmt.Errorf("invalid URL data")
	}
	return s.storage.Save(url)
}

func (s *DefaultURLService) UpdateStatus(ctx context.Context, id, status string) error {
	return s.storage.UpdateStatus(id, status, time.Now())
}
