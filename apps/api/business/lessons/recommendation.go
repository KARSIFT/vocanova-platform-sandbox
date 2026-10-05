package lessons

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Recommendation reports actual target coverage, not a proficiency estimate.
// A saved target is useful unless it is mastered or explicitly self-reported known.
type Recommendation struct {
	Lesson            Summary `json:"lesson"`
	Reason            string  `json:"reason" enum:"resume,focus_and_useful_words,useful_words"`
	UsefulTargetCount int     `json:"usefulTargetCount"`
	TotalTargetCount  int     `json:"totalTargetCount"`
	MatchesFocus      bool    `json:"matchesFocus"`
}

type RecommendationResult struct {
	Status         string          `json:"status" enum:"recommended,no_unfinished_lessons,no_useful_targets,content_unavailable"`
	Recommendation *Recommendation `json:"recommendation"`
}

// RecommendationData is an internal read model. States are ordered by latest
// activity first; coverage spans all owned vocabulary, not a visible list page.
type RecommendationData struct {
	States          []State
	Focus           string
	Categories      map[string]string
	Available       map[string]bool
	KnownOrMastered map[string]bool
}

type RecommendationRepository interface {
	ReadRecommendation(context.Context, uuid.UUID) (RecommendationData, error)
}

type RecommendationService struct{ repo RecommendationRepository }

func NewRecommendationService(repo RecommendationRepository) *RecommendationService {
	return &RecommendationService{repo: repo}
}

func (s *RecommendationService) Get(ctx context.Context, userID uuid.UUID) (RecommendationResult, error) {
	if userID == uuid.Nil {
		return RecommendationResult{}, ErrNotFound
	}
	data, err := s.repo.ReadRecommendation(ctx, userID)
	if err != nil {
		return RecommendationResult{}, err
	}
	completed := map[string]bool{}
	for _, state := range data.States {
		if state.CompletedAt != nil {
			completed[state.Snapshot.Definition.Key] = true
			continue
		}
		p := project(state)
		d := state.Snapshot.Definition
		count := 0
		for _, word := range p.Words {
			if !data.KnownOrMastered[word.MeaningID] {
				count++
			}
		}
		return RecommendationResult{Status: "recommended", Recommendation: &Recommendation{
			Lesson: Summary{Key: d.Key, Version: p.LessonVersion, Title: p.Title, SituationSlug: p.SituationSlug, SituationTitle: d.SituationTitle, Description: d.Description, WordCount: len(p.Words), StepCount: p.TotalSteps, Status: p.Status, SessionID: p.ID, CompletedSteps: p.CompletedSteps},
			Reason: "resume", UsefulTargetCount: count, TotalTargetCount: len(p.Words), MatchesFocus: data.Focus != "" && data.Categories[d.SituationSlug] == data.Focus,
		}}, nil
	}
	var best *Recommendation
	unfinished, unavailable := 0, 0
	for _, d := range catalog {
		if completed[d.Key] {
			continue
		}
		unfinished++
		if !data.Available[d.Key] {
			unavailable++
			continue
		}
		count := 0
		for _, word := range d.Words {
			if !data.KnownOrMastered[word.MeaningID] {
				count++
			}
		}
		if count == 0 {
			continue
		}
		matches := data.Focus != "" && data.Categories[d.SituationSlug] == data.Focus
		if best != nil && ((!matches && best.MatchesFocus) || (matches == best.MatchesFocus && count <= best.UsefulTargetCount)) {
			continue
		}
		reason := "useful_words"
		if matches {
			reason = "focus_and_useful_words"
		}
		best = &Recommendation{Lesson: Summary{Key: d.Key, Version: d.Version, Title: d.Title, SituationSlug: d.SituationSlug, SituationTitle: d.SituationTitle, Description: d.Description, WordCount: len(d.Words), StepCount: len(d.Words) * 3, Status: "not_started"}, Reason: reason, UsefulTargetCount: count, TotalTargetCount: len(d.Words), MatchesFocus: matches}
	}
	status := "no_useful_targets"
	if best != nil {
		status = "recommended"
	} else if unfinished == 0 {
		status = "no_unfinished_lessons"
	} else if unavailable > 0 {
		// Missing content is not proof that all unfinished targets are known.
		status = "content_unavailable"
	}
	return RecommendationResult{Status: status, Recommendation: best}, nil
}

