package content

import (
	"context"
	"sort"

	"github.com/google/uuid"
)

// MemoryRepository is a deterministic in-memory repository for service and route
// tests. It is not concurrency-safe.
type MemoryRepository struct {
	situations   []Situation
	words        []SeedWord
	meanings     []SeedMeaning
	examples     []SeedExample
	notes        []SeedNote
	journeyWords []SeedJourneyWord
}

// MemoryRepositoryData holds seed data for the memory repository.
type MemoryRepositoryData struct {
	Situations   []Situation
	Words        []SeedWord
	Meanings     []SeedMeaning
	Examples     []SeedExample
	Notes        []SeedNote
	JourneyWords []SeedJourneyWord
}

// NewMemoryRepository initializes an in-memory repository from data.
func NewMemoryRepository(data MemoryRepositoryData) *MemoryRepository {
	return &MemoryRepository{
		situations:   data.Situations,
		words:        data.Words,
		meanings:     data.Meanings,
		examples:     data.Examples,
		notes:        data.Notes,
		journeyWords: data.JourneyWords,
	}
}

func (r *MemoryRepository) ListSituations(ctx context.Context, req ListSituationsRequest) (*ListSituationsResponse, error) {
	items := make([]Situation, 0, len(r.situations))
	for _, s := range r.situations {
		if s.Status != "active" {
			continue
		}
		items = append(items, s)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].DisplayOrder == items[j].DisplayOrder {
			return items[i].ID.String() < items[j].ID.String()
		}
		return items[i].DisplayOrder < items[j].DisplayOrder
	})

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	start := 0
	if req.AfterCursor != "" {
		c, err := decodeSituationCursor(req.AfterCursor)
		if err != nil {
			return nil, ErrInvalidCursor
		}
		// Treat the cursor as an exclusive boundary. It may refer to a
		// situation that was removed after the prior request, or be past the
		// end of the collection.
		start = len(items)
		for i, s := range items {
			if s.DisplayOrder > c.DisplayOrder || (s.DisplayOrder == c.DisplayOrder && s.ID.String() > c.ID.String()) {
				start = i
				break
			}
		}
	}
	if start > len(items) {
		start = len(items)
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	page := items[start:end]
	resp := &ListSituationsResponse{Items: page}
	if end-start == limit && end < len(items) {
		last := page[len(page)-1]
		resp.NextCursor = encodeSituationCursor(situationCursor{DisplayOrder: last.DisplayOrder, ID: last.ID})
	}
	return resp, nil
}

func (r *MemoryRepository) GetSituationBySlug(ctx context.Context, slug string) (*Situation, error) {
	for _, s := range r.situations {
		if s.Slug == slug && s.Status == "active" {
			copy := s
			return &copy, nil
		}
	}
	return nil, ErrSituationNotFound
}

func (r *MemoryRepository) GetMeaningsBySituation(ctx context.Context, situationID uuid.UUID) ([]MeaningSummary, error) {
	var links []SeedJourneyWord
	for _, jw := range r.journeyWords {
		if jw.JourneySituationID == situationID {
			links = append(links, jw)
		}
	}
	sort.Slice(links, func(i, j int) bool {
		if links[i].IsCore != links[j].IsCore {
			return links[i].IsCore
		}
		if links[i].DisplayOrder == nil && links[j].DisplayOrder != nil {
			// PostgreSQL sorts nullable display_order values last for ASC,
			// matching the discovery query below.
			return false
		}
		if links[i].DisplayOrder != nil && links[j].DisplayOrder == nil {
			return true
		}
		if links[i].DisplayOrder != nil && *links[i].DisplayOrder != *links[j].DisplayOrder {
			return *links[i].DisplayOrder < *links[j].DisplayOrder
		}
		if links[i].RelevanceScore != links[j].RelevanceScore {
			return links[i].RelevanceScore > links[j].RelevanceScore
		}
		return links[i].MeaningID.String() < links[j].MeaningID.String()
	})

	var out []MeaningSummary
	for _, jw := range links {
		m, ok := r.findMeaning(jw.MeaningID)
		if !ok || m.Status != "active" {
			continue
		}
		w, ok := r.findWord(m.WordID)
		if !ok || w.Status != "active" {
			continue
		}
		out = append(out, MeaningSummary{
			MeaningID:       m.ID,
			WordID:          w.ID,
			WordSlug:        wordSlug(w.NormalizedText),
			WordText:        w.Text,
			PartOfSpeech:    m.PartOfSpeech,
			ShortDefinition: m.ShortDefinition,
		})
	}
	return out, nil
}

