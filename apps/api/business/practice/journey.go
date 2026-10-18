package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaninguses"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const JourneyContentVersion = "personal-mixed-v1"

// Journey is confirmed practice progress, separate from SRS and daily reviews.
type Journey struct {
	CompletedLessons      int      `json:"completedLessons"`
	TodayCompletedLessons int      `json:"todayCompletedLessons"`
	DailyGoal             int      `json:"dailyGoal"`
	CurrentSession        *Summary `json:"currentSession"`
	NextLesson            int      `json:"nextLesson"`
	SupportedKnownWords   int      `json:"supportedKnownWords"`
	TotalKnownWords       int      `json:"totalKnownWords"`
}

// Completed practice targets are studied, not mastered or self-reported known.
// Saved but never studied words alone do not enter the familiar-word pool.
const familiarSQL = `WITH familiar AS (
 SELECT meaning_id FROM user_word_knowledge WHERE user_id=$1 AND self_reported_known
 UNION SELECT meaning_id FROM user_words WHERE user_id=$1 AND deleted_at IS NULL AND status NOT IN('ignored','archived') AND (correct_review_count>0 OR status='mastered')
 UNION SELECT (w->>'meaningId')::uuid FROM lesson_sessions s CROSS JOIN LATERAL jsonb_array_elements(COALESCE(s.snapshot->'words','[]'::jsonb)) w WHERE s.user_id=$1 AND s.status='completed' AND w->>'meaningId' IS NOT NULL
 UNION SELECT (w->'Word'->>'MeaningID')::uuid FROM practice_sessions s CROSS JOIN LATERAL jsonb_array_elements(COALESCE(s.snapshot->'Steps','[]'::jsonb)) w WHERE s.user_id=$1 AND s.mode='journey' AND s.status='completed' AND w->'Word'->>'MeaningID' IS NOT NULL
 ) `

func (r *PostgreSQLRepository) Journey(ctx context.Context, u uuid.UUID, now time.Time) (*Journey, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL)`, u).Scan(&active); err != nil {
		return nil, err
	}
	if !active {
		return nil, ErrNotFound
	}
	out := &Journey{DailyGoal: 3}
	if err = tx.QueryRowContext(ctx, `SELECT count(*),count(*) FILTER(WHERE (completed_at AT TIME ZONE COALESCE((SELECT timezone FROM user_settings WHERE user_id=$1),'UTC'))::date=($2 AT TIME ZONE COALESCE((SELECT timezone FROM user_settings WHERE user_id=$1),'UTC'))::date) FROM practice_sessions WHERE user_id=$1 AND mode='journey' AND status='completed'`, u, now).Scan(&out.CompletedLessons, &out.TodayCompletedLessons); err != nil {
		return nil, err
	}
	out.NextLesson = out.CompletedLessons + 1
	if err = tx.QueryRowContext(ctx, familiarSQL+`SELECT count(*) FROM familiar f JOIN word_meanings wm ON wm.id=f.meaning_id JOIN canonical_words cw ON cw.id=wm.word_id WHERE wm.status='active' AND cw.status='active' AND wm.short_definition<>''`, u).Scan(&out.TotalKnownWords); err != nil {
		return nil, err
	}
	// Every active familiar definition supports meaning/listening. Authored
	// context exercises are added only where reviewed examples are available.
	out.SupportedKnownWords = out.TotalKnownWords
	st, err := scan(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM practice_sessions s WHERE user_id=$1 AND mode='journey' AND status='in_progress'`, u))
	if err == nil {
		summary := project(*st).Summary
		out.CurrentSession = &summary
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	return out, tx.Commit()
}

