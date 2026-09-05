package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/content"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// SituationDTO is a public journey situation projection.
type SituationDTO struct {
	ID               string `json:"id" format:"uuid" doc:"Situation identifier"`
	Slug             string `json:"slug" doc:"URL slug"`
	Title            string `json:"title" doc:"Human-readable title"`
	ShortDescription string `json:"shortDescription" doc:"Short description"`
	LevelBand        string `json:"levelBand,omitempty" doc:"Suggested level band"`
	Category         string `json:"category" doc:"Situation category"`
	DisplayOrder     int    `json:"displayOrder" doc:"Display order"`
}

// SituationMeaningDTO is a meaning entry shown in a situation drill-down.
type SituationMeaningDTO struct {
	MeaningID       string `json:"meaningId" format:"uuid" doc:"Meaning identifier"`
	WordID          string `json:"wordId" format:"uuid" doc:"Canonical word identifier"`
	WordSlug        string `json:"wordSlug" doc:"Canonical word URL slug"`
	WordText        string `json:"wordText" doc:"Canonical word text"`
	PartOfSpeech    string `json:"partOfSpeech" doc:"Part of speech"`
	ShortDefinition string `json:"shortDefinition" doc:"Short definition"`
	Saved           bool   `json:"saved" doc:"Whether the authenticated requester has saved this meaning"`
}

// WordExampleDTO is a canonical example sentence.
type WordExampleDTO struct {
	ID             string `json:"id" format:"uuid" doc:"Example identifier"`
	ExampleText    string `json:"exampleText" doc:"Example sentence"`
	SituationLabel string `json:"situationLabel,omitempty" doc:"Optional situation label"`
}

// WordUsageNoteDTO is a usage note for a meaning.
type WordUsageNoteDTO struct {
	ID       string `json:"id" format:"uuid" doc:"Note identifier"`
	NoteType string `json:"noteType" doc:"Note type"`
	NoteText string `json:"noteText" doc:"Note text"`
}

// WordMeaningDTO is a meaning with examples and usage notes.
type WordMeaningDTO struct {
	ID                string             `json:"id" format:"uuid" doc:"Meaning identifier"`
	PartOfSpeech      string             `json:"partOfSpeech" doc:"Part of speech"`
	ShortDefinition   string             `json:"shortDefinition" doc:"Short definition"`
	LearnerDefinition string             `json:"learnerDefinition,omitempty" doc:"Learner-friendly definition"`
	Saved             bool               `json:"saved" doc:"Whether the authenticated requester has saved this meaning"`
	UserWordID        string             `json:"userWordId,omitempty" format:"uuid" doc:"Saved record identifier when this meaning is saved"`
	Examples          []WordExampleDTO   `json:"examples" doc:"Example sentences"`
	UsageNotes        []WordUsageNoteDTO `json:"usageNotes" doc:"Usage notes"`
}

// WordDetailDTO is a canonical word with its meanings.
type WordDetailDTO struct {
	ID              string           `json:"id" format:"uuid" doc:"Word identifier"`
	Text            string           `json:"text" doc:"Word text"`
	Slug            string           `json:"slug" doc:"URL slug"`
	WordType        string           `json:"wordType" doc:"Word type"`
	DifficultyLevel string           `json:"difficultyLevel,omitempty" doc:"Difficulty level"`
	Meanings        []WordMeaningDTO `json:"meanings" doc:"Meanings"`
}

// ListSituationsInput requests a paginated list of active journey situations.
type ListSituationsInput struct {
	After string `query:"after" doc:"Opaque pagination cursor"`
	Limit int    `query:"limit" default:"20" doc:"Requested page size; defaults to 20 and is capped at 50"`
}

// ListSituationsOutput returns a page of situations.
type ListSituationsOutput struct {
	Body struct {
		Items      []SituationDTO `json:"items" doc:"Situations"`
		NextCursor string         `json:"nextCursor,omitempty" doc:"Opaque cursor for the next page"`
	}
}

// GetSituationInput requests a situation by slug.
type GetSituationInput struct {
	Slug string `path:"slug" doc:"Situation slug"`
}

// GetSituationOutput returns a situation with its meanings.
type GetSituationOutput struct {
	Body struct {
		Situation SituationDTO          `json:"situation" doc:"Situation"`
		Meanings  []SituationMeaningDTO `json:"meanings" doc:"Meanings in the situation"`
	}
}

// GetWordInput requests a canonical word by slug.
type GetWordInput struct {
	WordSlug string `path:"wordSlug" doc:"Canonical word slug"`
}

// GetWordOutput returns a canonical word with its meanings.
type GetWordOutput struct {
	Body struct {
		Word WordDetailDTO `json:"word" doc:"Canonical word detail"`
	}
}

