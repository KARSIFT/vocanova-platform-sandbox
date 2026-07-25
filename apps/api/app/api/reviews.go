package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/reviews"
	"github.com/danielgtaylor/huma/v2"
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

// RegisterReviews registers the review-domain read routes.
func RegisterReviews(api huma.API, svc *reviews.Service) {
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

func mapReviewsError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, reviews.ErrInvalidCursor):
		return huma.Error400BadRequest("invalid cursor")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}
