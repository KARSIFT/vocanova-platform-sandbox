package aifeedback

import (
	"context"
	"sync"
)

// MetricsEvent is privacy-safe telemetry: it groups feedback outcomes by prompt
// version, schema version, provider, model, and release. It never contains
// learner text, user identifiers, or provider credentials.
type MetricsEvent struct {
	PromptVersion  string
	SchemaVersion  string
	Provider       string
	Model          string
	Release        string
	Outcome        string
	DurationMs     int64
	LearningStatus string
	Count          int
}

// MetricsReportEvent is a privacy-safe report metric. It contains the report
// classification and the provider/model/release metadata only.
type MetricsReportEvent struct {
	Classification string
	Provider       string
	Model          string
	Release        string
	Count          int
}

// MetricsRecorder records privacy-safe feedback metrics. Implementations must
// never include learner text in metric labels or dimensions.
type MetricsRecorder interface {
	RecordFeedback(ctx context.Context, event MetricsEvent)
	RecordReport(ctx context.Context, event MetricsReportEvent)
}

// NoopMetricsRecorder discards metrics.
type NoopMetricsRecorder struct{}

// NewNoopMetricsRecorder creates a no-op metrics recorder.
func NewNoopMetricsRecorder() *NoopMetricsRecorder {
	return &NoopMetricsRecorder{}
}

// RecordFeedback implements MetricsRecorder with no side effects.
func (n *NoopMetricsRecorder) RecordFeedback(ctx context.Context, event MetricsEvent) {}

// RecordReport implements MetricsRecorder with no side effects.
func (n *NoopMetricsRecorder) RecordReport(ctx context.Context, event MetricsReportEvent) {}

// InMemoryMetricsRecorder is a test double that stores metrics in memory so
// tests can assert that privacy-safe dimensions are recorded.
type InMemoryMetricsRecorder struct {
	mu       sync.Mutex
	feedback []MetricsEvent
	reports  []MetricsReportEvent
}

// NewInMemoryMetricsRecorder creates an in-memory metrics recorder.
func NewInMemoryMetricsRecorder() *InMemoryMetricsRecorder {
	return &InMemoryMetricsRecorder{
		feedback: make([]MetricsEvent, 0),
		reports:  make([]MetricsReportEvent, 0),
	}
}

// RecordFeedback stores the event in memory.
func (m *InMemoryMetricsRecorder) RecordFeedback(ctx context.Context, event MetricsEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.feedback = append(m.feedback, event)
}

// RecordReport stores the event in memory.
func (m *InMemoryMetricsRecorder) RecordReport(ctx context.Context, event MetricsReportEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reports = append(m.reports, event)
}

// FeedbackEvents returns a copy of recorded feedback events.
func (m *InMemoryMetricsRecorder) FeedbackEvents() []MetricsEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]MetricsEvent, len(m.feedback))
	copy(out, m.feedback)
	return out
}

// ReportEvents returns a copy of recorded report events.
func (m *InMemoryMetricsRecorder) ReportEvents() []MetricsReportEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]MetricsReportEvent, len(m.reports))
	copy(out, m.reports)
	return out
}

// compile-time interface check.
var _ MetricsRecorder = (*NoopMetricsRecorder)(nil)
var _ MetricsRecorder = (*InMemoryMetricsRecorder)(nil)
