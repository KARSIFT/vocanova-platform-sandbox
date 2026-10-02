package api

import (
	"context"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/content"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type SearchCanonicalWordsInput struct {
	Knowledge string `query:"knowledge" doc:"Optional requester-owned filter: known (self-assessed), saved, unexplored (neither); empty means all"`
	Query     string `query:"q" doc:"Case-insensitive literal text in a word or definition; trimmed, maximum 100 characters"`
	Category  string `query:"category" doc:"Optional active situation category: daily_life, travel, work, study, social"`
	Level     string `query:"level" doc:"Optional effective meaning difficulty: a1, a2, b1, b2, c1, unknown"`
	After     string `query:"after" doc:"Opaque pagination cursor bound to the search filters"`
	Limit     int    `query:"limit" default:"20" doc:"Page size; defaults to 20 and is capped at 50"`
}

type SearchMeaningDTO struct {
	MeaningID         string `json:"meaningId" format:"uuid" doc:"Meaning identifier"`
	WordID            string `json:"wordId" format:"uuid" doc:"Canonical word identifier"`
	WordSlug          string `json:"wordSlug" doc:"Canonical word URL slug"`
	WordText          string `json:"wordText" doc:"Canonical word text"`
	PartOfSpeech      string `json:"partOfSpeech" doc:"Part of speech"`
	ShortDefinition   string `json:"shortDefinition" doc:"Definition of this meaning"`
	DifficultyLevel   string `json:"difficultyLevel" doc:"Meaning difficulty, falling back to word difficulty or unknown; an editorial band, not a learner proficiency estimate"`
	SelfReportedKnown bool   `json:"selfReportedKnown" doc:"Requester self-assessment, independent of saved state or review mastery"`
	Saved             bool   `json:"saved" doc:"Whether the requester has saved this meaning"`
	UserWordID        string `json:"userWordId,omitempty" format:"uuid" doc:"Requester-owned saved record, when saved"`
	ReviewState       string `json:"reviewState,omitempty" enum:"new,learning,reviewing,mastered,ignored,archived" doc:"Requester-owned learning state, when saved"`
	Due               bool   `json:"due" doc:"Whether this saved meaning is due according to server time"`
}

type SearchCanonicalWordsOutput struct {
	Body struct {
		Items      []SearchMeaningDTO `json:"items" doc:"Matching canonical meanings with requester saved state"`
		TotalCount int                `json:"totalCount" doc:"Total matching meanings across all pages"`
		NextCursor string             `json:"nextCursor,omitempty" doc:"Cursor for the next page"`
		HasMore    bool               `json:"hasMore" doc:"Whether another page exists"`
	}
}

func registerCanonicalSearch(api huma.API, svc *content.Service) {
	huma.Register(api, huma.Operation{OperationID: "SearchCanonicalWords", Method: http.MethodGet, Path: "/api/v1/canonical-words", Summary: "Search canonical meanings with saved state", Tags: []string{"Discovery"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()}, Responses: map[string]*huma.Response{"400": {Description: "Invalid query, filter or cursor"}, "401": {Description: "Authentication is required"}}}, func(ctx context.Context, input *SearchCanonicalWordsInput) (*SearchCanonicalWordsOutput, error) {
		resp, err := svc.Search(ctx, RequesterUserID(ctx), content.SearchRequest{Knowledge: input.Knowledge, Query: input.Query, Category: input.Category, Level: input.Level, AfterCursor: input.After, Limit: input.Limit})
		if err != nil {
			return nil, mapContentError(err)
		}
		out := &SearchCanonicalWordsOutput{}
		out.Body.Items = make([]SearchMeaningDTO, len(resp.Items))
		out.Body.TotalCount = resp.TotalCount
		out.Body.NextCursor = resp.NextCursor
		out.Body.HasMore = resp.NextCursor != ""
		for i, item := range resp.Items {
			id := ""
			if item.UserWordID != uuid.Nil {
				id = item.UserWordID.String()
			}
			out.Body.Items[i] = SearchMeaningDTO{MeaningID: item.MeaningID.String(), WordID: item.WordID.String(), WordSlug: item.WordSlug, WordText: item.WordText, PartOfSpeech: item.PartOfSpeech, ShortDefinition: item.ShortDefinition, DifficultyLevel: item.DifficultyLevel, Saved: item.Saved, SelfReportedKnown: item.SelfReportedKnown, UserWordID: id, ReviewState: item.ReviewState, Due: item.Due}
		}
		return out, nil
	})
}
