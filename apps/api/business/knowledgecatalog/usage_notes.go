package knowledgecatalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Each note belongs to one ranked editorial meaning. Pin its definition so a
// later change of sense requires reviewing the note as well.
var firstRangeNotes = []struct {
	word, definition, kind, text string
}{
	{"a", "Used before a singular countable noun when the listener does not yet know which one.", "grammar", "Use a before a consonant sound, as in a bus or a university."},
	{"an", "A form of a used before a word that starts with a vowel sound.", "grammar", "Choose an by sound. Say an hour because the h is silent."},
	{"the", "Used before a noun when the speaker and listener can identify which one is meant.", "grammar", "For a singular countable thing, use a when you first mention it and the once the listener knows which one. I saw a cat. The cat was asleep."},
	{"for", "Used to show the person, purpose, or reason something is intended for.", "collocation", "Use for with a purpose or intended recipient, as in a bag for shopping or a card for Mia."},
	{"just", "Only; no more than the amount or thing mentioned.", "collocation", "In just one ticket and just a minute, just limits the amount."},
	{"as", "Used to describe the role or purpose of someone or something.", "collocation", "Use as to name a role. She works as a nurse. Use like when comparing people or things."},
	{"too", "More than is wanted or needed.", "common_mistake", "Too hot means the heat causes a problem. Very hot describes the temperature without saying it causes a problem."},
	{"would", "Used to describe an imagined situation or make a polite request.", "register", "Would you open the window? is a polite request. It is natural at work, in a shop, or with someone you know."},
	{"by", "Used to show who does an action or how something is done.", "collocation", "Use by to name a method, as in pay by card, travel by bus, and contact us by email."},
	{"let", "To allow someone to do something.", "common_mistake", "Use the base verb without to after let. Say They let me leave early."},
}

type catalogUsageNote struct {
	ID        uuid.UUID `json:"id"`
	MeaningID uuid.UUID `json:"meaning_id"`
	Kind      string    `json:"kind"`
	Text      string    `json:"note_text"`
}

func rankedUsageNotes(catalog []CatalogEntry) ([]catalogUsageNote, error) {
	byWord := make(map[string]CatalogEntry, 100)
	for _, entry := range catalog[:min(len(catalog), 100)] {
		if entry.Source == "Vocanova editorial core" {
			byWord[entry.Word] = entry
		}
	}
	notes := make([]catalogUsageNote, 0, len(firstRangeNotes))
	for _, note := range firstRangeNotes {
		meaning, ok := byWord[note.word]
		if !ok || meaning.Definition != note.definition {
			return nil, fmt.Errorf("usage guidance for %q needs review against the catalog sense", note.word)
		}
		notes = append(notes, catalogUsageNote{
			ID:        uuid.NewSHA1(meaning.MeaningID, []byte("ranked-usage-note-v1")),
			MeaningID: meaning.MeaningID,
			Kind:      note.kind,
			Text:      note.text,
		})
	}
	return notes, nil
}

func seedRankedUsageNotes(ctx context.Context, tx *sql.Tx, catalog []CatalogEntry) error {
	notes, err := rankedUsageNotes(catalog)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(notes)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO usage_notes
	  (id,meaning_id,note_type,note_text,note_order,status,created_at,updated_at)
	  SELECT id,meaning_id,kind,note_text,1001,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
	  FROM jsonb_to_recordset($1::jsonb) AS x(id uuid,meaning_id uuid,kind text,note_text text)
	  ON CONFLICT(id) DO UPDATE SET note_type=EXCLUDED.note_type,note_text=EXCLUDED.note_text,updated_at=CURRENT_TIMESTAMP
	  WHERE usage_notes.meaning_id=EXCLUDED.meaning_id AND
	    (usage_notes.note_type IS DISTINCT FROM EXCLUDED.note_type OR usage_notes.note_text IS DISTINCT FROM EXCLUDED.note_text)`, string(payload))
	if err != nil {
		return fmt.Errorf("seed reviewed catalog usage notes: %w", err)
	}
	return nil
}
