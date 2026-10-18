// Package meaninguses holds original reviewed supplemental uses of familiar words.
// These senses share existing canonical words but never change the ranked map.
package meaninguses

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type Entry struct {
	WordID, MeaningID, ExampleID                            uuid.UUID
	Word, PartOfSpeech, Definition, Example, SourceSenseKey string
	MeaningOrder                                            int
}

type wordIdentity struct{ WordID, PrimaryMeaningID uuid.UUID }

// Identities are pinned to the existing public english-10000-v1 source.
var wordIdentities = map[string]wordIdentity{
	"light":   {uuid.MustParse("d3171fea-62e1-5152-8763-7e1e93afac63"), uuid.MustParse("fea2d3ee-7b71-5169-bd05-eca61f77ab2d")},
	"fast":    {uuid.MustParse("44f271c8-0316-5911-b287-525260e4672b"), uuid.MustParse("5c324296-e5df-5972-bd71-8aced7f75cff")},
	"well":    {uuid.MustParse("adc1e53c-3712-54a1-99aa-189737b16cfe"), uuid.MustParse("b416d749-2f17-513f-8b62-beedcc0c27a4")},
	"present": {uuid.MustParse("90d23847-fefc-5911-b17d-f134e172c0ad"), uuid.MustParse("99f8ea33-db8a-597d-94a1-74b3189ab393")},
	"record":  {uuid.MustParse("f2dbeda9-fa45-57b5-aa2c-5f4efb89e33c"), uuid.MustParse("45cb435d-695a-5acb-bd3e-f227ab6fe217")},
	"close":   {uuid.MustParse("ff39a43c-7432-578f-bf74-0853c2beff38"), uuid.MustParse("b74d7fde-9d83-5261-a612-2a76f29a4962")},
	"resume":  {uuid.MustParse("ecd746b9-be52-5640-8eab-17f86e033d7e"), uuid.MustParse("e5817d4f-eaf9-577e-881f-b100fe1e24f8")},
	"read":    {uuid.MustParse("645d3bfe-dcb9-56d9-929c-8b201dcbff6b"), uuid.MustParse("9e8813c5-81d2-584e-8197-882110c19402")},
	"object":  {uuid.MustParse("71dbed7b-eb93-5093-bae4-2ff1c2182fc0"), uuid.MustParse("7765db80-a8fc-5dd7-bf70-bcee11cb798d")},
	"minute":  {uuid.MustParse("4e7ab557-6e49-5672-8693-d3c43624a611"), uuid.MustParse("876a876e-6424-57af-9e0d-23a4e59dbb52")},
	"live":    {uuid.MustParse("ade9fc5c-d2ef-5873-9ae0-86b581c8a74a"), uuid.MustParse("89524b17-9cd6-54fc-9691-c221cf1ce490")},
	"fine":    {uuid.MustParse("cf01c68c-5401-59c2-a53a-229ada48dbba"), uuid.MustParse("622d3ed7-f297-5871-9bec-80eab9c3ada3")},
}

// Append only: source keys and explicit orders are stable public identities.
var reviewedUses = []struct {
	word, pos, key, definition, example string
	order                               int
}{
	{"light", "verb", "illuminate", "To make a place bright.", "These lamps light the room.", 1001},
	{"light", "adjective", "little-weight", "Having little weight.", "This light bag is easy to carry.", 1002},
	{"light", "adverb", "little-luggage", "With very little luggage.", "We travel light on short trips.", 1003},
	{"fast", "adverb", "quickly", "At high speed.", "The train moves fast.", 1004},
	{"well", "adjective", "healthy", "Healthy and not ill.", "I feel well today.", 1005},
	{"well", "noun", "water-source", "A deep hole used to obtain water from the ground.", "They get water from a well.", 1006},
	{"present", "noun", "gift", "Something given to another person as a gift.", "This present is for you.", 1007},
	{"present", "verb", "show", "To show or give something formally to other people.", "We will present our plan tomorrow.", 1008},
	{"record", "noun", "stored-information", "Information kept so that it can be checked later.", "Keep a record of every payment.", 1009},
	{"close", "verb", "shut", "To move something so that an opening is covered.", "Please close the door.", 1010},
	{"close", "adverb", "nearby", "At a short distance.", "Please stay close.", 1011},
	{"resume", "verb", "resume-after-pause", "To begin again after a pause.", "We will resume the lesson after lunch.", 1012},
	{"read", "verb", "past-or-participle", "Past tense or past participle of read: looked at and understood written words before now.", "Yesterday, I read the whole story.", 1013},
	{"object", "verb", "disagree", "To say that you disagree with something.", "Some people object to the new rule.", 1014},
	{"minute", "adjective", "extremely-small", "Extremely small.", "The picture shows minute details.", 1015},
	{"live", "adjective", "happening-now", "Happening now rather than recorded earlier.", "The live concert starts at eight.", 1016},
	{"live", "adverb", "during-event", "As an event is happening rather than from a recording.", "The concert will be broadcast live tonight.", 1017},
	{"fine", "noun", "money-penalty", "Money that someone must pay as punishment for breaking a rule.", "The driver paid a fine.", 1018},
	{"fine", "verb", "impose-money-penalty", "To require someone to pay money for breaking a rule.", "The court can fine a driver for speeding.", 1019},
}

