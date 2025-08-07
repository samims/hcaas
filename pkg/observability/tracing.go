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
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
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
// This function uses the recommended `grpc.NewClient` for a non-blocking
// connection to the OpenTelemetry collector.
func NewTracerProvider(
	ctx context.Context,
	serviceName string,
	collectorEndpoint string,
	logger *slog.Logger,
) (*TracerProvider, func(), error) {
	logger.Info("Initializing OpenTelemetry Tracer", "service", serviceName, "collector", collectorEndpoint)

	// Create a gRPC client connection to the OpenTelemetry collector.
	// The first argument is the string target address, followed by options.
	conn, err := grpc.NewClient(
		collectorEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("Failed to create gRPC connection to collector", slog.Any("error", err))
		return nil, nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// Create an OTLP exporter over the gRPC connection we just created.
	// This function correctly takes a context as its first argument.
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		logger.Error("Failed to create OTLP trace exporter", slog.Any("error", err))
		conn.Close() // Close the connection if the exporter creation fails
		return nil, nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Set resource attributes (service name, environment, etc.)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			// Set a unique identifier for this service instance using the container's hostname.
			// Useful for distinguishing between different instances in observability tools.
			semconv.ServiceInstanceID(os.Getenv("HOSTNAME")),
		),
	)
	if err != nil {
		logger.Error("Failed to create OpenTelemetry resource", slog.Any("error", err))
		conn.Close() // Close the connection if resource creation fails
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider with a BatchSpanProcessor, which is the recommended
	// way to process spans for production environments.
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
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
