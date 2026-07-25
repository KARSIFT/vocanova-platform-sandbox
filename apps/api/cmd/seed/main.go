// Command seed loads the deterministic P1 canonical content seed into a PostgreSQL
// database in a single transaction. It is safe to rerun: existing rows are updated
// to the seed values by their fixed primary keys.
package main

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

//go:embed voc026-p1.json
var seedJSON []byte

type seedData struct {
	JourneySituations []journeySituation `json:"journey_situations"`
	CanonicalWords    []canonicalWord    `json:"canonical_words"`
	WordMeanings      []wordMeaning      `json:"word_meanings"`
	WordExamples      []wordExample      `json:"word_examples"`
	UsageNotes        []usageNote        `json:"usage_notes"`
	JourneyWords      []journeyWord      `json:"journey_words"`
}

type journeySituation struct {
	ID               string  `json:"id"`
	Slug             string  `json:"slug"`
	Title            string  `json:"title"`
	ShortDescription string  `json:"short_description"`
	LevelBand        *string `json:"level_band"`
	Category         string  `json:"category"`
	Status           string  `json:"status"`
	DisplayOrder     int     `json:"display_order"`
}

type canonicalWord struct {
	ID              string  `json:"id"`
	Text            string  `json:"text"`
	NormalizedText  string  `json:"normalized_text"`
	WordType        string  `json:"word_type"`
	LanguageCode    string  `json:"language_code"`
	Status          string  `json:"status"`
	DifficultyLevel *string `json:"difficulty_level"`
	FrequencyRank   *int    `json:"frequency_rank"`
}

type wordMeaning struct {
	ID                string  `json:"id"`
	WordID            string  `json:"word_id"`
	PartOfSpeech      string  `json:"part_of_speech"`
	ShortDefinition   string  `json:"short_definition"`
	LearnerDefinition *string `json:"learner_definition"`
	MeaningOrder      int     `json:"meaning_order"`
	Status            string  `json:"status"`
	DifficultyLevel   *string `json:"difficulty_level"`
}

type wordExample struct {
	ID              string  `json:"id"`
	MeaningID       string  `json:"meaning_id"`
	ExampleText     string  `json:"example_text"`
	ExampleOrder    int     `json:"example_order"`
	DifficultyLevel *string `json:"difficulty_level"`
	SituationLabel  *string `json:"situation_label"`
	Status          string  `json:"status"`
}

type usageNote struct {
	ID        string `json:"id"`
	MeaningID string `json:"meaning_id"`
	NoteType  string `json:"note_type"`
	NoteText  string `json:"note_text"`
	NoteOrder int    `json:"note_order"`
	Status    string `json:"status"`
}