// RegisterContent registers the discovery and canonical word read routes.
// usersSvc is used, read-only, to look up the requester's onboarding
// main-use-case answer so the first page of Discover can surface
// goal-relevant situations first (VOC-1183); it may be nil, in which
// case Discover ordering falls back to plain display_order.
func RegisterContent(api huma.API, svc *content.Service, usersSvc *users.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "ListJourneySituations",
		Method:      http.MethodGet,
		Path:        "/api/v1/journey-situations",
		Summary:     "List active journey situations",
		Tags:        []string{"Discovery"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"400": {Description: "Invalid pagination cursor"},
			"401": {Description: "Authentication is required"},
		},
	}, func(ctx context.Context, input *ListSituationsInput) (*ListSituationsOutput, error) {
		resp, err := svc.ListSituations(ctx, content.ListSituationsRequest{
			AfterCursor:      input.After,
			Limit:            input.Limit,
			PriorityCategory: requesterMainUseCase(ctx, usersSvc),
		})
		if err != nil {
			return nil, mapContentError(err)
		}
		out := &ListSituationsOutput{}
		out.Body.Items = make([]SituationDTO, len(resp.Items))
		for i, s := range resp.Items {
			out.Body.Items[i] = situationToDTO(s)
		}
		out.Body.NextCursor = resp.NextCursor
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "GetJourneySituation",
		Method:      http.MethodGet,
		Path:        "/api/v1/journey-situations/{slug}",
		Summary:     "Get a journey situation with meanings",
		Tags:        []string{"Discovery"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
			"404": {Description: "Situation not found"},
		},
	}, func(ctx context.Context, input *GetSituationInput) (*GetSituationOutput, error) {
		detail, err := svc.GetSituation(ctx, RequesterUserID(ctx), input.Slug)
		if err != nil {
			return nil, mapContentError(err)
		}
		out := &GetSituationOutput{}
		out.Body.Situation = situationToDTO(detail.Situation)
		out.Body.Meanings = make([]SituationMeaningDTO, len(detail.Meanings))
		for i, m := range detail.Meanings {
			out.Body.Meanings[i] = SituationMeaningDTO{
				MeaningID:       m.MeaningID.String(),
				WordID:          m.WordID.String(),
				WordSlug:        m.WordSlug,
				WordText:        m.WordText,
				PartOfSpeech:    m.PartOfSpeech,
				ShortDefinition: m.ShortDefinition,
				Saved:           m.Saved,
			}
		}
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "GetCanonicalWord",
		Method:      http.MethodGet,
		Path:        "/api/v1/canonical-words/{wordSlug}",
		Summary:     "Get a canonical word with meanings",
		Tags:        []string{"Discovery"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"},
			"404": {Description: "Word not found"},
		},
	}, func(ctx context.Context, input *GetWordInput) (*GetWordOutput, error) {
		word, err := svc.GetWordDetail(ctx, RequesterUserID(ctx), input.WordSlug)
		if err != nil {
			return nil, mapContentError(err)
		}
		out := &GetWordOutput{}
		out.Body.Word = wordToDTO(word)
		return out, nil
	})
}

func situationToDTO(s content.Situation) SituationDTO {
	return SituationDTO{
		ID:               s.ID.String(),
		Slug:             s.Slug,
		Title:            s.Title,
		ShortDescription: s.ShortDescription,
		LevelBand:        s.LevelBand,
		Category:         s.Category,
		DisplayOrder:     s.DisplayOrder,
	}
}

func wordToDTO(w *content.WordDetail) WordDetailDTO {
	meanings := make([]WordMeaningDTO, len(w.Meanings))
	for i, m := range w.Meanings {
		examples := make([]WordExampleDTO, len(m.Examples))
		for j, e := range m.Examples {
			examples[j] = WordExampleDTO{
				ID:             e.ID.String(),
				ExampleText:    e.ExampleText,
				SituationLabel: e.SituationLabel,
			}
		}
		notes := make([]WordUsageNoteDTO, len(m.UsageNotes))
		for j, n := range m.UsageNotes {
			notes[j] = WordUsageNoteDTO{
				ID:       n.ID.String(),
				NoteType: n.NoteType,
				NoteText: n.NoteText,
			}
		}
		meanings[i] = WordMeaningDTO{
			ID:                m.ID.String(),
			PartOfSpeech:      m.PartOfSpeech,
			ShortDefinition:   m.ShortDefinition,
			LearnerDefinition: m.LearnerDefinition,
			Saved:             m.Saved,
			UserWordID:        m.UserWordID.String(),
			Examples:          examples,
			UsageNotes:        notes,
		}
	}
	return WordDetailDTO{
		ID:              w.ID.String(),
		Text:            w.Text,
		Slug:            w.Slug,
		WordType:        w.WordType,
		DifficultyLevel: w.DifficultyLevel,
		Meanings:        meanings,
	}
}

// requesterMainUseCase looks up the requester's completed onboarding
// profile and returns its main use case (e.g. "work", "travel"), which
// maps 1:1 onto journey_situations.category. It returns "" (no
// prioritization) whenever usersSvc is nil, there is no requester, no
// onboarding profile exists yet, or the lookup fails — a learner who
// hasn't onboarded, or any lookup error, simply gets the unprioritized
// display_order ordering rather than a broken Discover page.
func requesterMainUseCase(ctx context.Context, usersSvc *users.Service) string {
	if usersSvc == nil {
		return ""
	}
	uid := RequesterUserID(ctx)
	if uid == uuid.Nil {
		return ""
	}
	profile, err := usersSvc.GetOnboarding(ctx, uid)
	if err != nil || profile == nil || profile.Status != users.OnboardingStatusCompleted {
		return ""
	}
	return profile.MainUseCase
}

func mapContentError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, content.ErrSituationNotFound):
		return huma.Error404NotFound("situation not found")
	case errors.Is(err, content.ErrWordNotFound):
		return huma.Error404NotFound("word not found")
	case errors.Is(err, content.ErrInvalidCursor):
		return huma.Error400BadRequest("invalid cursor")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}
