// internal/observability/observability.go
package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0" // Using a consistent, up-to-date schema
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TracerProvider holds the configured OpenTelemetry TracerProvider.
// This struct makes the tracer a dependency that can be injected.
type TracerProvider struct {
	provider *trace.TracerProvider
	logger   *slog.Logger
}

// NewTracerProvider initializes and returns a new TracerProvider.
// The returned function should be called during application shutdown.
func NewTracerProvider(
	ctx context.Context,
	serviceName string,
	collectorEndpoint string,
	logger *slog.Logger,
) (*TracerProvider, func(), error) {
	logger.Info("Initializing OpenTelemetry Tracer", "service", serviceName, "collector", collectorEndpoint)

	// Create a gRPC client connection to the OpenTelemetry collector.
	conn, err := grpc.NewClient(
		collectorEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("Failed to create gRPC connection to collector", slog.Any("error", err))
		return nil, nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// Create an OTLP exporter over the gRPC connection.
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		logger.Error("Failed to create OTLP trace exporter", slog.Any("error", err))
		// The connection should be closed if the exporter creation fails.
		conn.Close()
		return nil, nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Create a resource that describes this application.
	// We use a single resource with explicit attributes to avoid schema conflicts.
	// The resource.NewWithAttributes function does not return an error.
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion("1.0.0"),
		// Set a unique identifier for this service instance using the container's hostname.
		semconv.ServiceInstanceID(os.Getenv("HOSTNAME")),
	)

	// Create a new trace provider with a BatchSpanProcessor, which is recommended
	// for production environments.
	bsp := trace.NewBatchSpanProcessor(exporter)
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)

	// Register the global tracer provider
	otel.SetTracerProvider(tp)

	logger.Info("TracerProvider initialized", slog.String("service", serviceName))

	cleanup := func() {
		logger.Info("Shutting down TracerProvider")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tp.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown TracerProvider", slog.Any("error", err))
		} else {
			logger.Info("TracerProvider shut down successfully")
		}

		// Ensure the gRPC connection is also closed during shutdown.
		if err := conn.Close(); err != nil {
			logger.Error("Failed to close gRPC connection", slog.Any("error", err))
		}
	}

	return &TracerProvider{provider: tp, logger: logger}, cleanup, nil
}

// Provider returns the underlying *trace.TracerProvider.
// This allows other components to access it for creating new tracers.
func (t *TracerProvider) Provider() *trace.TracerProvider {
	return t.provider
}
