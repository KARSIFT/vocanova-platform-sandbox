package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// DueWordDTO is a due saved meaning projected for review. It includes the backend
// review step for the prompt engine but must not be rendered raw to the learner.
type DueWordDTO struct {
	UserWordID      string `json:"userWordId" format:"uuid" doc:"Saved record identifier"`
	MeaningID       string `json:"meaningId" format:"uuid" doc:"Meaning identifier"`
	WordID          string `json:"wordId" format:"uuid" doc:"Canonical word identifier"`
	WordText        string `json:"wordText" doc:"Canonical word text"`
	WordSlug        string `json:"wordSlug" doc:"Canonical word URL slug"`
	PartOfSpeech    string `json:"partOfSpeech" doc:"Part of speech"`
	ShortDefinition string `json:"shortDefinition" doc:"Short definition"`
	Status          string `json:"status" doc:"Learning status"`
	ReviewStep      int    `json:"reviewStep" doc:"Backend review step"`
}

// ListReviewsDueInput requests a paginated list of the authenticated requester's
// due review words.
type ListReviewsDueInput struct {
	After string `query:"after" doc:"Opaque pagination cursor"`
	Limit int    `query:"limit" default:"20" doc:"Maximum items to return (1-100)"`
}

// ListReviewsDueOutput returns a page of due words and the total due count.
type ListReviewsDueOutput struct {
	Body struct {
		Items      []DueWordDTO `json:"items" doc:"Due words"`
		NextCursor string       `json:"nextCursor,omitempty" doc:"Opaque cursor for the next page"`
		TotalCount int          `json:"totalCount" doc:"Total due words for the requester"`
	}
}

// ReviewAttemptDTO is an immutable review submission result.
type ReviewAttemptDTO struct {
	AttemptID               string    `json:"attemptId" format:"uuid" doc:"Immutable attempt identifier"`
	UserWordID              string    `json:"userWordId" format:"uuid" doc:"Saved word identifier"`
	MeaningID               string    `json:"meaningId" format:"uuid" doc:"Meaning identifier"`
	AttemptType             string    `json:"attemptType" doc:"Attempt category"`
	PromptType              string    `json:"promptType" enum:"multiple_choice,self_check" doc:"Prompt type"`
	Result                  string    `json:"result" enum:"correct,incorrect,skipped" doc:"Objective result"`
	Rating                  string    `json:"rating" enum:"again,hard,good,easy" doc:"Learner rating"`
	ReviewStepBefore        int       `json:"reviewStepBefore" doc:"Review step before this attempt"`
	ReviewStepAfter         int       `json:"reviewStepAfter" doc:"Review step after this attempt"`
	AnsweredAt              time.Time `json:"answeredAt" format:"date-time" doc:"When the answer was submitted"`
	ResponseTimeMs          int       `json:"responseTimeMs" doc:"Client-measured response time in milliseconds"`
	SelectedOptionMeaningID string    `json:"selectedOptionMeaningId,omitempty" format:"uuid" doc:"Chosen option for multiple choice"`
	TypedAnswer             string    `json:"typedAnswer,omitempty" doc:"Typed answer if applicable"`
	WasHintUsed             bool      `json:"wasHintUsed" doc:"Whether a hint was shown"`
	Source                  string    `json:"source" doc:"Origin of the attempt"`
	ClientAttemptID         string    `json:"clientAttemptId" doc:"Client-provided idempotency identifier"`
	NextReviewAt            time.Time `json:"nextReviewAt" format:"date-time" doc:"When the word will next be due"`
}

// SubmitReviewInput requests a review submission for the authenticated requester.
type SubmitReviewInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" doc:"User-scoped idempotency key"`
	Body           struct {
		UserWordID              string         `json:"userWordId" format:"uuid" required:"true" doc:"Saved word identifier"`
		MeaningID               string         `json:"meaningId" format:"uuid" required:"true" doc:"Meaning identifier"`
		AttemptType             string         `json:"attemptType,omitempty" enum:"review" default:"review" doc:"Attempt category"`
		PromptType              string         `json:"promptType" enum:"multiple_choice,self_check" required:"true" doc:"Prompt type"`
		Result                  string         `json:"result" enum:"correct,incorrect,skipped" required:"true" doc:"Objective result"`
		Rating                  string         `json:"rating,omitempty" enum:"again,hard,good,easy" doc:"Learner rating"`
		AnsweredAt              time.Time      `json:"answeredAt" format:"date-time" required:"true" doc:"When the answer was submitted"`
		ResponseTimeMs          int            `json:"responseTimeMs,omitempty" default:"0" doc:"Client-measured response time in milliseconds"`
		SelectedOptionMeaningID string         `json:"selectedOptionMeaningId,omitempty" format:"uuid" doc:"Chosen option for multiple choice; required for non-skipped multiple-choice attempts and must agree with result"`
		TypedAnswer             string         `json:"typedAnswer,omitempty" doc:"Typed answer if applicable"`
		WasHintUsed             bool           `json:"wasHintUsed,omitempty" default:"false" doc:"Whether a hint was shown"`
		Source                  string         `json:"source,omitempty" enum:"review,review_session" default:"review" doc:"Origin of the attempt"`
		ClientAttemptID         string         `json:"clientAttemptId" required:"true" doc:"Client-provided idempotency identifier"`
		Metadata                map[string]any `json:"metadata,omitempty" doc:"Extra prompt context"`
	}
}

// SubmitReviewOutput returns the recorded review attempt and updated schedule.
type SubmitReviewOutput struct {
	Body ReviewAttemptDTO
}

