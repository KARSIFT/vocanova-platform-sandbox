package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// SentenceFeedbackResultDTO is the public response body for a sentence-feedback
// submission. It exposes only the fields the frontend is permitted to display.
type SentenceFeedbackResultDTO struct {
	FeedbackID              string     `json:"feedbackId,omitempty" format:"uuid" doc:"Feedback result identifier"`
	SentenceID              string     `json:"sentenceId,omitempty" format:"uuid" doc:"Learner sentence identifier"`
	AttemptID               string     `json:"attemptId,omitempty" format:"uuid" doc:"AI feedback attempt identifier"`
	TargetWordID            string     `json:"targetWordId,omitempty" format:"uuid" doc:"Canonical target word identifier"`
	ProcessingStatus        string     `json:"processingStatus,omitempty" enum:"pending,completed,failed,skipped" doc:"Public processing state"`
	Status                  string     `json:"status,omitempty" enum:"correct,needs_improvement,incorrect" doc:"Pedagogical outcome when feedback succeeded"`
	OriginalSentence        string     `json:"originalSentence" maxLength:"300" doc:"The sentence the learner submitted"`
	CorrectedSentence       *string    `json:"correctedSentence" doc:"Corrected sentence when useful"`
	Headline                string     `json:"headline,omitempty" maxLength:"60" doc:"Encouraging, honest feedback headline"`
	Explanation             string     `json:"explanation,omitempty" maxLength:"240" doc:"Short explanation of the result"`
	ImprovementTip          *string    `json:"improvementTip" maxLength:"160" doc:"One improvement tip when useful"`
	TargetWordUsedCorrectly bool       `json:"targetWordUsedCorrectly" doc:"Whether the target vocabulary was used correctly"`
	GrammarAcceptable       bool       `json:"grammarAcceptable" doc:"Whether grammar was acceptable for the learner level"`
	MeaningClear            bool       `json:"meaningClear" doc:"Whether the intended meaning was clear"`
	Naturalness             string     `json:"naturalness,omitempty" enum:"natural,understandable,unnatural" doc:"Sentence naturalness"`
	MissionCompleted        bool       `json:"missionCompleted" doc:"Backend-confirmed mission state"`
	CanRetry                bool       `json:"canRetry" doc:"Whether the learner may safely retry"`
	Reported                bool       `json:"reported" doc:"Whether the learner has reported this feedback"`
	ErrorCode               string     `json:"errorCode,omitempty" doc:"Stable public error code on failure"`
	ErrorMessage            string     `json:"errorMessage,omitempty" doc:"Safe retryable message on failure"`
	CrisisResourceMessage   string     `json:"crisisResourceMessage,omitempty" doc:"Non-clinical crisis resources for self-harm content"`
	CreatedAt               *time.Time `json:"createdAt,omitempty" format:"date-time" doc:"When the sentence was submitted"`
}