func (r *MemoryRepository) GetWordBySlug(ctx context.Context, slug string) (*WordDetail, error) {
	for _, w := range r.words {
		if w.Status != "active" {
			continue
		}
		if wordSlug(w.NormalizedText) == slug {
			return r.buildWordDetail(w), nil
		}
	}
	return nil, ErrWordNotFound
}

func (r *MemoryRepository) buildWordDetail(w SeedWord) *WordDetail {
	wd := WordDetail{
		ID:              w.ID,
		Text:            w.Text,
		NormalizedText:  w.NormalizedText,
		Slug:            wordSlug(w.NormalizedText),
		WordType:        w.WordType,
		DifficultyLevel: w.DifficultyLevel,
	}
	var meaningIDs []uuid.UUID
	for _, m := range r.meanings {
		if m.WordID == w.ID && m.Status == "active" {
			wd.Meanings = append(wd.Meanings, WordMeaning{
				ID:                m.ID,
				PartOfSpeech:      m.PartOfSpeech,
				ShortDefinition:   m.ShortDefinition,
				LearnerDefinition: m.LearnerDefinition,
				MeaningOrder:      m.MeaningOrder,
			})
			meaningIDs = append(meaningIDs, m.ID)
		}
	}
	sort.Slice(wd.Meanings, func(i, j int) bool {
		return wd.Meanings[i].ID.String() < wd.Meanings[j].ID.String()
	})
	sort.Slice(wd.Meanings, func(i, j int) bool {
		if wd.Meanings[i].MeaningOrder == wd.Meanings[j].MeaningOrder {
			return wd.Meanings[i].ID.String() < wd.Meanings[j].ID.String()
		}
		return wd.Meanings[i].MeaningOrder < wd.Meanings[j].MeaningOrder
	})
	for i := range wd.Meanings {
		wd.Meanings[i].Examples = r.examplesFor(wd.Meanings[i].ID)
		wd.Meanings[i].UsageNotes = r.notesFor(wd.Meanings[i].ID)
	}
	return &wd
}

func (r *MemoryRepository) findWord(id uuid.UUID) (SeedWord, bool) {
	for _, w := range r.words {
		if w.ID == id {
			return w, true
		}
	}
	return SeedWord{}, false
}

func (r *MemoryRepository) findMeaning(id uuid.UUID) (SeedMeaning, bool) {
	for _, m := range r.meanings {
		if m.ID == id {
			return m, true
		}
	}
	return SeedMeaning{}, false
}

func (r *MemoryRepository) examplesFor(meaningID uuid.UUID) []WordExample {
	var out []WordExample
	for _, e := range r.examples {
		if e.MeaningID == meaningID && e.Status == "active" {
			out = append(out, WordExample{ID: e.ID, ExampleText: e.ExampleText, SituationLabel: e.SituationLabel})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID.String() < out[j].ID.String() })
	return out
}

func (r *MemoryRepository) notesFor(meaningID uuid.UUID) []WordUsageNote {
	var out []WordUsageNote
	for _, n := range r.notes {
		if n.MeaningID == meaningID && n.Status == "active" {
			out = append(out, WordUsageNote{ID: n.ID, NoteType: n.NoteType, NoteText: n.NoteText})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID.String() < out[j].ID.String() })
	return out
}

// SeedWord is an in-memory canonical word record for tests.
type SeedWord struct {
	ID              uuid.UUID
	Text            string
	NormalizedText  string
	WordType        string
	LanguageCode    string
	Status          string
	DifficultyLevel string
}

// SeedMeaning is an in-memory word meaning record for tests.
type SeedMeaning struct {
	ID                uuid.UUID
	WordID            uuid.UUID
	PartOfSpeech      string
	ShortDefinition   string
	LearnerDefinition string
	MeaningOrder      int
	Status            string
	DifficultyLevel   string
}