type journeyWord struct {
	ID                 string `json:"id"`
	JourneySituationID string `json:"journey_situation_id"`
	MeaningID          string `json:"meaning_id"`
	RelevanceScore     int    `json:"relevance_score"`
	DisplayOrder       *int   `json:"display_order"`
	IsCore             bool   `json:"is_core"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	seed, err := loadSeed()
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := applySeed(tx, seed); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	fmt.Printf("seed applied: %d situations, %d words, %d meanings, %d examples, %d notes, %d journey_words\n",
		len(seed.JourneySituations), len(seed.CanonicalWords), len(seed.WordMeanings), len(seed.WordExamples), len(seed.UsageNotes), len(seed.JourneyWords))
	return nil
}

func loadSeed() (seedData, error) {
	var seed seedData
	if err := json.Unmarshal(seedJSON, &seed); err != nil {
		return seed, fmt.Errorf("parse seed json: %w", err)
	}
	return seed, nil
}

func applySeed(tx *sql.Tx, seed seedData) error {
	now := "now()"

	if len(seed.JourneySituations) > 0 {
		stmt := `INSERT INTO journey_situations (id, slug, title, short_description, level_band, category, status, display_order, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, ` + now + `, ` + now + `)
			ON CONFLICT (id) DO UPDATE SET
				slug = EXCLUDED.slug,
				title = EXCLUDED.title,
				short_description = EXCLUDED.short_description,
				level_band = EXCLUDED.level_band,
				category = EXCLUDED.category,
				status = EXCLUDED.status,
				display_order = EXCLUDED.display_order,
				updated_at = EXCLUDED.updated_at`
		for _, s := range seed.JourneySituations {
			if _, err := tx.Exec(stmt, s.ID, s.Slug, s.Title, s.ShortDescription, s.LevelBand, s.Category, s.Status, s.DisplayOrder); err != nil {
				return fmt.Errorf("upsert journey_situations %s: %w", s.Slug, err)
			}
		}
	}

	if len(seed.CanonicalWords) > 0 {
		stmt := `INSERT INTO canonical_words (id, text, normalized_text, word_type, language_code, status, difficulty_level, frequency_rank, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, ` + now + `, ` + now + `)
			ON CONFLICT (id) DO UPDATE SET
				text = EXCLUDED.text,
				normalized_text = EXCLUDED.normalized_text,
				word_type = EXCLUDED.word_type,
				language_code = EXCLUDED.language_code,
				status = EXCLUDED.status,
				difficulty_level = EXCLUDED.difficulty_level,
				frequency_rank = EXCLUDED.frequency_rank,
				updated_at = EXCLUDED.updated_at`
		for _, w := range seed.CanonicalWords {
			if _, err := tx.Exec(stmt, w.ID, w.Text, w.NormalizedText, w.WordType, w.LanguageCode, w.Status, w.DifficultyLevel, w.FrequencyRank); err != nil {
				return fmt.Errorf("upsert canonical_words %s: %w", w.NormalizedText, err)
			}
		}
	}

	if len(seed.WordMeanings) > 0 {
		stmt := `INSERT INTO word_meanings (id, word_id, part_of_speech, short_definition, learner_definition, meaning_order, status, difficulty_level, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, ` + now + `, ` + now + `)
			ON CONFLICT (id) DO UPDATE SET
				word_id = EXCLUDED.word_id,
				part_of_speech = EXCLUDED.part_of_speech,
				short_definition = EXCLUDED.short_definition,
				learner_definition = EXCLUDED.learner_definition,
				meaning_order = EXCLUDED.meaning_order,
				status = EXCLUDED.status,
				difficulty_level = EXCLUDED.difficulty_level,
				updated_at = EXCLUDED.updated_at`
		for _, m := range seed.WordMeanings {
			if _, err := tx.Exec(stmt, m.ID, m.WordID, m.PartOfSpeech, m.ShortDefinition, m.LearnerDefinition, m.MeaningOrder, m.Status, m.DifficultyLevel); err != nil {
				return fmt.Errorf("upsert word_meanings %s: %w", m.ID, err)
			}
		}
	}

	if len(seed.WordExamples) > 0 {
		stmt := `INSERT INTO word_examples (id, meaning_id, example_text, example_order, difficulty_level, situation_label, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, ` + now + `, ` + now + `)
			ON CONFLICT (id) DO UPDATE SET
				meaning_id = EXCLUDED.meaning_id,
				example_text = EXCLUDED.example_text,
				example_order = EXCLUDED.example_order,
				difficulty_level = EXCLUDED.difficulty_level,
				situation_label = EXCLUDED.situation_label,
				status = EXCLUDED.status,
				updated_at = EXCLUDED.updated_at`
		for _, e := range seed.WordExamples {
			if _, err := tx.Exec(stmt, e.ID, e.MeaningID, e.ExampleText, e.ExampleOrder, e.DifficultyLevel, e.SituationLabel, e.Status); err != nil {
				return fmt.Errorf("upsert word_examples %s: %w", e.ID, err)
			}
		}
	}

	if len(seed.UsageNotes) > 0 {
		stmt := `INSERT INTO usage_notes (id, meaning_id, note_type, note_text, note_order, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, ` + now + `, ` + now + `)
			ON CONFLICT (id) DO UPDATE SET
				meaning_id = EXCLUDED.meaning_id,
				note_type = EXCLUDED.note_type,
				note_text = EXCLUDED.note_text,
				note_order = EXCLUDED.note_order,
				status = EXCLUDED.status,
				updated_at = EXCLUDED.updated_at`
		for _, n := range seed.UsageNotes {
			if _, err := tx.Exec(stmt, n.ID, n.MeaningID, n.NoteType, n.NoteText, n.NoteOrder, n.Status); err != nil {
				return fmt.Errorf("upsert usage_notes %s: %w", n.ID, err)
			}
		}
	}

	if len(seed.JourneyWords) > 0 {
		stmt := `INSERT INTO journey_words (id, journey_situation_id, meaning_id, relevance_score, display_order, is_core, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, ` + now + `, ` + now + `)
			ON CONFLICT (id) DO UPDATE SET
				journey_situation_id = EXCLUDED.journey_situation_id,
				meaning_id = EXCLUDED.meaning_id,
				relevance_score = EXCLUDED.relevance_score,
				display_order = EXCLUDED.display_order,
				is_core = EXCLUDED.is_core,
				updated_at = EXCLUDED.updated_at`
		for _, j := range seed.JourneyWords {
			if _, err := tx.Exec(stmt, j.ID, j.JourneySituationID, j.MeaningID, j.RelevanceScore, j.DisplayOrder, j.IsCore); err != nil {
				return fmt.Errorf("upsert journey_words %s: %w", j.ID, err)
			}
		}
	}

	return nil
}