// Entries returns value copies, so callers cannot mutate the reviewed content.
func Entries() []Entry {
	out := make([]Entry, 0, len(reviewedUses))
	for _, row := range reviewedUses {
		word := wordIdentities[row.word]
		key := row.pos + ":" + row.key
		id := uuid.NewSHA1(word.WordID, []byte("vocanova:meaning-use-v1:"+key))
		out = append(out, Entry{WordID: word.WordID, MeaningID: id, ExampleID: uuid.NewSHA1(id, []byte("vocanova:meaning-use-v1:example")), Word: row.word, PartOfSpeech: row.pos, Definition: row.definition, Example: row.example, SourceSenseKey: "vocanova:meaning-use-v1:" + row.word + ":" + key, MeaningOrder: row.order})
	}
	return out
}

func SupportsMeaning(id, word, definition string) bool {
	for _, entry := range Entries() {
		if entry.MeaningID.String() == id && entry.Word == word && entry.Definition == definition {
			return true
		}
	}
	return false
}

// ReviewedContext is pronunciation guidance, never speech to prepend to the word.
// In particular, the primary read use is present tense; its past use has a
// separate frozen meaning identity and an explicitly past-tense example.
func ReviewedContext(id, word, definition string) (string, bool) {
	for _, entry := range Entries() {
		if entry.MeaningID.String() == id && entry.Word == word && entry.Definition == definition {
			return entry.Example, true
		}
	}
	if id == "9e8813c5-81d2-584e-8197-882110c19402" && word == "read" && definition == "To look at written words and understand what they mean." {
		return "I read a book every evening.", true
	}
	if id == "45cb435d-695a-5acb-bd3e-f227ab6fe217" && word == "record" && definition == "To store information, sound, or pictures so that they can be used later." {
		return "We record every lesson.", true
	}
	if id == "446b7b9e-8190-5987-9401-0378e95ff493" && word == "follow-up" && definition == "A later message or action." {
		return "The manager scheduled a follow-up for Monday.", true
	}
	if id == "f582a137-c77d-58df-942e-884fe336cd64" && word == "deadline" && definition == "When work must be submitted." {
		return "The essay deadline is at noon.", true
	}
	return "", false
}

// Seed runs in the caller's canonical-content transaction. It adds only these
// reviewed senses/examples; conflicting identities fail without overwriting rows.
func Seed(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("seed meaning uses: transaction required")
	}
	for word, expected := range wordIdentities {
		var text, normalized, language, status string
		if err := tx.QueryRowContext(ctx, `SELECT text,normalized_text,language_code,status FROM canonical_words WHERE id=$1 FOR SHARE`, expected.WordID).Scan(&text, &normalized, &language, &status); err != nil {
			return fmt.Errorf("seed meaning uses: missing canonical word %q: %w", word, err)
		}
		if text != word || normalized != word || language != "en" || status != "active" {
			return fmt.Errorf("seed meaning uses: canonical identity mismatch for %q", word)
		}
		var primary bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM word_meanings WHERE id=$1 AND word_id=$2 AND status='active')`, expected.PrimaryMeaningID, expected.WordID).Scan(&primary); err != nil {
			return fmt.Errorf("seed meaning uses: primary check for %q: %w", word, err)
		}
		if !primary {
			return fmt.Errorf("seed meaning uses: primary meaning mismatch for %q", word)
		}
	}
	for _, entry := range Entries() {
		if _, err := tx.ExecContext(ctx, `INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,learner_definition,meaning_order,status,difficulty_level,created_at,updated_at) VALUES($1,$2,$3,$4,$4,$5,'active','unknown',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP) ON CONFLICT(id) DO NOTHING`, entry.MeaningID, entry.WordID, entry.PartOfSpeech, entry.Definition, entry.MeaningOrder); err != nil {
			return fmt.Errorf("seed meaning uses %q: %w", entry.SourceSenseKey, err)
		}
		var matches bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM word_meanings WHERE id=$1 AND word_id=$2 AND part_of_speech=$3 AND short_definition=$4 AND meaning_order=$5 AND status='active')`, entry.MeaningID, entry.WordID, entry.PartOfSpeech, entry.Definition, entry.MeaningOrder).Scan(&matches); err != nil {
			return fmt.Errorf("seed meaning uses: validate %q: %w", entry.SourceSenseKey, err)
		}
		if !matches {
			return fmt.Errorf("seed meaning uses: conflicting meaning identity for %q", entry.SourceSenseKey)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO word_examples(id,meaning_id,example_text,example_order,status,created_at,updated_at) VALUES($1,$2,$3,1,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP) ON CONFLICT(id) DO NOTHING`, entry.ExampleID, entry.MeaningID, entry.Example); err != nil {
			return fmt.Errorf("seed meaning uses example %q: %w", entry.SourceSenseKey, err)
		}
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM word_examples WHERE id=$1 AND meaning_id=$2 AND example_text=$3 AND example_order=1 AND status='active')`, entry.ExampleID, entry.MeaningID, entry.Example).Scan(&matches); err != nil {
			return fmt.Errorf("seed meaning uses: validate example %q: %w", entry.SourceSenseKey, err)
		}
		if !matches {
			return fmt.Errorf("seed meaning uses: conflicting example identity for %q", entry.SourceSenseKey)
		}
	}
	return nil
}