// SeedExample is an in-memory example sentence record for tests.
type SeedExample struct {
	ID             uuid.UUID
	MeaningID      uuid.UUID
	ExampleText    string
	ExampleOrder   int
	Status         string
	SituationLabel string
}

// SeedNote is an in-memory usage note record for tests.
type SeedNote struct {
	ID        uuid.UUID
	MeaningID uuid.UUID
	NoteType  string
	NoteText  string
	NoteOrder int
	Status    string
}

// SeedJourneyWord is an in-memory journey word record for tests.
type SeedJourneyWord struct {
	ID                 uuid.UUID
	JourneySituationID uuid.UUID
	MeaningID          uuid.UUID
	RelevanceScore     int
	DisplayOrder       *int
	IsCore             bool
}

// MemorySavedStateReader is an in-memory read-only saved-state provider for tests.
type MemorySavedStateReader struct {
	data        map[uuid.UUID]map[uuid.UUID]bool
	userWordIDs map[uuid.UUID]map[uuid.UUID]uuid.UUID
	states      map[uuid.UUID]map[uuid.UUID]SavedWordState
}

// NewMemorySavedStateReader creates a saved-state reader from user->meaning maps.
func NewMemorySavedStateReader(data map[uuid.UUID]map[uuid.UUID]bool) *MemorySavedStateReader {
	return &MemorySavedStateReader{data: data}
}

// NewMemorySavedStateReaderWithIDs creates a saved-state reader that also returns
// the user_word_id for each saved meaning.
func NewMemorySavedStateReaderWithIDs(
	saved map[uuid.UUID]map[uuid.UUID]bool,
	userWordIDs map[uuid.UUID]map[uuid.UUID]uuid.UUID,
) *MemorySavedStateReader {
	return &MemorySavedStateReader{data: saved, userWordIDs: userWordIDs}
}

// NewMemorySavedStateReaderWithStates creates a reader with learner-safe Word
// Detail states. It is intentionally separate from the boolean-only helper so
// older situation-list tests remain focused on their own projection.
func NewMemorySavedStateReaderWithStates(states map[uuid.UUID]map[uuid.UUID]SavedWordState) *MemorySavedStateReader {
	saved := make(map[uuid.UUID]map[uuid.UUID]bool, len(states))
	userWordIDs := make(map[uuid.UUID]map[uuid.UUID]uuid.UUID, len(states))
	for userID, userStates := range states {
		saved[userID] = make(map[uuid.UUID]bool, len(userStates))
		userWordIDs[userID] = make(map[uuid.UUID]uuid.UUID, len(userStates))
		for meaningID, state := range userStates {
			saved[userID][meaningID] = true
			userWordIDs[userID][meaningID] = state.UserWordID
		}
	}
	return &MemorySavedStateReader{data: saved, userWordIDs: userWordIDs, states: states}
}

func (r *MemorySavedStateReader) IsSaved(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := make(map[uuid.UUID]bool, len(meaningIDs))
	userMap := r.data[userID]
	for _, id := range meaningIDs {
		out[id] = userMap[id]
	}
	return out, nil
}

// SavedUserWordIDs implements SavedStateReader.
func (r *MemorySavedStateReader) SavedUserWordIDs(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	out := make(map[uuid.UUID]uuid.UUID, len(meaningIDs))
	userMap := r.userWordIDs[userID]
	for _, id := range meaningIDs {
		if uid, ok := userMap[id]; ok && uid != uuid.Nil {
			out[id] = uid
		}
	}
	return out, nil
}

// SavedWordStates implements SavedStateReader.
func (r *MemorySavedStateReader) SavedWordStates(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]SavedWordState, error) {
	out := make(map[uuid.UUID]SavedWordState, len(meaningIDs))
	for _, id := range meaningIDs {
		if state, ok := r.states[userID][id]; ok {
			out[id] = state
			continue
		}
		if r.data[userID][id] {
			out[id] = SavedWordState{UserWordID: r.userWordIDs[userID][id], Status: "new", Due: true}
		}
	}
	return out, nil
}

// MustParseUUID is a test helper that panics on invalid input.
func MustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}
