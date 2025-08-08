package tracing

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer provides Google Cloud compliant tracing
type Tracer struct {
	tracer trace.Tracer
	logger *slog.Logger
}

// NewTracer creates a new tracer instance
func NewTracer(tracer trace.Tracer, logger *slog.Logger) *Tracer {
	return &Tracer{
		tracer: tracer,
		logger: logger,
	}
}

// StartSpan creates a new span with Google Cloud attributes
func (t *Tracer) StartSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := t.tracer.Start(ctx, operation,
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	return ctx, span
}

// StartClientSpan creates a new client span
func (t *Tracer) StartClientSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	ctx, span := t.tracer.Start(ctx, operation,
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindClient),
	)
	return ctx, span
}

// RecordError records an error on the span
func (t *Tracer) RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// AddAttributes adds attributes to span
func (t *Tracer) AddAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// AddGoogleCloudAttributes adds Google Cloud specific attributes
func (t *Tracer) AddGoogleCloudAttributes(span trace.Span, projectID, region, zone string) {
	span.SetAttributes(
		attribute.String("gcp.project_id", projectID),
		attribute.String("gcp.region", region),
		attribute.String("gcp.zone", zone),
	)
}

// AddServiceAttributes adds service-specific attributes
func (t *Tracer) AddServiceAttributes(span trace.Span, serviceName, serviceVersion, environment string) {
	span.SetAttributes(
		attribute.String("service.name", serviceName),
		attribute.String("service.version", serviceVersion),
		attribute.String("service.environment", environment),
	)
}

// AddRequestAttributes adds HTTP request attributes
func (t *Tracer) AddRequestAttributes(span trace.Span, method, path, userAgent string, statusCode int) {
	span.SetAttributes(
		attribute.String("http.method", method),
		attribute.String("http.route", path),
		attribute.String("http.user_agent", userAgent),
		attribute.Int("http.status_code", statusCode),
	)
}

// AddDatabaseAttributes adds database operation attributes
func (t *Tracer) AddDatabaseAttributes(span trace.Span, operation, table string, duration time.Duration) {
	span.SetAttributes(
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
		attribute.Float64("db.duration_ms", float64(duration.Milliseconds())),
	)
}

// AddKafkaAttributes adds Kafka operation attributes
func (t *Tracer) AddKafkaAttributes(span trace.Span, topic, operation string, partition int32, offset int64) {
	span.SetAttributes(
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination", topic),
		attribute.String("messaging.operation", operation),
		attribute.Int64("messaging.kafka.partition", int64(partition)),
		attribute.Int64("messaging.kafka.offset", offset),
	)
}

// GetTracer returns the global tracer
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