// ReadRecommendation uses one read-only snapshot, so concurrently changing
// preferences, lessons and vocabulary cannot produce mixed coverage counts.
func (r *PostgreSQLRepository) ReadRecommendation(ctx context.Context, userID uuid.UUID) (RecommendationData, error) {
	data := RecommendationData{Categories: map[string]string{}, Available: map[string]bool{}, KnownOrMastered: map[string]bool{}}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return data, err
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(p.main_use_case,o.main_use_case,'') FROM users u
 LEFT JOIN user_learning_preferences p ON p.user_id=u.id
 LEFT JOIN user_onboarding_profiles o ON o.user_id=u.id
 WHERE u.id=$1 AND u.status='active' AND u.deleted_at IS NULL`, userID).Scan(&data.Focus)
	if errors.Is(err, sql.ErrNoRows) {
		return data, ErrNotFound
	}
	if err != nil {
		return data, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT `+stateColumns+` FROM lesson_sessions WHERE user_id=$1 ORDER BY updated_at DESC,id`, userID)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		state, e := scanState(rows)
		if e != nil {
			rows.Close()
			return data, e
		}
		data.States = append(data.States, *state)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return data, err
	}
	rows, err = tx.QueryContext(ctx, `SELECT meaning_id FROM user_word_knowledge WHERE user_id=$1 AND self_reported_known
 UNION SELECT meaning_id FROM user_words WHERE user_id=$1 AND deleted_at IS NULL AND status='mastered'`, userID)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return data, err
		}
		data.KnownOrMastered[id] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return data, err
	}
	rows, err = tx.QueryContext(ctx, `SELECT slug,category FROM journey_situations WHERE status='active'`)
	if err != nil {
		return data, err
	}
	for rows.Next() {
		var slug, category string
		if err = rows.Scan(&slug, &category); err != nil {
			rows.Close()
			return data, err
		}
		data.Categories[slug] = category
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return data, err
	}
	content, err := loadRecommendationWords(ctx, tx)
	if err != nil {
		return data, err
	}
	for _, d := range catalog {
		words := make([]Word, 0, len(d.Words))
		for _, ref := range d.Words {
			if word, ok := content[d.SituationSlug][ref.MeaningID]; ok {
				words = append(words, word)
			}
		}
		if len(words) != len(d.Words) {
			continue
		}
		// Apply the same canonical identity/completeness guard as starting a lesson.
		_, e := buildSnapshot(d, words, "recommendation")
		data.Available[d.Key] = e == nil
	}
	return data, tx.Commit()
}

// Read the catalog's canonical targets once, retaining each active situation
// link so a meaning linked elsewhere cannot make a lesson falsely available.
func loadRecommendationWords(ctx context.Context, tx *sql.Tx) (map[string]map[string]Word, error) {
	ids, slugs := []string{}, []string{}
	seenIDs, seenSlugs := map[string]bool{}, map[string]bool{}
	for _, d := range catalog {
		if !seenSlugs[d.SituationSlug] {
			slugs = append(slugs, d.SituationSlug)
			seenSlugs[d.SituationSlug] = true
		}
		for _, ref := range d.Words {
			if !seenIDs[ref.MeaningID] {
				ids = append(ids, ref.MeaningID)
				seenIDs[ref.MeaningID] = true
			}
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT js.slug,wm.id,cw.text,wm.part_of_speech,wm.short_definition,
 COALESCE((SELECT example_text FROM word_examples WHERE meaning_id=wm.id AND status='active' ORDER BY example_order,id LIMIT 1),''),
 COALESCE((SELECT note_text FROM usage_notes WHERE meaning_id=wm.id AND status='active' ORDER BY note_order,id LIMIT 1),'')
 FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id
 JOIN journey_words jw ON jw.meaning_id=wm.id
 JOIN journey_situations js ON js.id=jw.journey_situation_id
 WHERE wm.id=ANY($1::uuid[]) AND wm.status='active' AND cw.status='active'
 AND js.slug=ANY($2::text[]) AND js.status='active'`, pq.Array(ids), pq.Array(slugs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	content := map[string]map[string]Word{}
	for rows.Next() {
		var slug string
		var word Word
		if err := rows.Scan(&slug, &word.MeaningID, &word.WordText, &word.PartOfSpeech, &word.Definition, &word.Example, &word.UsageNote); err != nil {
			return nil, err
		}
		word.WordSlug = strings.ReplaceAll(strings.ToLower(word.WordText), " ", "-")
		if content[slug] == nil {
			content[slug] = map[string]Word{}
		}
		content[slug][word.MeaningID] = word
	}
	return content, rows.Err()
}
