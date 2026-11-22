package practice

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaninguses"
	"github.com/lib/pq"
)

// MistakesContentVersion marks a mistakes snapshot containing at least one
// ranked (non-starter) target. Starter-only mistakes keep ContentVersion, so
// existing sessions and behavior are unchanged.
const MistakesContentVersion = "mistakes-ranked-v1"

// rankedTargets holds mistake targets outside the reviewed starter catalog.
// Recognition targets have no curated typed cue: they are practised with a
// frozen meaning choice whose distractors come from the reviewed starter pool.
type rankedTargets struct {
	words       []Word
	recognition map[string]bool
	pool        []Word
}

// cueHidesAnswer matches whole English words and phrases across punctuation.
func cueHidesAnswer(cue, answer string) bool {
	words := func(text string) string {
		return strings.Join(strings.FieldsFunc(normalize(text), func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsMark(r)
		}), " ")
	}
	cue, answer = words(cue), words(answer)
	return cue != "" && answer != "" && !strings.Contains(" "+cue+" ", " "+answer+" ")
}

// recognitionStep freezes one meaning-choice question for a ranked target. The
// word's own pronunciation is attached, so the step never claims missing audio.
func recognitionStep(w Word, pool []Word, seed string) (privateStep, error) {
	step, err := journeyChoice(w, pool, seed, "meaning_choice")
	if err != nil {
		return step, err
	}
	step.Public.Prompt = "Choose the meaning of this word."
	step.Public.SpeechText = pronunciationText(w)
	step.Public.SpeechLanguage = "en-US"
	return step, nil
}

// loadRankedTargets loads only safely practisable ranked meanings among the
// unresolved sources. Unsupported senses are omitted, so availability never
// counts a mistake that a session could not include.
func loadRankedTargets(ctx context.Context, tx *sql.Tx, sources map[string]Source, seed string) (*rankedTargets, error) {
	ids := []string{}
	for id := range sources {
		if _, starter := referenceFor(id); !starter {
			ids = append(ids, id)
		}
	}
	out := &rankedTargets{words: []Word{}, recognition: map[string]bool{}}
	if len(ids) == 0 {
		return out, nil
	}
	pool, err := loadWords(ctx, tx)
	if errors.Is(err, ErrContentUnavailable) {
		// Without the reviewed distractor pool no ranked target is safe.
		// Starter mistakes stay available and List keeps working.
		return out, nil
	}
	if err != nil {
		return nil, err
	}
	out.pool = pool
	rows, err := tx.QueryContext(ctx, `SELECT wm.id::text,cw.id::text,cw.text,wm.short_definition,wm.part_of_speech,cw.language_code FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id WHERE wm.id=ANY($1::uuid[]) AND wm.status='active' AND cw.status='active' ORDER BY wm.id`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var w Word
		var wordID, language string
		if err = rows.Scan(&w.MeaningID, &wordID, &w.WordText, &w.Definition, &w.PartOfSpeech, &language); err != nil {
			return nil, err
		}
		w.WordSlug = strings.ReplaceAll(strings.ToLower(w.WordText), " ", "-")
		if language != "en" || strings.TrimSpace(w.Definition) == "" || strings.TrimSpace(w.WordText) == "" {
			continue
		}
		curated := false
		if SupportsListMeaning(w.MeaningID, wordID, w.WordText, w.PartOfSpeech, w.Definition, language) {
			for _, entry := range meaninguses.Entries() {
				if entry.MeaningID.String() == w.MeaningID {
					w.Example = entry.Example
					curated = true
					break
				}
			}
		}
		if curated && cueHidesAnswer(recallDefinition(w), w.WordText) {
			out.words = append(out.words, w)
			continue
		}
		// A definition that spells its own headword would make recognition
		// trivial, so such a sense is not offered as a graded mistake.
		if !cueHidesAnswer(w.Definition, w.WordText) {
			continue
		}
		if _, err := recognitionStep(w, pool, seed); err != nil {
			continue
		}
		out.recognition[w.MeaningID] = true
		out.words = append(out.words, w)
	}
	return out, rows.Err()
}

// supportedMistakes keeps starter sources and safely practisable ranked ones.
func supportedMistakes(sources map[string]Source, ranked *rankedTargets) map[string]Source {
	out := map[string]Source{}
	for id, src := range sources {
		if _, starter := referenceFor(id); starter {
			out[id] = src
		}
	}
	for _, w := range ranked.words {
		if src, ok := sources[w.MeaningID]; ok {
			out[w.MeaningID] = src
		}
	}
	return out
}