// LearnerSentenceDTO is the public, privacy-safe retained sentence projection.
type LearnerSentenceDTO struct {
	ID                      string    `json:"id" format:"uuid" doc:"Learner sentence identifier"`
	FeedbackID              string    `json:"feedbackId,omitempty" format:"uuid" doc:"Latest feedback identifier"`
	TargetWordID            string    `json:"targetWordId,omitempty" format:"uuid" doc:"Canonical target word identifier"`
	ProcessingStatus        string    `json:"processingStatus" enum:"pending,completed,failed,skipped" doc:"Public processing state"`
	Status                  string    `json:"status,omitempty" enum:"correct,needs_improvement,incorrect" doc:"Learning result when processing completed"`
	OriginalSentence        string    `json:"originalSentence" doc:"Sentence submitted by the learner"`
	CorrectedSentence       *string   `json:"correctedSentence" doc:"Corrected sentence when useful"`
	Headline                string    `json:"headline,omitempty" maxLength:"60" doc:"Encouraging, honest feedback headline"`
	Explanation             string    `json:"explanation,omitempty" maxLength:"240" doc:"Short explanation of the result"`
	ImprovementTip          *string   `json:"improvementTip" maxLength:"160" doc:"One improvement tip when useful"`
	TargetWordUsedCorrectly bool      `json:"targetWordUsedCorrectly" doc:"Whether the target vocabulary was used correctly"`
	GrammarAcceptable       bool      `json:"grammarAcceptable" doc:"Whether grammar was acceptable for the learner level"`
	MeaningClear            bool      `json:"meaningClear" doc:"Whether the intended meaning was clear"`
	Naturalness             string    `json:"naturalness,omitempty" enum:"natural,understandable,unnatural" doc:"Sentence naturalness when processing completed"`
	Reported                bool      `json:"reported" doc:"Whether the learner reported this feedback"`
	CreatedAt               time.Time `json:"createdAt" format:"date-time" doc:"When the sentence was submitted"`
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

type ListLearnerSentencesInput struct {
	After string `query:"after" doc:"Opaque pagination cursor"`
	Limit int    `query:"limit" default:"20" doc:"Requested page size; defaults to 20 and is capped at 50"`
}

type ListLearnerSentencesOutput struct {
	Body struct {
		Items      []LearnerSentenceDTO `json:"items" doc:"Retained learner sentences"`
		NextCursor string               `json:"nextCursor,omitempty" doc:"Opaque cursor for the next page"`
		HasMore    bool                 `json:"hasMore" doc:"Whether another page is available"`
	}
}

type GetLearnerSentenceInput struct {
	SentenceID string `path:"id" format:"uuid" required:"true" doc:"Learner sentence identifier"`
}

type GetLearnerSentenceOutput struct {
	Body LearnerSentenceDTO
}

// ReportSentenceFeedbackInput reports a quality concern for a feedback attempt.
type ReportSentenceFeedbackInput struct {
	AttemptID      string `path:"attemptId" format:"uuid" required:"true" doc:"AI feedback attempt identifier"`
	IdempotencyKey string `header:"Idempotency-Key" required:"true" doc:"User-scoped idempotency key"`
	Body           struct {
		Reason string `json:"reason" enum:"already_correct,correction_changed_meaning,explanation_unclear,inappropriate,something_else" required:"true" doc:"Learner-selected report reason"`
	}
}

// ReportSentenceFeedbackOutput is an empty successful response.
type ReportSentenceFeedbackOutput struct{}

// RegisterAIFeedback registers the sentence-feedback write and report routes.
func RegisterAIFeedback(api huma.API, svc *aifeedback.Service, authSvc *auth.Service) {
	submit := func(ctx context.Context, input *SubmitSentenceFeedbackInput) (*SubmitSentenceFeedbackOutput, error) {
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
	}

	registerSubmission := func(operationID, path string) {
		huma.Register(api, huma.Operation{
			OperationID: operationID,
			Method:      http.MethodPost,
			Path:        path,
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
		}, submit)
	}

	registerSubmission("CreateLearnerSentence", "/api/v1/learner-sentences")
	registerSubmission("SubmitSentenceFeedback", "/api/v1/sentence-feedback")

	huma.Register(api, huma.Operation{
		OperationID: "ListLearnerSentences",
		Method:      http.MethodGet,
		Path:        "/api/v1/learner-sentences",
		Summary:     "List the authenticated learner's retained sentences",
		Tags:        []string{"AI Feedback"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"400": {Description: "Invalid pagination cursor"},
			"401": {Description: "Authentication is required"},
		},
	}, func(ctx context.Context, input *ListLearnerSentencesInput) (*ListLearnerSentencesOutput, error) {
		page, err := svc.ListLearnerSentences(ctx, aifeedback.ListLearnerSentencesRequest{
			UserID: RequesterUserID(ctx), AfterCursor: input.After, Limit: input.Limit,
		})
		if err != nil {
			return nil, mapAIFeedbackError(err)
		}
		out := &ListLearnerSentencesOutput{}
		out.Body.Items = make([]LearnerSentenceDTO, len(page.Items))
		for i, item := range page.Items {
			out.Body.Items[i] = learnerSentenceToDTO(item)
		}
		out.Body.NextCursor = page.NextCursor
		out.Body.HasMore = page.NextCursor != ""
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "GetLearnerSentence",
		Method:      http.MethodGet,
		Path:        "/api/v1/learner-sentences/{id}",
		Summary:     "Get one retained learner sentence",
		Tags:        []string{"AI Feedback"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
			"404": {Description: "Learner sentence not found"},
		},
	}, func(ctx context.Context, input *GetLearnerSentenceInput) (*GetLearnerSentenceOutput, error) {
		item, err := svc.GetLearnerSentence(ctx, RequesterUserID(ctx), parseUUID(input.SentenceID))
		if err != nil {
			return nil, mapAIFeedbackError(err)
		}
		return &GetLearnerSentenceOutput{Body: learnerSentenceToDTO(*item)}, nil
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
			"409": {Description: "Idempotency key conflict"},
		},
	}, func(ctx context.Context, input *ReportSentenceFeedbackInput) (*ReportSentenceFeedbackOutput, error) {
		if err := svc.ReportFeedback(ctx, RequesterUserID(ctx), parseUUID(input.AttemptID), input.Body.Reason, input.IdempotencyKey); err != nil {
			if errors.Is(err, aifeedback.ErrReportIdempotencyConflict) {
				return nil, huma.Error409Conflict("idempotency key conflict")
			}
			return nil, mapAIFeedbackError(err)
		}
		return &ReportSentenceFeedbackOutput{}, nil
	})
}