func (r *PostgreSQLRepository) startJourney(ctx context.Context, tx *sql.Tx, u uuid.UUID, req StartRequest, key, fp string, now time.Time) (*State, error) {
	st, err := scan(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM practice_sessions s WHERE user_id=$1 AND mode='journey' AND status='in_progress'`, u))
	if err == nil {
		if err = record(ctx, tx, st, "start", key, key, fp, req, nil, nil, now); err != nil {
			return nil, err
		}
		return st, tx.Commit()
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	var completed int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM practice_sessions WHERE user_id=$1 AND mode='journey' AND status='completed'`, u).Scan(&completed); err != nil {
		return nil, err
	}
	// Prefer the least recently practised familiar words, rather than always
	// selecting a learner's first page of saved words. Randomize only equal-recency
	// ties, once per frozen session, so all eligible uses have a fair turn.
	id := uuid.New()
	coreIDs := make([]string, len(catalog))
	for i, ref := range catalog {
		coreIDs[i] = ref.MeaningID
	}
	rows, err := tx.QueryContext(ctx, familiarSQL+`, candidates AS (
 SELECT wm.id,cw.text,wm.short_definition,wm.part_of_speech,wm.id=ANY($2::uuid[]) authored,
 (SELECT max(pa.created_at) FROM practice_actions pa WHERE pa.user_id=$1 AND pa.meaning_id=wm.id) last_practised
 FROM familiar f JOIN word_meanings wm ON wm.id=f.meaning_id JOIN canonical_words cw ON cw.id=wm.word_id
 WHERE wm.status='active' AND cw.status='active' AND wm.short_definition<>''
 ), ranked AS(SELECT *,row_number() OVER(PARTITION BY authored ORDER BY last_practised ASC NULLS FIRST,md5($3::text||id::text)) ordinal FROM candidates)
 SELECT id,text,short_definition,part_of_speech FROM ranked WHERE ordinal<=12 ORDER BY last_practised ASC NULLS FIRST,md5($3::text||id::text)`, u, pq.Array(coreIDs), id.String())
	if err != nil {
		return nil, err
	}
	known := []Word{}
	for rows.Next() {
		var w Word
		if err = rows.Scan(&w.MeaningID, &w.WordText, &w.Definition, &w.PartOfSpeech); err != nil {
			rows.Close()
			return nil, err
		}
		w.WordSlug = strings.ReplaceAll(strings.ToLower(w.WordText), " ", "-")
		known = append(known, w)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	// Distractors and new teaching targets come from the reviewed starter pool.
	pool, err := loadWords(ctx, tx)
	if err != nil {
		return nil, err
	}
	// Other active uses of a familiar spelling are eligible for explicit teaching,
	// not silently treated as known. Keep the pool bounded and never invent a POS.
	siblings, err := tx.QueryContext(ctx, familiarSQL+`SELECT DISTINCT wm.id,cw.text,wm.short_definition,wm.part_of_speech,md5($2::text||wm.id::text) tie
 FROM familiar f JOIN word_meanings familiar_meaning ON familiar_meaning.id=f.meaning_id
 JOIN word_meanings wm ON wm.word_id=familiar_meaning.word_id JOIN canonical_words cw ON cw.id=wm.word_id
 WHERE familiar_meaning.status='active' AND wm.status='active' AND cw.status='active' AND cw.language_code='en' AND wm.short_definition<>''
 AND NOT EXISTS(SELECT 1 FROM familiar existing WHERE existing.meaning_id=wm.id)
 ORDER BY tie LIMIT 24`, u, id.String())
	if err != nil {
		return nil, err
	}
	poolIDs := map[string]bool{}
	for _, w := range pool {
		poolIDs[w.MeaningID] = true
	}
	for siblings.Next() {
		var w Word
		var tie string
		if err = siblings.Scan(&w.MeaningID, &w.WordText, &w.Definition, &w.PartOfSpeech, &tie); err != nil {
			siblings.Close()
			return nil, err
		}
		if !poolIDs[w.MeaningID] {
			w.WordSlug = strings.ReplaceAll(strings.ToLower(w.WordText), " ", "-")
			pool = append(pool, w)
			poolIDs[w.MeaningID] = true
		}
	}
	err = siblings.Err()
	siblings.Close()
	if err != nil {
		return nil, err
	}
	examples, err := tx.QueryContext(ctx, `SELECT DISTINCT ON(meaning_id) meaning_id,example_text FROM word_examples WHERE status='active' ORDER BY meaning_id,example_order,id`)
	if err != nil {
		return nil, err
	}
	byExample := map[string]string{}
	reviewedExamples := map[string]string{}
	for _, entry := range meaninguses.Entries() {
		reviewedExamples[entry.MeaningID.String()] = entry.Example
	}
	for examples.Next() {
		var id, text string
		if err = examples.Scan(&id, &text); err != nil {
			examples.Close()
			return nil, err
		}
		if _, ok := referenceFor(id); ok || (reviewedExamples[id] != "" && reviewedExamples[id] == text) {
			byExample[id] = text
		}
	}
	err = examples.Err()
	examples.Close()
	if err != nil {
		return nil, err
	}
	for i := range known {
		known[i].Example = byExample[known[i].MeaningID]
	}
	for i := range pool {
		pool[i].Example = byExample[pool[i].MeaningID]
	}
	// Query all familiar identities when picking new targets, not just the 24
	// selected above. A familiar word must never be mislabeled as new.
	ids, err := tx.QueryContext(ctx, familiarSQL+`SELECT meaning_id FROM familiar`, u)
	if err != nil {
		return nil, err
	}
	familiar := map[string]bool{}
	for ids.Next() {
		var id string
		if err = ids.Scan(&id); err != nil {
			ids.Close()
			return nil, err
		}
		familiar[id] = true
	}
	err = ids.Err()
	ids.Close()
	if err != nil {
		return nil, err
	}
	snap, err := buildJourney(known, pool, familiar, completed+1, id.String())
	if err != nil {
		return nil, err
	}
	st = &State{ID: id, UserID: u, Snapshot: snap, CreatedAt: now, UpdatedAt: now}
	raw, err := json.Marshal(snap)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO practice_sessions(id,user_id,mode,lesson_key,snapshot,total_steps,created_at,updated_at)VALUES($1,$2,'journey','',$3,$4,$5,$5)`, id, u, string(raw), len(snap.Steps), now)
	if err != nil {
		return nil, err
	}
	if err = record(ctx, tx, st, "start", key, key, fp, req, nil, nil, now); err != nil {
		return nil, err
	}
	return st, tx.Commit()
}

func buildJourney(known, pool []Word, familiar map[string]bool, ordinal int, seed string) (Snapshot, error) {
	snap := Snapshot{Mode: "journey", ContentVersion: JourneyContentVersion, GradingVersion: GradingVersion, JourneyLesson: ordinal, Steps: []privateStep{}}
	selected := append([]Word{}, known...)
	if len(selected) > 2 {
		selected = selected[:2]
	}
	// Include real familiar authored context when available, alongside the next
	// general familiar word. Each group still rotates by actual last practice.
	var authored, general *Word
	for i := range known {
		_, safe := blankExample(known[i].Example, known[i].WordText)
		if safe && len(strings.Fields(known[i].Example)) <= 18 {
			if authored == nil {
				authored = &known[i]
			}
		} else if general == nil {
			general = &known[i]
		}
	}
	if authored != nil && general != nil {
		selected = []Word{*authored, *general}
	}
	newTargets := map[string]bool{}
	if len(selected) == 0 || ordinal%3 == 0 {
		candidates := append([]Word{}, pool...)
		familiarSpellings := map[string]bool{}
		for _, w := range known {
			familiarSpellings[normalize(w.WordText)] = true
		}
		sort.Slice(candidates, func(i, j int) bool {
			// At the normal teaching interval, introduce a different actual use of
			// a familiar word before unrelated new vocabulary. Never skip teaching.
			iSibling := familiarSpellings[normalize(candidates[i].WordText)]
			jSibling := familiarSpellings[normalize(candidates[j].WordText)]
			if iSibling != jSibling {
				return iSibling
			}
			return order(seed, candidates[i].MeaningID) < order(seed, candidates[j].MeaningID)
		})
		want := 1
		if len(selected) == 0 {
			want = 2
		}
		for _, w := range candidates {
			if familiar[w.MeaningID] || newTargets[w.MeaningID] {
				continue
			}
			if len(selected) == 2 {
				selected = selected[:1]
			}
			selected = append(selected, w)
			newTargets[w.MeaningID] = true
			want--
			if want == 0 {
				break
			}
		}
	}
	if len(selected) == 0 {
		return snap, ErrContentUnavailable
	}
	snap.NewWords = len(newTargets)
	for _, w := range selected {
		if newTargets[w.MeaningID] {
			snap.Steps = append(snap.Steps, privateStep{Word: w, Public: Step{ID: uuid.NewString(), Kind: "teach", WordText: w.WordText, Prompt: w.Definition, Context: w.Example, SpeechText: pronunciationText(w), SpeechLanguage: "en-US", Choices: []Choice{}}})
		}
		listen, err := journeyChoice(w, pool, seed, "listening_choice")
		if err != nil {
			return snap, err
		}
		snap.Steps = append(snap.Steps, listen)
		gap, ok := blankExample(w.Example, w.WordText)
		if ok {
			step, err := journeyChoice(w, pool, seed, "gap_choice")
			if err != nil {
				return snap, err
			}
			step.Public.Context = gap
			snap.Steps = append(snap.Steps, step)
		}
		tokens := strings.Fields(w.Example)
		if ok && len(tokens) >= 3 && len(tokens) <= 18 {
			step := privateStep{Word: w, Public: Step{ID: uuid.NewString(), Kind: "sentence_build", WordText: w.WordText, Prompt: "Build the sentence you hear.", SpeechText: w.Example, SpeechLanguage: "en-US", Choices: []Choice{}}}
			for _, token := range tokens {
				tid := uuid.NewString()
				step.CorrectTokens = append(step.CorrectTokens, tid)
				step.Public.Choices = append(step.Public.Choices, Choice{ID: tid, Text: token})
			}
			sort.Slice(step.Public.Choices, func(i, j int) bool {
				return order(seed+step.Public.ID, step.Public.Choices[i].ID) < order(seed+step.Public.ID, step.Public.Choices[j].ID)
			})
			snap.Steps = append(snap.Steps, step)
		} else {
			step, err := journeyChoice(w, pool, seed, "meaning_choice")
			if err != nil {
				return snap, err
			}
			snap.Steps = append(snap.Steps, step)
		}
	}
	return snap, nil
}

func blankExample(example, target string) (string, bool) {
	if example == "" || target == "" {
		return "", false
	}
	pattern, err := regexp.Compile(`(?i)\b` + regexp.QuoteMeta(target) + `\b`)
	if err != nil {
		return "", false
	}
	if !pattern.MatchString(example) {
		return "", false
	}
	return pattern.ReplaceAllString(example, "_____"), true
}

func pronunciationText(word Word) string {
	// Retain the reviewed noun alias for historical snapshots without POS, but
	// never turn the verb (to continue) into the CV noun pronunciation.
	if word.WordText == "resume" && (word.PartOfSpeech == "noun" || word.PartOfSpeech == "") {
		return "résumé"
	}
	return word.WordText
}

func journeyChoice(w Word, pool []Word, seed, kind string) (privateStep, error) {
	step := privateStep{Word: w, Public: Step{ID: uuid.NewString(), Kind: kind, Choices: []Choice{}}}
	step.Public.Prompt = "Choose the meaning."
	step.Public.WordText = w.WordText
	if kind == "listening_choice" {
		step.Public.WordText = ""
		step.Public.Prompt = "Listen and choose the meaning."
		step.Public.SpeechText = pronunciationText(w)
		step.Public.SpeechLanguage = "en-US"
	}
	if kind == "gap_choice" {
		step.Public.WordText = ""
		step.Public.Prompt = "Choose the missing word."
	}
	options := []Word{w}
	candidates := append([]Word{}, pool...)
	sort.Slice(candidates, func(i, j int) bool {
		return order(seed+w.MeaningID+kind, candidates[i].MeaningID) < order(seed+w.MeaningID+kind, candidates[j].MeaningID)
	})
	choiceText := func(candidate Word) string {
		if kind == "gap_choice" {
			return normalize(candidate.WordText)
		}
		return normalize(candidate.Definition)
	}
	seenText := map[string]bool{choiceText(w): true}
	for _, candidate := range candidates {
		if candidate.MeaningID == w.MeaningID || normalize(candidate.WordText) == normalize(w.WordText) || normalize(candidate.Definition) == normalize(w.Definition) {
			continue
		}
		if seenText[choiceText(candidate)] {
			continue
		}
		aliases, err := knowledgecatalog.SemanticAliases(w.WordText, candidate.WordText)
		if err != nil {
			return step, ErrContentUnavailable
		}
		if aliases {
			continue
		}
		options = append(options, candidate)
		seenText[choiceText(candidate)] = true
		if len(options) == 4 {
			break
		}
	}
	if len(options) < 3 {
		return step, ErrContentUnavailable
	}
	for _, candidate := range options {
		id := uuid.NewString()
		text := candidate.Definition
		if kind == "gap_choice" {
			text = candidate.WordText
		}
		step.Public.Choices = append(step.Public.Choices, Choice{ID: id, Text: text})
		if candidate.MeaningID == w.MeaningID {
			step.CorrectChoice = id
		}
	}
	sort.Slice(step.Public.Choices, func(i, j int) bool {
		return order(seed+step.Public.ID, step.Public.Choices[i].ID) < order(seed+step.Public.ID, step.Public.Choices[j].ID)
	})
	return step, nil
}
