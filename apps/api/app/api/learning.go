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
	Query string `query:"q" doc:"Literal word or short-definition search; up to 100 Unicode characters"`
	Stage string `query:"stage" enum:"new,learning,reviewing,mastered,ignored,archived" doc:"Learner-facing review stage; omit for all saved meanings"`
	Due   bool   `query:"due" default:"false" doc:"Only meanings currently eligible for scheduled review, without a daily target cap"`
	After string `query:"after" doc:"Opaque pagination cursor"`
	Limit int    `query:"limit" default:"20" doc:"Requested page size; defaults to 20 and is capped at 50"`
}

// ListSavedWordsOutput returns a page of saved meanings.
type ListSavedWordsOutput struct {
	Body struct {
		Items      []SavedCollectionMeaningDTO `json:"items" doc:"Saved meanings"`
		TotalCount int                         `json:"totalCount" doc:"Full requester-owned count matching the filters, before pagination"`
		NextCursor string                      `json:"nextCursor,omitempty" doc:"Opaque cursor for the next page"`
		HasMore    bool                        `json:"hasMore" doc:"Whether another page is available"`
	}
}

// SavedCollectionMeaningDTO adds current read-only scheduling metadata to list
// items without changing save or detail response contracts.
type SavedCollectionMeaningDTO struct {
	SavedMeaningDTO
	ReviewState string `json:"reviewState" doc:"Learner-facing review stage; reviewed legacy new records are learning"`
	Due         bool   `json:"due" doc:"Currently eligible for a scheduled review, without a daily target cap"`
}

// GetSavedWordInput requests one learner-owned saved record.
type GetSavedWordInput struct {
	UserWordID string `path:"userWordId" format:"uuid" required:"true" doc:"Saved record identifier"`
}

// GetSavedWordOutput returns one saved meaning with canonical summary data.
type GetSavedWordOutput struct {
	Body SavedMeaningDTO
}

// RegisterLearning registers the user-words save/unsave/list routes.
func RegisterLearning(api huma.API, svc *learning.Service, authSvc *auth.Service) {
	registerKnowledgeSummary(api, svc)
	huma.Register(api, huma.Operation{
		OperationID: "ListSavedWords",
		Method:      http.MethodGet,
		Path:        "/api/v1/user-words",
		Summary:     "List the authenticated requester's saved meanings",
		Tags:        []string{"Learning"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"400": {Description: "Invalid collection filter or pagination cursor"},
			"401": {Description: "Authentication is required"},
		},
	}, func(ctx context.Context, input *ListSavedWordsInput) (*ListSavedWordsOutput, error) {
		resp, err := svc.ListSavedWords(ctx, learning.ListSavedWordsRequest{
			UserID:      RequesterUserID(ctx),
			Query:       input.Query,
			Stage:       input.Stage,
			DueOnly:     input.Due,
			AfterCursor: input.After,
			Limit:       input.Limit,
		})
		if err != nil {
			return nil, mapLearningError(err)
		}
		out := &ListSavedWordsOutput{}
		out.Body.Items = make([]SavedCollectionMeaningDTO, len(resp.Items))
		for i, m := range resp.Items {
			out.Body.Items[i] = SavedCollectionMeaningDTO{SavedMeaningDTO: savedMeaningToDTO(m), ReviewState: m.ReviewState, Due: m.Due}
		}
		out.Body.TotalCount = resp.TotalCount
		out.Body.NextCursor = resp.NextCursor
		out.Body.HasMore = resp.NextCursor != ""
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "GetSavedWord",
		Method:      http.MethodGet,
		Path:        "/api/v1/user-words/records/{userWordId}",
		Summary:     "Get one saved word owned by the authenticated requester",
		Tags:        []string{"Learning"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
			"404": {Description: "Saved word not found"},
		},
	}, func(ctx context.Context, input *GetSavedWordInput) (*GetSavedWordOutput, error) {
		m, err := svc.GetSavedWord(ctx, RequesterUserID(ctx), parseUUID(input.UserWordID))
		if err != nil {
			return nil, mapLearningError(err)
		}
		return &GetSavedWordOutput{Body: savedMeaningToDTO(*m)}, nil
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
	case errors.Is(err, learning.ErrInvalidSavedFilter):
		return huma.Error400BadRequest("invalid saved collection filter")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}

func parseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
