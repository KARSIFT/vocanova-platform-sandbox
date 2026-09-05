package content

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgreSQLRepository implements Repository against the VOC-026 P1 migration
// schema.
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository creates a repository backed by db.
func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

// wordSlug returns a URL slug derived from normalized text.
func wordSlug(normalized string) string {
	return strings.ReplaceAll(normalized, " ", "-")
}

func (r *PostgreSQLRepository) ListSituations(ctx context.Context, req ListSituationsRequest) (*ListSituationsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var displayOrder sql.NullInt32
	var afterID uuid.UUID
	if req.AfterCursor != "" {
		c, err := decodeSituationCursor(req.AfterCursor)
		if err != nil {
			return nil, ErrInvalidCursor
		}
		displayOrder = sql.NullInt32{Int32: int32(c.DisplayOrder), Valid: true}
		afterID = c.ID
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, slug, title, short_description, level_band, category, status, display_order, created_at, updated_at
		 FROM journey_situations
		 WHERE status = 'active'
		   AND ($1::integer IS NULL OR display_order > $1 OR (display_order = $1 AND id > $2))
		 ORDER BY display_order ASC, id ASC
		 LIMIT $3`,
		displayOrder, afterID, limit+1)
	if err != nil {
		return nil, fmt.Errorf("list situations: %w", err)
	}
	defer rows.Close()

	items := []Situation{}
	for rows.Next() {
		var s Situation
		var levelBand, category sql.NullString
		if err := rows.Scan(&s.ID, &s.Slug, &s.Title, &s.ShortDescription, &levelBand, &category, &s.Status, &s.DisplayOrder, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan situation: %w", err)
		}
		s.LevelBand = levelBand.String
		s.Category = category.String
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list situations rows: %w", err)
	}

	resp := &ListSituationsResponse{}
	if len(items) > limit {
		items = items[:limit]
		last := items[len(items)-1]
		resp.NextCursor = encodeSituationCursor(situationCursor{DisplayOrder: last.DisplayOrder, ID: last.ID})
	}
	resp.Items = items
	return resp, nil
}

func (r *PostgreSQLRepository) GetSituationBySlug(ctx context.Context, slug string) (*Situation, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, slug, title, short_description, level_band, category, status, display_order, created_at, updated_at
		 FROM journey_situations
		 WHERE slug = $1 AND status = 'active'`, slug)
	var s Situation
	var levelBand, category sql.NullString
	err := row.Scan(&s.ID, &s.Slug, &s.Title, &s.ShortDescription, &levelBand, &category, &s.Status, &s.DisplayOrder, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSituationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get situation: %w", err)
	}
	s.LevelBand = levelBand.String
	s.Category = category.String
	return &s, nil
}

func (r *PostgreSQLRepository) GetMeaningsBySituation(ctx context.Context, situationID uuid.UUID) ([]MeaningSummary, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT m.id, m.word_id, cw.text, cw.normalized_text, m.part_of_speech, m.short_definition, jw.display_order
		 FROM journey_words jw
		 JOIN word_meanings m ON m.id = jw.meaning_id
		 JOIN canonical_words cw ON cw.id = m.word_id
		 WHERE jw.journey_situation_id = $1
		   AND m.status = 'active'
		   AND cw.status = 'active'
		 ORDER BY jw.is_core DESC, jw.display_order ASC NULLS LAST,
		          jw.relevance_score DESC, m.id ASC`,
		situationID)
	if err != nil {
		return nil, fmt.Errorf("list meanings by situation: %w", err)
	}
	defer rows.Close()

	meanings := []MeaningSummary{}
	for rows.Next() {
		var m MeaningSummary
		var normalizedText string
		var displayOrder sql.NullInt64
		if err := rows.Scan(&m.MeaningID, &m.WordID, &m.WordText, &normalizedText, &m.PartOfSpeech, &m.ShortDefinition, &displayOrder); err != nil {
			return nil, fmt.Errorf("scan meaning: %w", err)
		}
		m.WordSlug = wordSlug(normalizedText)
		meanings = append(meanings, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list meanings rows: %w", err)
	}
	return meanings, nil
}

func (r *PostgreSQLRepository) GetWordBySlug(ctx context.Context, slug string) (*WordDetail, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, text, normalized_text, word_type, difficulty_level
		 FROM canonical_words
		 WHERE status = 'active' AND REPLACE(normalized_text, ' ', '-') = $1`,
		slug)
	var w WordDetail
	var difficultyLevel sql.NullString
	err := row.Scan(&w.ID, &w.Text, &w.NormalizedText, &w.WordType, &difficultyLevel)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrWordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get word: %w", err)
	}
	w.Slug = wordSlug(w.NormalizedText)
	w.DifficultyLevel = difficultyLevel.String

	meanings, err := r.getMeaningsByWord(ctx, w.ID)
	if err != nil {
		return nil, err
	}
	w.Meanings = meanings
	return &w, nil
}

