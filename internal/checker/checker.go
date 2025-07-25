package checker

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/samims/hcaas/internal/metrics"
	"github.com/samims/hcaas/internal/service"
)

const Healthy = "healthy"
const UnHealthy = "unhealthy"

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
			uc.CheckAllURLs()
		}
	}
}

func (uc *URLChecker) CheckAllURLs() {
	ctx := context.Background()
	urls, err := uc.svc.GetAll(ctx)
	if err != nil {
		uc.logger.Error("Failed to fetch URLs", slog.Any("error", err))
		return
	}

	for _, url := range urls {
		status := uc.ping(url.Address)

		err := uc.svc.UpdateStatus(ctx, url.ID, status)
		if err != nil {
			uc.logger.Error("Failed to update URL status",
				slog.String("urlID", url.ID),
				slog.String("status", status),
				slog.Any("error", err),
			)
			continue
		}
		uc.logger.Info("URL status updated",
			slog.String("urlID", url.ID),
			slog.String("address", url.Address),
			slog.String("status", status),
		)
	}
}

// ping is a simple HTTP GET to check health
func (uc *URLChecker) ping(target string) string {
	resp, err := uc.httpClient.Get(target)
	if err != nil || resp.StatusCode >= http.StatusBadRequest {
		uc.logger.Warn("Invalid URL", slog.String("address", target), slog.Any("error", err))
		metrics.URLCheckStatus.WithLabelValues("up").Inc()
		return UnHealthy
	}
	metrics.URLCheckStatus.WithLabelValues("up").Inc()
	return Healthy
}
