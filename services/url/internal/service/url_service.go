package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	appErr "github.com/samims/hcaas/services/url/internal/errors"
	"github.com/samims/hcaas/services/url/internal/model"
	"github.com/samims/hcaas/services/url/internal/storage"
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
	tracer trace.Tracer
}

func NewURLService(store storage.HealthCheckStorage, logger *slog.Logger) URLService {
	l := logger.With("layer", "service", "component", "urlService")
	return &urlService{
		store:  store,
		logger: l,
		tracer: otel.Tracer("url-service"),
	}
}

func (s *urlService) GetAll(ctx context.Context) ([]model.URL, error) {
	ctx, span := s.tracer.Start(ctx, "GetAll")
	defer span.End()

	s.logger.Info("GetAll called")

	urls, err := s.store.FindAll()
	if err != nil {
		s.logger.Error("failed to fetch URLs", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, appErr.NewInternal("failed to fetch URLs: %v", err)
	}

	span.SetAttributes(attribute.Int("url.count", len(urls)))
	s.logger.Info("GetAll succeeded", slog.Int("count", len(urls)))
	return urls, nil
}

func (s *urlService) GetByID(ctx context.Context, id string) (*model.URL, error) {
	ctx, span := s.tracer.Start(ctx, "GetByID")
	defer span.End()

	span.SetAttributes(attribute.String("url.id", id))
	s.logger.Info("GetByID called", slog.String("id", id))

	url, err := s.store.FindByID(id)
	if err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			s.logger.Warn("URL not found", slog.String("id", id))
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, appErr.NewNotFound(fmt.Sprintf("URL with ID %s not found", id))
		}
		s.logger.Error("failed to fetch URL by ID", slog.String("id", id))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, appErr.NewInternal("failed to fetch URL by ID: %v", err)
	}

	s.logger.Info("GetByID succeeded", slog.String("id", id))
	return &url, nil
}

func (s *urlService) Add(ctx context.Context, url model.URL) error {
	ctx, span := s.tracer.Start(ctx, "Add")
	defer span.End()

	span.SetAttributes(
		attribute.String("url.address", url.Address),
		attribute.String("url.id", url.ID),
	)
	s.logger.Info("Add url called", slog.String("url", url.Address))

	if url.ID == "" {
		url.ID = uuid.New().String()
	}
	if err := s.store.Save(&url); err != nil {
		if errors.Is(err, appErr.ErrConflict) {
			s.logger.Warn("URL already exists", slog.String("URL", url.Address))
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return appErr.NewConflict("URL with ID %s already exists", url.ID)
		}
		s.logger.Error("failed to add URL", slog.String("id", url.ID), slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return appErr.NewInternal("failed to add URL: %v", err)
	}

	span.SetAttributes(attribute.String("url.id", url.ID))
	s.logger.Info("Add succeeded", slog.String("id", url.ID))
	return nil
}
func (s *urlService) UpdateStatus(ctx context.Context, id string, status string) error {
	ctx, span := s.tracer.Start(ctx, "UpdateStatus")
	defer span.End()

	span.SetAttributes(
		attribute.String("url.id", id),
		attribute.String("url.status", status),
	)
	s.logger.Info("UpdateStatus called", slog.String("id", id), slog.String("status", status))

	if err := s.store.UpdateStatus(id, status, time.Now()); err != nil {
		if errors.Is(err, appErr.ErrNotFound) {
			s.logger.Warn("URL not found for update", slog.String("id", id))
			err := appErr.NewNotFound(fmt.Sprintf("cannot update: URL with ID %s not found", id))
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		s.logger.Error("failed to update status", slog.String("id", id), slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return appErr.NewInternal("failed to update URL status: %v", err)
	}

	s.logger.Info("UpdateStatus succeeded", slog.String("id", id), slog.String("status", status))
	return nil
}
