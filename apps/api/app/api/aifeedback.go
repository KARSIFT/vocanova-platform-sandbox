package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// SentenceFeedbackResultDTO is the public response body for a sentence-feedback
// submission. It exposes only the fields the frontend is permitted to display.
type SentenceFeedbackResultDTO struct {
	SentenceID            string  `json:"sentenceId,omitempty" format:"uuid" doc:"Learner sentence identifier"`
	AttemptID             string  `json:"attemptId,omitempty" format:"uuid" doc:"AI feedback attempt identifier"`
	Status                string  `json:"status,omitempty" enum:"correct,needs_improvement,incorrect" doc:"Pedagogical outcome when feedback succeeded"`
	OriginalSentence      string  `json:"originalSentence" doc:"The sentence the learner submitted"`
	CorrectedSentence     *string `json:"correctedSentence,omitempty" doc:"Corrected sentence when useful"`
	Explanation           string  `json:"explanation,omitempty" doc:"Short explanation of the result"`
	ImprovementTip        *string `json:"improvementTip,omitempty" doc:"One improvement tip when useful"`
	MissionCompleted      bool    `json:"missionCompleted" doc:"Backend-confirmed mission state (P3 stub: always false)"`
	CanRetry              bool    `json:"canRetry" doc:"Whether the learner may safely retry"`
	Reported              bool    `json:"reported" doc:"Whether the learner has reported this feedback"`
	ErrorCode             string  `json:"errorCode,omitempty" doc:"Stable public error code on failure"`
	ErrorMessage          string  `json:"errorMessage,omitempty" doc:"Safe retryable message on failure"`
	CrisisResourceMessage string  `json:"crisisResourceMessage,omitempty" doc:"Non-clinical crisis resources for self-harm content"`
}

// SubmitSentenceFeedbackInput requests AI feedback for a learner sentence.
type SubmitSentenceFeedbackInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" doc:"User-scoped idempotency key"`
	Body           struct {
		SentenceText string `json:"sentenceText" maxLength:"300" required:"true" doc:"Learner sentence to check"`
		Source       string `json:"source" enum:"word_detail,review,daily_mission,free_practice" required:"true" doc:"Origin of the sentence"`
		AttemptID    string `json:"attemptId" format:"uuid" required:"true" doc:"Learner-owned attempt identifier (user_word_id for word_detail, review_attempt_id for review)"`
	}
}

// SubmitSentenceFeedbackOutput returns the sentence-feedback result.
type SubmitSentenceFeedbackOutput struct {
	Body SentenceFeedbackResultDTO
}

// ReportSentenceFeedbackInput reports a quality concern for a feedback attempt.
type ReportSentenceFeedbackInput struct {
	AttemptID string `path:"attemptId" format:"uuid" required:"true" doc:"AI feedback attempt identifier"`
	Body      struct {
		Reason         string `json:"reason" maxLength:"200" required:"true" doc:"Short reason for the report"`
		Classification string `json:"classification,omitempty" maxLength:"100" doc:"Optional classification (e.g. incorrect, unsafe, unclear)"`
	}
}

// ReportSentenceFeedbackOutput is an empty successful response.
type ReportSentenceFeedbackOutput struct{}

// RegisterAIFeedback registers the sentence-feedback write and report routes.
func RegisterAIFeedback(api huma.API, svc *aifeedback.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "SubmitSentenceFeedback",
		Method:      http.MethodPost,
		Path:        "/api/v1/sentence-feedback",
		Summary:     "Submit a learner sentence for AI feedback",
		Tags:        []string{"AI Feedback"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"200": {Description: "Feedback result, which may include a business-level error code"},
			"401": {Description: "Authentication is required"},
			"403": {Description: "Invalid CSRF token"},
			"404": {Description: "Owner or target resource not found"},
			"409": {Description: "Idempotency key conflict"},
		},
	}, func(ctx context.Context, input *SubmitSentenceFeedbackInput) (*SubmitSentenceFeedbackOutput, error) {
		result, err := svc.SubmitSentenceFeedback(ctx, aifeedback.SubmitSentenceFeedbackRequest{
			UserID:         RequesterUserID(ctx),
			SentenceText:   input.Body.SentenceText,
			Source:         input.Body.Source,
			AttemptID:      parseUUID(input.Body.AttemptID),
			IdempotencyKey: input.IdempotencyKey,
		})
		if err != nil {
			return nil, mapAIFeedbackError(err)
		}
		if result.ErrorCode == aifeedback.ErrorCodeIdempotencyConflict {
			return nil, huma.Error409Conflict("idempotency key conflict")
		}
		if result.ErrorCode == aifeedback.ValidationCodeAttemptNotEligible {
			return nil, huma.Error404NotFound("target not found")
		}
		return &SubmitSentenceFeedbackOutput{Body: sentenceFeedbackResultToDTO(result)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "ReportSentenceFeedback",
		Method:      http.MethodPost,
		Path:        "/api/v1/sentence-feedback/{attemptId}/reports",
		Summary:     "Report a quality concern for a feedback attempt",
		Tags:        []string{"AI Feedback"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"204": {Description: "Report recorded"},
			"401": {Description: "Authentication is required"},
			"403": {Description: "Invalid CSRF token"},
			"404": {Description: "Feedback attempt not found"},
		},
	}, func(ctx context.Context, input *ReportSentenceFeedbackInput) (*ReportSentenceFeedbackOutput, error) {
		if err := svc.ReportFeedback(ctx, RequesterUserID(ctx), parseUUID(input.AttemptID), input.Body.Reason, input.Body.Classification); err != nil {
			return nil, mapAIFeedbackError(err)
		}
		return &ReportSentenceFeedbackOutput{}, nil
	})
}

func sentenceFeedbackResultToDTO(r *aifeedback.SentenceFeedbackResult) SentenceFeedbackResultDTO {
	dto := SentenceFeedbackResultDTO{
		Status:                r.Status,
		OriginalSentence:      r.OriginalSentence,
		CorrectedSentence:     r.CorrectedSentence,
		Explanation:           r.Explanation,
		ImprovementTip:        r.ImprovementTip,
		MissionCompleted:      r.MissionCompleted,
		CanRetry:              r.CanRetry,
		Reported:              r.Reported,
		ErrorCode:             r.ErrorCode,
		ErrorMessage:          r.ErrorMessage,
		CrisisResourceMessage: r.CrisisResourceMessage,
	}
	if r.SentenceID != uuid.Nil {
		dto.SentenceID = r.SentenceID.String()
	}
	if r.AttemptID != uuid.Nil {
		dto.AttemptID = r.AttemptID.String()
	}
	return dto
}

func mapAIFeedbackError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, aifeedback.ErrTargetNotFound):
		return huma.Error404NotFound("target not found")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}
