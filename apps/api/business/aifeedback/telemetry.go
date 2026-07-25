package aifeedback

import (
	"context"

	"github.com/google/uuid"
)

// FeedbackEvent is privacy-safe telemetry data: it never includes learner text
// in metric labels. It groups by prompt/schema version, provider, model, and
// outcome only.
type FeedbackEvent struct {
	UserID         uuid.UUID
	PromptVersion  string
	SchemaVersion  string
	Provider       string
	Model          string
	Outcome        string
	DurationMs     int64
	LearningStatus string
}

// FeedbackReport is a privacy-safe report of a quality concern. It never includes
// learner text; only attempt identifiers, classifications, and a short reason.
type FeedbackReport struct {
	UserID         uuid.UUID
	AttemptID      uuid.UUID
	Reason         string
	Classification string
}

// TelemetryRecorder records privacy-safe feedback events and reports.
type TelemetryRecorder interface {
	Record(ctx context.Context, event FeedbackEvent)
	RecordReport(ctx context.Context, report FeedbackReport)
}

// NoopTelemetryRecorder discards telemetry events. It is used for tests and as
// a default before the observability layer is wired.
type NoopTelemetryRecorder struct{}

// NewNoopTelemetryRecorder creates a no-op telemetry recorder.
func NewNoopTelemetryRecorder() *NoopTelemetryRecorder {
	return &NoopTelemetryRecorder{}
}

// Record implements TelemetryRecorder with no side effects.
func (n *NoopTelemetryRecorder) Record(ctx context.Context, event FeedbackEvent) {}

// RecordReport implements TelemetryRecorder with no side effects.
func (n *NoopTelemetryRecorder) RecordReport(ctx context.Context, report FeedbackReport) {}
