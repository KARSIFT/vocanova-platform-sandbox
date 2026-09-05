package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// SavedMeaningDTO is a learner-owned saved meaning with canonical word details.
type SavedMeaningDTO struct {
	UserWordID      string    `json:"userWordId" format:"uuid" doc:"Saved record identifier"`
	MeaningID       string    `json:"meaningId" format:"uuid" doc:"Meaning identifier"`
	WordID          string    `json:"wordId" format:"uuid" doc:"Canonical word identifier"`
	WordText        string    `json:"wordText" doc:"Canonical word text"`
	WordSlug        string    `json:"wordSlug" doc:"Canonical word URL slug"`
	PartOfSpeech    string    `json:"partOfSpeech" doc:"Part of speech"`
	ShortDefinition string    `json:"shortDefinition" doc:"Short definition"`
	Status          string    `json:"status" doc:"Learning status"`
	Source          string    `json:"source" doc:"Origin of the save"`
	Saved           bool      `json:"saved" doc:"Whether the meaning is currently saved"`
	AddedAt         time.Time `json:"addedAt" format:"date-time" doc:"When the meaning was added to the learner list"`
}

// SaveUserWordInput requests saving a meaning for the authenticated requester.
type SaveUserWordInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" doc:"User-scoped idempotency key"`
	Body           struct {
		MeaningID string `json:"meaningId" format:"uuid" required:"true" doc:"Meaning identifier to save"`
		Source    string `json:"source" enum:"journey,search,manual" required:"true" doc:"Origin of the save"`
	}
}

// SaveUserWordOutput returns the saved meaning.
type SaveUserWordOutput struct {
	Body SavedMeaningDTO
}

// UnsaveUserWordInput requests removing a saved meaning for the authenticated requester.
type UnsaveUserWordInput struct {
	MeaningID string `path:"meaningId" format:"uuid" required:"true" doc:"Meaning identifier to unsave"`
}

// UnsaveUserWordOutput is an empty successful response.
type UnsaveUserWordOutput struct{}

// ListSavedWordsInput requests a paginated list of the authenticated requester's saved meanings.
type ListSavedWordsInput struct {
	After string `query:"after" doc:"Opaque pagination cursor"`
	Limit int    `query:"limit" default:"20" doc:"Requested page size; defaults to 20 and is capped at 50"`
}

// ListSavedWordsOutput returns a page of saved meanings.
type ListSavedWordsOutput struct {
	Body struct {
		Items      []SavedMeaningDTO `json:"items" doc:"Saved meanings"`
		NextCursor string            `json:"nextCursor,omitempty" doc:"Opaque cursor for the next page"`
	}
}

// RegisterLearning registers the user-words save/unsave/list routes.
func RegisterLearning(api huma.API, svc *learning.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "ListSavedWords",
		Method:      http.MethodGet,
		Path:        "/api/v1/user-words",
		Summary:     "List the authenticated requester's saved meanings",
		Tags:        []string{"Learning"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"400": {Description: "Invalid pagination cursor"},
			"401": {Description: "Authentication is required"},
		},
	}, func(ctx context.Context, input *ListSavedWordsInput) (*ListSavedWordsOutput, error) {
		resp, err := svc.ListSavedWords(ctx, learning.ListSavedWordsRequest{
			UserID:      RequesterUserID(ctx),
			AfterCursor: input.After,
			Limit:       input.Limit,
		})
		if err != nil {
			return nil, mapLearningError(err)
		}
		out := &ListSavedWordsOutput{}
		out.Body.Items = make([]SavedMeaningDTO, len(resp.Items))
		for i, m := range resp.Items {
			out.Body.Items[i] = savedMeaningToDTO(m)
		}
		out.Body.NextCursor = resp.NextCursor
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "SaveUserWord",
		Method:      http.MethodPost,
		Path:        "/api/v1/user-words",
		Summary:     "Save a meaning for the authenticated requester",
		Tags:        []string{"Learning"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"400": {Description: "Idempotency key required"},
			"401": {Description: "Authentication is required"},
			"403": {Description: "Invalid CSRF token"},
			"404": {Description: "Meaning not found"},
			"409": {Description: "Idempotency key conflict"},
		},
	}, func(ctx context.Context, input *SaveUserWordInput) (*SaveUserWordOutput, error) {
		m, err := svc.SaveUserWord(ctx, learning.SaveUserWordRequest{
			UserID:         RequesterUserID(ctx),
			MeaningID:      parseUUID(input.Body.MeaningID),
			Source:         input.Body.Source,
			IdempotencyKey: input.IdempotencyKey,
		})
		if err != nil {
			return nil, mapLearningError(err)
		}
		return &SaveUserWordOutput{Body: savedMeaningToDTO(*m)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "UnsaveUserWord",
		Method:      http.MethodDelete,
		Path:        "/api/v1/user-words/{meaningId}",
		Summary:     "Remove a saved meaning for the authenticated requester",
		Tags:        []string{"Learning"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
			"403": {Description: "Invalid CSRF token"},
			"404": {Description: "Saved meaning not found"},
		},
	}, func(ctx context.Context, input *UnsaveUserWordInput) (*UnsaveUserWordOutput, error) {
		if err := svc.UnsaveUserWord(ctx, RequesterUserID(ctx), parseUUID(input.MeaningID)); err != nil {
			return nil, mapLearningError(err)
		}
		return &UnsaveUserWordOutput{}, nil
	})
}

func savedMeaningToDTO(m learning.SavedMeaning) SavedMeaningDTO {
	return SavedMeaningDTO{
		UserWordID:      m.UserWordID.String(),
		MeaningID:       m.MeaningID.String(),
		WordID:          m.WordID.String(),
		WordText:        m.WordText,
		WordSlug:        m.WordSlug,
		PartOfSpeech:    m.PartOfSpeech,
		ShortDefinition: m.ShortDefinition,
		Status:          m.Status,
		Source:          m.Source,
		Saved:           m.Saved,
		AddedAt:         m.AddedAt,
	}
}

func mapLearningError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, learning.ErrMeaningNotFound):
		return huma.Error404NotFound("meaning not found")
	case errors.Is(err, learning.ErrUserWordNotFound):
		return huma.Error404NotFound("saved meaning not found")
	case errors.Is(err, learning.ErrIdempotencyConflict):
		return huma.Error409Conflict("idempotency key conflict")
	case errors.Is(err, learning.ErrIdempotencyKeyRequired):
		return huma.Error400BadRequest("idempotency key required")
	case errors.Is(err, learning.ErrInvalidCursor):
		return huma.Error400BadRequest("invalid cursor")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}

func parseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
