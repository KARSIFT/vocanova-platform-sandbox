package aifeedback

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	ProcessingStatusPending   = "pending"
	ProcessingStatusCompleted = "completed"
	ProcessingStatusFailed    = "failed"
	ProcessingStatusSkipped   = "skipped"
)

var ErrInvalidCursor = errors.New("invalid cursor")

// LearnerSentence is the owner-scoped public projection of a submission and
// its latest feedback attempt. Provider names, prompts, costs, and internal
// failure details deliberately do not cross this boundary.
type LearnerSentence struct {
	ID                      uuid.UUID
	FeedbackID              uuid.UUID
	TargetWordID            uuid.UUID
	ProcessingStatus        string
	Status                  string
	OriginalSentence        string
	CorrectedSentence       *string
	Headline                string
	Explanation             string
	ImprovementTip          *string
	TargetWordUsedCorrectly bool
	GrammarAcceptable       bool
	MeaningClear            bool
	Naturalness             string
	Reported                bool
	CreatedAt               time.Time
}

type ListLearnerSentencesRequest struct {
	UserID      uuid.UUID
	AfterCursor string
	Limit       int
}

type ListLearnerSentencesResponse struct {
	Items      []LearnerSentence
	NextCursor string
}

type learnerSentenceCursor struct {
	SubmittedAt time.Time `json:"submittedAt"`
	ID          uuid.UUID `json:"id"`
}

func encodeLearnerSentenceCursor(cursor learnerSentenceCursor) string {
	data, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeLearnerSentenceCursor(value string) (learnerSentenceCursor, error) {
	var cursor learnerSentenceCursor
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return cursor, ErrInvalidCursor
	}
	if err := json.Unmarshal(data, &cursor); err != nil || cursor.SubmittedAt.IsZero() || cursor.ID == uuid.Nil {
		return learnerSentenceCursor{}, ErrInvalidCursor
	}
	return cursor, nil
}

func processingStatus(sentenceStatus, attemptStatus string) string {
	switch {
	case sentenceStatus == SentenceStatusArchived || attemptStatus == AttemptStatusCancelled:
		return ProcessingStatusSkipped
	case sentenceStatus == SentenceStatusFeedbackReady || attemptStatus == AttemptStatusSucceeded:
		return ProcessingStatusCompleted
	case sentenceStatus == SentenceStatusFeedbackFailed || attemptStatus == AttemptStatusFailed:
		return ProcessingStatusFailed
	default:
		return ProcessingStatusPending
	}
}

func learnerSentenceFromStored(
	sentenceID, feedbackID, targetWordID uuid.UUID,
	sentenceStatus, attemptStatus, original string,
	feedback map[string]any,
	feedbackText string,
	reported bool,
	createdAt time.Time,
) LearnerSentence {
	item := LearnerSentence{
		ID:               sentenceID,
		FeedbackID:       feedbackID,
		TargetWordID:     targetWordID,
		ProcessingStatus: processingStatus(sentenceStatus, attemptStatus),
		OriginalSentence: original,
		Reported:         reported,
		CreatedAt:        createdAt,
	}
	if item.ProcessingStatus != ProcessingStatusCompleted {
		return item
	}
	projection := completedFeedbackFromStored(feedback, feedbackText)
	item.Status = projection.Status
	item.Headline = projection.Headline
	item.Explanation = projection.Explanation
	item.Naturalness = projection.Naturalness
	item.TargetWordUsedCorrectly = projection.TargetWordUsedCorrectly
	item.GrammarAcceptable = projection.GrammarAcceptable
	item.MeaningClear = projection.MeaningClear
	item.CorrectedSentence = projection.CorrectedSentence
	item.ImprovementTip = projection.ImprovementTip
	return item
}

// completedFeedbackFromStored also understands rows written before the public
// diagnostic fields were added. Those records retain the outcome and human
// explanation, so the documented classification rules provide conservative
// defaults instead of returning an invalid completed response.
func completedFeedbackFromStored(feedback map[string]any, explanationFallback string) ProviderFeedback {
	projection := ProviderFeedback{
		Status:                  stringValue(feedback, "status"),
		Headline:                stringValue(feedback, "headline"),
		Explanation:             stringValue(feedback, "explanation"),
		Naturalness:             stringValue(feedback, "naturalness"),
		TargetWordUsedCorrectly: boolValue(feedback, "target_word_used_correctly"),
		GrammarAcceptable:       boolValue(feedback, "grammar_acceptable"),
		MeaningClear:            boolValue(feedback, "meaning_clear"),
	}
	if projection.Explanation == "" {
		projection.Explanation = explanationFallback
	}
	if value, ok := feedback["corrected_sentence"].(string); ok && value != "" {
		projection.CorrectedSentence = &value
	}
	if value, ok := feedback["improvement_tip"].(string); ok && value != "" {
		projection.ImprovementTip = &value
	}

	_, hasTargetDiagnostic := feedback["target_word_used_correctly"]
	_, hasGrammarDiagnostic := feedback["grammar_acceptable"]
	_, hasMeaningDiagnostic := feedback["meaning_clear"]
	switch projection.Status {
	case LearningStatusCorrect:
		if !hasTargetDiagnostic {
			projection.TargetWordUsedCorrectly = true
		}
		if !hasGrammarDiagnostic {
			projection.GrammarAcceptable = true
		}
		if !hasMeaningDiagnostic {
			projection.MeaningClear = true
		}
		if projection.Naturalness == "" {
			projection.Naturalness = NaturalnessNatural
		}
	case LearningStatusNeedsImprovement:
		if !hasMeaningDiagnostic {
			projection.MeaningClear = true
		}
		if projection.Naturalness == "" {
			projection.Naturalness = NaturalnessUnderstandable
		}
	case LearningStatusIncorrect:
		if projection.Naturalness == "" {
			projection.Naturalness = NaturalnessUnnatural
		}
	}
	return projection
}

func boolValue(values map[string]any, key string) bool {
	value, _ := values[key].(bool)
	return value
}