func sentenceFeedbackResultToDTO(r *aifeedback.SentenceFeedbackResult) SentenceFeedbackResultDTO {
	dto := SentenceFeedbackResultDTO{
		FeedbackID:              uuidString(r.FeedbackID),
		TargetWordID:            uuidString(r.TargetWordID),
		ProcessingStatus:        r.ProcessingStatus,
		Status:                  r.Status,
		OriginalSentence:        r.OriginalSentence,
		CorrectedSentence:       r.CorrectedSentence,
		Headline:                r.Headline,
		Explanation:             r.Explanation,
		ImprovementTip:          r.ImprovementTip,
		TargetWordUsedCorrectly: r.TargetWordUsedCorrectly,
		GrammarAcceptable:       r.GrammarAcceptable,
		MeaningClear:            r.MeaningClear,
		Naturalness:             r.Naturalness,
		MissionCompleted:        r.MissionCompleted,
		CanRetry:                r.CanRetry,
		Reported:                r.Reported,
		ErrorCode:               r.ErrorCode,
		ErrorMessage:            r.ErrorMessage,
		CrisisResourceMessage:   r.CrisisResourceMessage,
		CreatedAt:               optionalTime(r.CreatedAt),
	}
	if r.SentenceID != uuid.Nil {
		dto.SentenceID = r.SentenceID.String()
	}
	if r.AttemptID != uuid.Nil {
		dto.AttemptID = r.AttemptID.String()
	}
	return dto
}

func learnerSentenceToDTO(item aifeedback.LearnerSentence) LearnerSentenceDTO {
	return LearnerSentenceDTO{
		ID: item.ID.String(), FeedbackID: uuidString(item.FeedbackID), TargetWordID: uuidString(item.TargetWordID),
		ProcessingStatus: item.ProcessingStatus, Status: item.Status, OriginalSentence: item.OriginalSentence,
		CorrectedSentence: item.CorrectedSentence, Headline: item.Headline, Explanation: item.Explanation,
		ImprovementTip: item.ImprovementTip, TargetWordUsedCorrectly: item.TargetWordUsedCorrectly,
		GrammarAcceptable: item.GrammarAcceptable, MeaningClear: item.MeaningClear,
		Naturalness: item.Naturalness, Reported: item.Reported, CreatedAt: item.CreatedAt,
	}
}

func uuidString(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

func optionalTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func mapAIFeedbackError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, aifeedback.ErrTargetNotFound):
		return huma.Error404NotFound("target not found")
	case errors.Is(err, aifeedback.ErrInvalidCursor):
		return huma.Error400BadRequest("invalid cursor")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}
