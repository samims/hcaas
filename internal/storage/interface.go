package storage

import (
	"time"

	"github.com/samims/hcaas/internal/model"
)

type HealthCheckStorage interface {
	Save(url model.URL) error
	FindAll() ([]model.URL, error)
	FindByID(id string) (model.URL, error)
	UpdateStatus(id, status string, checkedAt time.Time) error
}
