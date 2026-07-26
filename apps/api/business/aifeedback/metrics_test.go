package aifeedback

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoopMetricsRecorderDoesNotStoreEvents(t *testing.T) {
	rec := NewNoopMetricsRecorder()
	rec.RecordFeedback(context.Background(), MetricsEvent{Outcome: "test"})
	rec.RecordReport(context.Background(), MetricsReportEvent{Classification: "test"})
	// No panic and no observable state.
	assert.NotNil(t, rec)
}

func TestInMemoryMetricsRecorderStoresFeedbackEvents(t *testing.T) {
	rec := NewInMemoryMetricsRecorder()
	event := MetricsEvent{
		PromptVersion:  PromptVersionSentenceFeedbackV1,
		SchemaVersion:  SchemaVersionFeedbackV1,
		Provider:       ProviderMock,
		Model:          "mock",
		Release:        "test-release",
		Outcome:        "success",
		DurationMs:     123,
		LearningStatus: LearningStatusCorrect,
		Count:          1,
	}
	rec.RecordFeedback(context.Background(), event)

	assert.Len(t, rec.FeedbackEvents(), 1)
	assert.Equal(t, event, rec.FeedbackEvents()[0])
}

func TestInMemoryMetricsRecorderStoresReportEvents(t *testing.T) {
	rec := NewInMemoryMetricsRecorder()
	event := MetricsReportEvent{
		Classification: "incorrect",
		Provider:       ProviderMock,
		Model:          "mock",
		Release:        "test-release",
		Count:          1,
	}
	rec.RecordReport(context.Background(), event)

	assert.Len(t, rec.ReportEvents(), 1)
	assert.Equal(t, event, rec.ReportEvents()[0])
}

func TestMetricsEventNeverContainsLearnerText(t *testing.T) {
	event := MetricsEvent{
		Outcome:    "validation_failed",
		DurationMs: 10,
		Count:      1,
	}
	assert.Empty(t, event.LearningStatus)
	assert.Empty(t, event.PromptVersion)
	assert.Empty(t, event.SchemaVersion)
	assert.Empty(t, event.Provider)
	assert.Empty(t, event.Model)
	assert.Empty(t, event.Release)
	// UserID and sentence fields are intentionally absent from the struct.
}