// RegisterReviews registers the review-domain read and write routes.
func RegisterReviews(api huma.API, svc *reviews.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "GetReviewsDue",
		Method:      http.MethodGet,
		Path:        "/api/v1/reviews/due",
		Summary:     "List the authenticated requester's due review words",
		Tags:        []string{"Reviews"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
		},
	}, func(ctx context.Context, input *ListReviewsDueInput) (*ListReviewsDueOutput, error) {
		resp, err := svc.ListDueWords(ctx, reviews.ListDueWordsRequest{
			UserID:      RequesterUserID(ctx),
			AfterCursor: input.After,
			Limit:       input.Limit,
		})
		if err != nil {
			return nil, mapReviewsError(err)
		}
		out := &ListReviewsDueOutput{}
		out.Body.Items = make([]DueWordDTO, len(resp.Items))
		for i, d := range resp.Items {
			out.Body.Items[i] = dueWordToDTO(d)
		}
		out.Body.NextCursor = resp.NextCursor
		out.Body.TotalCount = resp.TotalCount
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "SubmitReview",
		Method:      http.MethodPost,
		Path:        "/api/v1/reviews/submissions",
		Summary:     "Submit a review attempt and update the schedule",
		Tags:        []string{"Reviews"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"200": {Description: "Review attempt recorded and schedule updated"},
			"400": {Description: "Invalid request"},
			"401": {Description: "Authentication is required"},
			"403": {Description: "Invalid CSRF token"},
			"404": {Description: "Saved word not found"},
			"409": {Description: "Idempotency conflict"},
		},
	}, func(ctx context.Context, input *SubmitReviewInput) (*SubmitReviewOutput, error) {
		attempt, err := svc.SubmitReview(ctx, reviews.SubmitReviewRequest{
			UserID:                  RequesterUserID(ctx),
			UserWordID:              parseUUID(input.Body.UserWordID),
			MeaningID:               parseUUID(input.Body.MeaningID),
			AttemptType:             input.Body.AttemptType,
			PromptType:              input.Body.PromptType,
			Result:                  input.Body.Result,
			Rating:                  input.Body.Rating,
			AnsweredAt:              input.Body.AnsweredAt.UTC(),
			ResponseTimeMs:          input.Body.ResponseTimeMs,
			SelectedOptionMeaningID: parseNullableUUID(input.Body.SelectedOptionMeaningID),
			TypedAnswer:             nullableString(input.Body.TypedAnswer),
			WasHintUsed:             input.Body.WasHintUsed,
			Source:                  input.Body.Source,
			ClientAttemptID:         input.Body.ClientAttemptID,
			Metadata:                input.Body.Metadata,
			IdempotencyKey:          input.IdempotencyKey,
		})
		if err != nil {
			return nil, mapReviewsError(err)
		}
		out := &SubmitReviewOutput{}
		out.Body = reviewAttemptToDTO(*attempt)
		return out, nil
	})
}

func dueWordToDTO(d reviews.DueWord) DueWordDTO {
	return DueWordDTO{
		UserWordID:      d.UserWordID.String(),
		MeaningID:       d.MeaningID.String(),
		WordID:          d.WordID.String(),
		WordText:        d.WordText,
		WordSlug:        d.WordSlug,
		PartOfSpeech:    d.PartOfSpeech,
		ShortDefinition: d.ShortDefinition,
		Status:          d.Status,
		ReviewStep:      d.ReviewStep,
	}
}

func reviewAttemptToDTO(a reviews.ReviewAttempt) ReviewAttemptDTO {
	dto := ReviewAttemptDTO{
		AttemptID:        a.ID.String(),
		UserWordID:       a.UserWordID.String(),
		MeaningID:        a.MeaningID.String(),
		AttemptType:      a.AttemptType,
		PromptType:       a.PromptType,
		Result:           a.Result,
		Rating:           a.Rating,
		ReviewStepBefore: a.ReviewStepBefore,
		ReviewStepAfter:  a.ReviewStepAfter,
		AnsweredAt:       a.AnsweredAt,
		ResponseTimeMs:   a.ResponseTimeMs,
		WasHintUsed:      a.WasHintUsed,
		Source:           a.Source,
		ClientAttemptID:  a.ClientAttemptID,
		NextReviewAt:     a.NextReviewAt,
	}
	if a.SelectedOptionMeaningID != nil {
		dto.SelectedOptionMeaningID = a.SelectedOptionMeaningID.String()
	}
	if a.TypedAnswer != nil {
		dto.TypedAnswer = *a.TypedAnswer
	}
	return dto
}

func parseNullableUUID(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func mapReviewsError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, reviews.ErrInvalidCursor):
		return huma.Error400BadRequest("invalid cursor")
	case errors.Is(err, reviews.ErrUserWordNotFound):
		return huma.Error404NotFound("saved word not found")
	case errors.Is(err, reviews.ErrInvalidPromptType),
		errors.Is(err, reviews.ErrInvalidResult),
		errors.Is(err, reviews.ErrInvalidRating),
		errors.Is(err, reviews.ErrInvalidRatingForResult),
		errors.Is(err, reviews.ErrMultipleChoiceSelectionRequired),
		errors.Is(err, reviews.ErrMultipleChoiceResultMismatch),
		errors.Is(err, reviews.ErrInvalidAttemptType),
		errors.Is(err, reviews.ErrInvalidSource),
		errors.Is(err, reviews.ErrClientAttemptIDRequired),
		errors.Is(err, reviews.ErrInvalidAnsweredAt),
		errors.Is(err, reviews.ErrInvalidResponseTimeMs):
		return huma.Error400BadRequest(err.Error())
	case errors.Is(err, reviews.ErrIdempotencyKeyRequired):
		return huma.Error422UnprocessableEntity("idempotency key required")
	case errors.Is(err, reviews.ErrIdempotencyConflict):
		return huma.Error409Conflict("idempotency conflict")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}
