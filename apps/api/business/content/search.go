package content

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var ErrInvalidSearch = errors.New("invalid search query or filter")

type SearchRequest struct {
	UserID      uuid.UUID
	Knowledge   string
	Query       string
	Category    string
	Level       string
	AfterCursor string
	Limit       int
}

// SearchMeaning is one canonical meaning, with a requester-scoped saved state.
type SearchMeaning struct {
	MeaningSummary
	DifficultyLevel string
	UserWordID      uuid.UUID
	ReviewState     string
	Due             bool
	normalizedText  string
}

type SearchResponse struct {
	Items      []SearchMeaning
	TotalCount int
	NextCursor string
}

type searchCursor struct {
	Version int       `json:"v"`
	Filter  string    `json:"f"`
	Text    string    `json:"t"`
	ID      uuid.UUID `json:"i"`
}

func prepareSearch(req SearchRequest) (SearchRequest, searchCursor, error) {
	req.Query = strings.ToLower(strings.TrimSpace(req.Query))
	if !utf8.ValidString(req.Query) || strings.ContainsRune(req.Query, 0) || utf8.RuneCountInString(req.Query) > 100 {
		return req, searchCursor{}, ErrInvalidSearch
	}
	switch req.Knowledge {
	case "", "known", "saved", "unexplored":
	default:
		return req, searchCursor{}, ErrInvalidSearch
	}
	switch req.Category {
	case "", "daily_life", "travel", "work", "study", "social":
	default:
		return req, searchCursor{}, ErrInvalidSearch
	}
	switch req.Level {
	case "", "a1", "a2", "b1", "b2", "c1", "unknown":
	default:
		return req, searchCursor{}, ErrInvalidSearch
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50
	}
	var cursor searchCursor
	if req.AfterCursor != "" {
		if len(req.AfterCursor) > 2048 {
			return req, cursor, ErrInvalidCursor
		}
		data, err := base64.RawURLEncoding.DecodeString(req.AfterCursor)
		if err != nil || json.Unmarshal(data, &cursor) != nil || cursor.Version != 1 || cursor.ID == uuid.Nil || cursor.Text == "" || strings.ContainsRune(cursor.Text, 0) || cursor.Filter != searchFilter(req) {
			return req, cursor, ErrInvalidCursor
		}
	}
	return req, cursor, nil
}

func searchFilter(req SearchRequest) string {
	data, _ := json.Marshal([]string{req.Query, req.Category, req.Level, req.Knowledge, req.UserID.String()})
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func searchNextCursor(req SearchRequest, item SearchMeaning) string {
	data, _ := json.Marshal(searchCursor{Version: 1, Filter: searchFilter(req), Text: item.normalizedText, ID: item.MeaningID})
	return base64.RawURLEncoding.EncodeToString(data)
}

func (s *Service) Search(ctx context.Context, userID uuid.UUID, req SearchRequest) (*SearchResponse, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	req.UserID = userID
	resp, err := s.repo.SearchMeanings(ctx, req)
	if err != nil {
		return nil, err
	}
	idsKnown := make([]uuid.UUID, len(resp.Items))
	for i, item := range resp.Items {
		idsKnown[i] = item.MeaningID
	}
	known, err := s.knownStates(ctx, userID, idsKnown)
	if err != nil {
		return nil, err
	}
	for i := range resp.Items {
		resp.Items[i].SelfReportedKnown = known[resp.Items[i].MeaningID]
	}
	if s.reader == nil || len(resp.Items) == 0 {
		return resp, nil
	}
	ids := make([]uuid.UUID, len(resp.Items))
	for i, item := range resp.Items {
		ids[i] = item.MeaningID
	}
	states, err := s.reader.SavedWordStates(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	for i := range resp.Items {
		if state, ok := states[resp.Items[i].MeaningID]; ok {
			resp.Items[i].Saved = true
			resp.Items[i].UserWordID = state.UserWordID
			resp.Items[i].ReviewState = state.Status
			resp.Items[i].Due = state.Due
		}
	}
	return resp, nil
}

func (r *MemoryRepository) SearchMeanings(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	req, cursor, err := prepareSearch(req)
	if err != nil {
		return nil, err
	}
	items := []SearchMeaning{}
	for _, meaning := range r.meanings {
		word, ok := r.findWord(meaning.WordID)
		if !ok || word.Status != "active" || meaning.Status != "active" {
			continue
		}
		level := meaning.DifficultyLevel
		if level == "" {
			level = word.DifficultyLevel
		}
		if level == "" {
			level = "unknown"
		}
		if req.Level != "" && level != req.Level {
			continue
		}
		if req.Query != "" && !strings.Contains(strings.ToLower(word.Text), req.Query) && !strings.Contains(strings.ToLower(meaning.ShortDefinition), req.Query) && !strings.Contains(strings.ToLower(meaning.LearnerDefinition), req.Query) {
			continue
		}
		if req.Category != "" {
			matches := false
			for _, link := range r.journeyWords {
				if link.MeaningID != meaning.ID {
					continue
				}
				for _, situation := range r.situations {
					if situation.ID == link.JourneySituationID && situation.Status == "active" && situation.Category == req.Category {
						matches = true
					}
				}
			}
			if !matches {
				continue
			}
		}
		items = append(items, SearchMeaning{MeaningSummary: MeaningSummary{MeaningID: meaning.ID, WordID: word.ID, WordText: word.Text, WordSlug: wordSlug(word.NormalizedText), PartOfSpeech: meaning.PartOfSpeech, ShortDefinition: meaning.ShortDefinition}, DifficultyLevel: level, normalizedText: word.NormalizedText})
	}
	if req.Knowledge != "" {
		ids := make([]uuid.UUID, len(items))
		for i, item := range items {
			ids[i] = item.MeaningID
		}
		known := map[uuid.UUID]bool{}
		saved := map[uuid.UUID]bool{}
		if r.knowledge != nil {
			known, err = r.knowledge.KnownStates(ctx, req.UserID, ids)
			if err != nil {
				return nil, err
			}
		}
		if r.savedReader != nil {
			saved, err = r.savedReader.IsSaved(ctx, req.UserID, ids)
			if err != nil {
				return nil, err
			}
		}
		filtered := items[:0]
		for _, item := range items {
			k, v := known[item.MeaningID], saved[item.MeaningID]
			if (req.Knowledge == "known" && k) || (req.Knowledge == "saved" && v) || (req.Knowledge == "unexplored" && !k && !v) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].normalizedText != items[j].normalizedText {
			return items[i].normalizedText < items[j].normalizedText
		}
		return items[i].MeaningID.String() < items[j].MeaningID.String()
	})
	resp := &SearchResponse{Items: []SearchMeaning{}, TotalCount: len(items)}
	for _, item := range items {
		if cursor.ID != uuid.Nil && (item.normalizedText < cursor.Text || (item.normalizedText == cursor.Text && item.MeaningID.String() <= cursor.ID.String())) {
			continue
		}
		resp.Items = append(resp.Items, item)
		if len(resp.Items) > req.Limit {
			resp.Items = resp.Items[:req.Limit]
			resp.NextCursor = searchNextCursor(req, resp.Items[len(resp.Items)-1])
			break
		}
	}
	return resp, nil
}