func (r *PostgreSQLRepository) getMeaningsByWord(ctx context.Context, wordID uuid.UUID) ([]WordMeaning, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, part_of_speech, short_definition, learner_definition, meaning_order
		 FROM word_meanings
		 WHERE word_id = $1 AND status = 'active'
		 ORDER BY meaning_order ASC, id ASC`,
		wordID)
	if err != nil {
		return nil, fmt.Errorf("list meanings by word: %w", err)
	}
	defer rows.Close()

	meaningIDs := []uuid.UUID{}
	meanings := []WordMeaning{}
	for rows.Next() {
		var m WordMeaning
		var learnerDef sql.NullString
		if err := rows.Scan(&m.ID, &m.PartOfSpeech, &m.ShortDefinition, &learnerDef, &m.MeaningOrder); err != nil {
			return nil, fmt.Errorf("scan word meaning: %w", err)
		}
		m.LearnerDefinition = learnerDef.String
		meaningIDs = append(meaningIDs, m.ID)
		meanings = append(meanings, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list meanings rows: %w", err)
	}

	if len(meaningIDs) == 0 {
		return meanings, nil
	}

	examples, err := r.getExamplesByMeanings(ctx, meaningIDs)
	if err != nil {
		return nil, err
	}
	notes, err := r.getNotesByMeanings(ctx, meaningIDs)
	if err != nil {
		return nil, err
	}

	for i := range meanings {
		meanings[i].Examples = examples[meanings[i].ID]
		meanings[i].UsageNotes = notes[meanings[i].ID]
	}
	return meanings, nil
}

func (r *PostgreSQLRepository) getExamplesByMeanings(ctx context.Context, meaningIDs []uuid.UUID) (map[uuid.UUID][]WordExample, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, meaning_id, example_text, situation_label
		 FROM word_examples
		 WHERE meaning_id = ANY($1::uuid[]) AND status = 'active'
		 ORDER BY example_order ASC, id ASC`,
		pq.Array(meaningIDs))
	if err != nil {
		return nil, fmt.Errorf("list examples: %w", err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID][]WordExample)
	for rows.Next() {
		var e WordExample
		var situationLabel sql.NullString
		if err := rows.Scan(&e.ID, &e.MeaningID, &e.ExampleText, &situationLabel); err != nil {
			return nil, fmt.Errorf("scan example: %w", err)
		}
		e.SituationLabel = situationLabel.String
		out[e.MeaningID] = append(out[e.MeaningID], e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list examples rows: %w", err)
	}
	return out, nil
}

func (r *PostgreSQLRepository) getNotesByMeanings(ctx context.Context, meaningIDs []uuid.UUID) (map[uuid.UUID][]WordUsageNote, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, meaning_id, note_type, note_text
		 FROM usage_notes
		 WHERE meaning_id = ANY($1::uuid[]) AND status = 'active'
		 ORDER BY note_order ASC, id ASC`,
		pq.Array(meaningIDs))
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID][]WordUsageNote)
	for rows.Next() {
		var n WordUsageNote
		if err := rows.Scan(&n.ID, &n.MeaningID, &n.NoteType, &n.NoteText); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		out[n.MeaningID] = append(out[n.MeaningID], n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list notes rows: %w", err)
	}
	return out, nil
}
