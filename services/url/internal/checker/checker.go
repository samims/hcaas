package checker

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/samims/hcaas/services/url/internal/metrics"
	"github.com/samims/hcaas/services/url/internal/model"
	"github.com/samims/hcaas/services/url/internal/service"
)

const (
	Healthy   = "healthy"
	UnHealthy = "unhealthy"
)

type URLChecker struct {
	svc        service.URLService
	logger     *slog.Logger
	httpClient *http.Client
	interval   time.Duration
}

func NewURLChecker(svc service.URLService, logger *slog.Logger, client *http.Client, interval time.Duration) *URLChecker {
	return &URLChecker{
		svc:        svc,
		logger:     logger,
		httpClient: client,
		interval:   interval,
	}
}

func (uc *URLChecker) Start(ctx context.Context) {
	uc.logger.Info("URLChecker started")

	ticker := time.NewTicker(uc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			uc.logger.Info("URLChecker stopped")
			return
		case <-ticker.C:
			uc.CheckAllURLs(ctx)
		}
	}
}

func (uc *URLChecker) CheckAllURLs(ctx context.Context) {
	urls, err := uc.svc.GetAll(ctx)
	if err != nil {
		uc.logger.Error("Failed to fetch URLs", slog.Any("error", err))
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit to 10 concurrent checks

	for _, url := range urls {
		wg.Add(1)
		go func(url model.URL) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			uc.logger.Info("Checking URL", slog.String("id", url.ID), slog.String("address", url.Address))

			status := uc.ping(ctx, url.Address)

			err := uc.svc.UpdateStatus(ctx, url.ID, status)
			if err != nil {
				uc.logger.Error("Failed to update URL status",
					slog.String("urlID", url.ID),
					slog.String("status", status),
					slog.Any("error", err),
				)
			} else {
				uc.logger.Info("URL status updated",
					slog.String("urlID", url.ID),
					slog.String("address", url.Address),
					slog.String("status", status),
				)
			}
		}(url)
	}

	wg.Wait()
}

// ping performs a HTTP GET with timeout and metrics
func (uc *URLChecker) ping(parentCtx context.Context, target string) string {
	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		uc.logger.Warn("Failed to create HTTP request", slog.String("address", target), slog.Any("error", err))
		metrics.URLCheckStatus.WithLabelValues(model.StatusDown).Inc()
		return UnHealthy
	}

	start := time.Now()
	resp, err := uc.httpClient.Do(req)
	duration := time.Since(start).Seconds()

	if err != nil {
		uc.logger.Warn("HTTP request failed", slog.String("address", target), slog.Any("error", err))
		metrics.URLCheckStatus.WithLabelValues(model.StatusDown).Inc()
		metrics.URLCheckDuration.WithLabelValues(model.StatusDown).Observe(duration)
		return UnHealthy
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		uc.logger.Warn("Unhealthy HTTP status code",
			slog.String("address", target),
			slog.Int("statusCode", resp.StatusCode),
		)
		metrics.URLCheckStatus.WithLabelValues(model.StatusDown).Inc()
		metrics.URLCheckDuration.WithLabelValues(model.StatusDown).Observe(duration)
		return UnHealthy
	}

	metrics.URLCheckStatus.WithLabelValues(model.StatusUP).Inc()
	metrics.URLCheckDuration.WithLabelValues(model.StatusUP).Observe(duration)
	return Healthy
}
