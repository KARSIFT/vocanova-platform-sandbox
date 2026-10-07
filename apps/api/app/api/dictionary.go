package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/dictionary"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/danielgtaylor/huma/v2"
)

type DictionaryLookupInput struct {
	Query string `query:"q" doc:"One explicitly submitted English word, maximum 48 letters with optional internal apostrophes or hyphens; not a phrase or browse request"`
}
type DictionaryDefinitionDTO struct {
	Definition string `json:"definition"`
	Example    string `json:"example,omitempty"`
}
type DictionaryMeaningDTO struct {
	PartOfSpeech string                    `json:"partOfSpeech"`
	Definitions  []DictionaryDefinitionDTO `json:"definitions" maxItems:"2"`
}
type DictionaryLicenseDTO struct {
	Name string `json:"name"`
	URL  string `json:"url" format:"uri"`
	Text string `json:"text" doc:"Complete original license notice and disclaimer"`
}
type DictionaryAttributionDTO struct {
	Provider    string                 `json:"provider"`
	ProviderURL string                 `json:"providerUrl" format:"uri"`
	SourceURLs  []string               `json:"sourceUrls"`
	Licenses    []DictionaryLicenseDTO `json:"licenses"`
}
type DictionaryEntryDTO struct {
	Word        string                   `json:"word"`
	Meanings    []DictionaryMeaningDTO   `json:"meanings" maxItems:"4"`
	Attribution DictionaryAttributionDTO `json:"attribution"`
}
type DictionaryLookupOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         DictionaryEntryDTO
}

type dictionaryLookup interface {
	Lookup(context.Context, string) (dictionary.Entry, error)
}

func RegisterDictionary(api huma.API, svc dictionaryLookup) {
	registerDictionary(api, svc, auth.NewFixedWindowRateLimiter(clock.Real{}, time.Minute, 20))
}

func registerDictionary(api huma.API, svc dictionaryLookup, limiter auth.RateLimiter) {
	huma.Register(api, huma.Operation{
		OperationID: "LookupDictionary", Method: http.MethodGet, Path: "/api/v1/dictionary",
		Summary: "Read a locally bundled English dictionary entry without saving or grading it",
		Tags:    []string{"Discovery"}, Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"400": {Description: "Enter one English word"}, "401": {Description: "Authentication is required"},
			"404": {Description: "No entry in the bundled dictionary"}, "429": {Description: "Lookup limit reached; retry in a minute"},
			"503": {Description: "Dictionary temporarily unavailable"},
		},
	}, func(ctx context.Context, input *DictionaryLookupInput) (*DictionaryLookupOutput, error) {
		if c := HumaContext(ctx); c != nil {
			c.SetHeader("Cache-Control", "no-store")
		}
		word, err := dictionary.NormalizeWord(input.Query)
		if err != nil {
			return nil, huma.Error400BadRequest("Enter one English word, up to 48 letters, with optional apostrophes or hyphens.")
		}
		allowed, err := limiter.Allow(ctx, "dictionary:"+RequesterUserID(ctx).String())
		if err != nil {
			return nil, huma.Error503ServiceUnavailable("Dictionary is unavailable right now.")
		}
		if !allowed {
			if c := HumaContext(ctx); c != nil {
				c.SetHeader("Retry-After", "60")
			}
			return nil, huma.Error429TooManyRequests("Try another lookup in a minute.")
		}
		entry, err := svc.Lookup(ctx, word)
		if err != nil {
			if errors.Is(err, dictionary.ErrNotFound) {
				return nil, huma.Error404NotFound("This word is not in the dictionary yet.")
			}
			return nil, huma.Error503ServiceUnavailable("Dictionary is unavailable right now.")
		}
		out := &DictionaryLookupOutput{CacheControl: "no-store", Body: DictionaryEntryDTO{
			Word: entry.Word, Meanings: []DictionaryMeaningDTO{},
			Attribution: DictionaryAttributionDTO{Provider: "WordNet 3.0", ProviderURL: dictionary.SourceURL,
				SourceURLs: []string{dictionary.SourceURL}, Licenses: []DictionaryLicenseDTO{{Name: "WordNet 3.0", URL: dictionary.LicenseURL, Text: entry.License}}},
		}}
		for _, meaning := range entry.Meanings {
			m := DictionaryMeaningDTO{PartOfSpeech: meaning.PartOfSpeech, Definitions: []DictionaryDefinitionDTO{}}
			for _, definition := range meaning.Definitions {
				m.Definitions = append(m.Definitions, DictionaryDefinitionDTO{Definition: definition.Definition, Example: definition.Example})
			}
			out.Body.Meanings = append(out.Body.Meanings, m)
		}
		return out, nil
	})
}
