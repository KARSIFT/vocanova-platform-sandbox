package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"strings"
	"time"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db} }

const columns = `s.id,s.user_id,s.snapshot,s.current_step,s.revision,s.feedback,s.first_answers_correct,s.questions_answered,s.created_at,s.updated_at,s.completed_at`

type scanner interface{ Scan(...any) error }

func scan(row scanner) (*State, error) {
	var st State
	var snapshot, feedback []byte
	err := row.Scan(&st.ID, &st.UserID, &snapshot, &st.Index, &st.Revision, &feedback, &st.FirstAnswersCorrect, &st.QuestionsAnswered, &st.CreatedAt, &st.UpdatedAt, &st.CompletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(snapshot, &st.Snapshot); err != nil {
		return nil, err
	}
	if len(feedback) > 0 {
		if err = json.Unmarshal(feedback, &st.Feedback); err != nil {
			return nil, err
		}
	}
	// Content is frozen in the saved snapshot. The expanded course must not
	// invalidate sessions created with the reviewed original 21-word catalog.
	supportedContent := st.Snapshot.ContentVersion == "starter-21-v1" || st.Snapshot.ContentVersion == ContentVersion
	if len(st.Snapshot.Steps) == 0 || st.Index < 0 || st.Index > len(st.Snapshot.Steps) || !supportedContent || st.Snapshot.GradingVersion != GradingVersion {
		return nil, ErrContentUnavailable
	}
	return &st, nil
}
func (r *PostgreSQLRepository) Get(ctx context.Context, u, id uuid.UUID) (*State, error) {
	return scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM practice_sessions s JOIN users u ON u.id=s.user_id WHERE s.id=$1 AND s.user_id=$2 AND u.status='active' AND u.deleted_at IS NULL`, id, u))
}
func lockUser(ctx context.Context, tx *sql.Tx, u uuid.UUID) error {
	var id uuid.UUID
	err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL FOR UPDATE`, u).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func mistakes(ctx context.Context, q queryer, u uuid.UUID) (map[string]Source, error) {
	ids := make([]string, len(catalog))
	for i, r := range catalog {
		ids[i] = r.MeaningID
	}
	// Rank before removing resolutions. Resolving the latest mistake must not
	// resurrect older mistakes; resolving an old snapshot cannot erase a new one.
	// The frozen step identifies the target, including typed answers without a
	// choice ID. Retain feedback fallback for historical actions/snapshots.
	rows, err := q.QueryContext(ctx, `WITH events AS (
 SELECT 'lesson'::text kind,la.id,
 COALESCE((SELECT NULLIF(step->'step'->'word'->>'meaningId','')
   FROM jsonb_array_elements(ls.snapshot->'steps') step
   WHERE step->'step'->>'id'=la.action->>'stepId' LIMIT 1),
   NULLIF(la.result->'feedback'->>'correctChoiceId','')) meaning,la.created_at
 FROM lesson_actions la JOIN lesson_sessions ls ON ls.id=la.session_id AND ls.user_id=la.user_id
 WHERE la.user_id=$1 AND la.operation='action' AND la.action->>'action'='answer' AND la.result->'feedback'->>'correct'='false'
 UNION ALL SELECT 'review',ra.id,ra.meaning_id::text,ra.created_at FROM review_attempts ra WHERE ra.user_id=$1 AND ra.prompt_type='multiple_choice' AND ra.result='incorrect'
 UNION ALL SELECT 'practice',pa.id,pa.meaning_id::text,pa.created_at FROM practice_actions pa WHERE pa.user_id=$1 AND pa.correct=false AND pa.action->>'action'='answer'
 ),ranked AS(SELECT *,row_number() OVER(PARTITION BY meaning ORDER BY created_at DESC,id DESC,kind) rn FROM events WHERE meaning=ANY($2::text[]))
 SELECT e.kind,e.id,e.meaning,e.created_at FROM ranked e JOIN word_meanings wm ON wm.id::text=e.meaning JOIN canonical_words cw ON cw.id=wm.word_id JOIN users u ON u.id=$1
 WHERE rn=1 AND wm.status='active' AND cw.status='active' AND u.status='active' AND u.deleted_at IS NULL AND NOT EXISTS(SELECT 1 FROM practice_mistake_resolutions r WHERE r.user_id=$1 AND r.source_kind=e.kind AND r.source_id=e.id)`, u, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]Source{}
	for rows.Next() {
		var s Source
		var m string
		if err = rows.Scan(&s.Kind, &s.ID, &m, &s.CreatedAt); err != nil {
			return nil, err
		}
		out[m] = s
	}
	return out, rows.Err()
}
func (r *PostgreSQLRepository) List(ctx context.Context, u uuid.UUID) ([]State, int, error) {
	sources, err := mistakes(ctx, r.db, u)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM practice_sessions s JOIN users u ON u.id=s.user_id WHERE s.user_id=$1 AND u.status='active' AND u.deleted_at IS NULL ORDER BY (s.status='in_progress') DESC,s.updated_at DESC,s.id LIMIT 20`, u)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []State{}
	for rows.Next() {
		st, e := scan(rows)
		if e != nil {
			return nil, 0, e
		}
		out = append(out, *st)
	}
	return out, len(sources), rows.Err()
}
func loadWords(ctx context.Context, tx *sql.Tx) ([]Word, error) {
	return loadReferences(ctx, tx, catalog)
}
func loadReferences(ctx context.Context, tx *sql.Tx, refs []reference) ([]Word, error) {
	ids := make([]string, len(refs))
	for i, r := range refs {
		ids[i] = r.MeaningID
	}
	rows, err := tx.QueryContext(ctx, `SELECT wm.id,cw.text,wm.short_definition FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id WHERE wm.id=ANY($1::uuid[]) AND wm.status='active' AND cw.status='active'`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Word{}
	for rows.Next() {
		var w Word
		if err = rows.Scan(&w.MeaningID, &w.WordText, &w.Definition); err != nil {
			return nil, err
		}
		ref, ok := referenceFor(w.MeaningID)
		if !ok || w.WordText != ref.WordText || w.Definition == "" {
			return nil, ErrContentUnavailable
		}
		w.LessonKey = ref.LessonKey
		w.WordSlug = strings.ReplaceAll(strings.ToLower(w.WordText), " ", "-")
		out = append(out, w)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(out) != len(refs) {
		return nil, ErrContentUnavailable
	}
	return out, nil
}
func (r *PostgreSQLRepository) Start(ctx context.Context, u uuid.UUID, req StartRequest, key string, now time.Time) (*State, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = lockUser(ctx, tx, u); err != nil {
		return nil, err
	}
	fp := fingerprint(req)
	var prior string
	var id uuid.UUID
	err = tx.QueryRowContext(ctx, `SELECT session_id,fingerprint FROM practice_actions WHERE user_id=$1 AND operation='start' AND idempotency_key=$2`, u, key).Scan(&id, &prior)
	if err == nil {
		if prior != fp {
			return nil, ErrConflict
		}
		st, e := scan(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM practice_sessions s WHERE s.id=$1 AND s.user_id=$2`, id, u))
		if e != nil {
			return nil, e
		}
		return st, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	sources := map[string]Source{}
	if req.Mode == "mistakes" {
		sources, err = mistakes(ctx, tx, u)
		if err != nil {
			return nil, err
		}
	}
	listName := ""
	if req.ListID != "" {
		var revision int
		err = tx.QueryRowContext(ctx, `SELECT name,revision FROM user_word_lists WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL FOR UPDATE`, req.ListID, u).Scan(&listName, &revision)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		if revision != *req.ListRevision {
			return nil, ErrConflict
		}
		rows, e := tx.QueryContext(ctx, `SELECT m.meaning_id::text,cw.text,wm.short_definition FROM user_word_list_members m JOIN word_meanings wm ON wm.id=m.meaning_id JOIN canonical_words cw ON cw.id=wm.word_id WHERE m.list_id=$1 AND wm.status='active' AND cw.status='active'`, req.ListID)
		if e != nil {
			return nil, e
		}
		for rows.Next() {
			var meaning, text, definition string
			if e = rows.Scan(&meaning, &text, &definition); e != nil {
				rows.Close()
				return nil, e
			}
			if SupportsMeaning(meaning, text, definition) {
				sources[meaning] = Source{}
			}
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return nil, e
		}
	}
	refs := catalog
	if req.ListID != "" {
		refs = []reference{}
		lessons := map[string]bool{}
		for _, ref := range catalog {
			if _, ok := sources[ref.MeaningID]; ok {
				lessons[ref.LessonKey] = true
			}
		}
		for _, ref := range catalog {
			_, member := sources[ref.MeaningID]
			if member || (req.Mode == "listening_choice" && lessons[ref.LessonKey]) {
				refs = append(refs, ref)
			}
		}
		if len(refs) == 0 {
			return nil, ErrListEmpty
		}
	}
	words, err := loadReferences(ctx, tx, refs)
	if err != nil {
		return nil, err
	}
	id = uuid.New()
	snap, err := build(req, words, sources, id.String())
	if err != nil {
		return nil, err
	}
	if req.ListID != "" {
		snap.ListID = req.ListID
		snap.ListName = listName
		revision := *req.ListRevision
		snap.ListRevision = &revision
	}
	st := &State{ID: id, UserID: u, Snapshot: snap, CreatedAt: now, UpdatedAt: now}
	raw, err := json.Marshal(snap)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO practice_sessions(id,user_id,mode,lesson_key,snapshot,total_steps,created_at,updated_at)VALUES($1,$2,$3,$4,$5,$6,$7,$7)`, id, u, req.Mode, req.LessonKey, string(raw), len(snap.Steps), now)
	if err != nil {
		return nil, err
	}
	if err = record(ctx, tx, st, "start", key, key, fp, req, nil, nil, now); err != nil {
		return nil, err
	}
	return st, tx.Commit()
}
func (r *PostgreSQLRepository) Act(ctx context.Context, u, id uuid.UUID, a Action, key string, now time.Time) (*State, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = lockUser(ctx, tx, u); err != nil {
		return nil, err
	}
	st, err := scan(tx.QueryRowContext(ctx, `SELECT `+columns+` FROM practice_sessions s WHERE s.id=$1 AND s.user_id=$2 FOR UPDATE`, id, u))
	if err != nil {
		return nil, err
	}
	fp := fingerprint(struct {
		ID     uuid.UUID
		Action Action
	}{id, a})
	rows, err := tx.QueryContext(ctx, `SELECT session_id,idempotency_key,client_action_id,fingerprint FROM practice_actions WHERE user_id=$1 AND operation='action' AND(idempotency_key=$2 OR(session_id=$3 AND client_action_id=$4))`, u, key, id, a.ClientActionID)
	if err != nil {
		return nil, err
	}
	replay := false
	for rows.Next() {
		var sid uuid.UUID
		var k, client, prior string
		if err = rows.Scan(&sid, &k, &client, &prior); err != nil {
			rows.Close()
			return nil, err
		}
		if sid != id || k != key || client != a.ClientActionID || prior != fp {
			rows.Close()
			return nil, ErrConflict
		}
		replay = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if replay {
		return st, tx.Commit()
	}
	var meaning any
	var correct any
	var source *Source
	if st.Index < len(st.Snapshot.Steps) {
		step := st.Snapshot.Steps[st.Index]
		meaning = step.Word.MeaningID
		source = step.Source
	}
	resolved, err := apply(st, a, now)
	if err != nil {
		return nil, err
	}
	if a.Action == "answer" {
		correct = st.Feedback.Correct
	}
	var feedback any
	if st.Feedback != nil {
		raw, e := json.Marshal(st.Feedback)
		if e != nil {
			return nil, e
		}
		feedback = string(raw)
	}
	status := "in_progress"
	if st.CompletedAt != nil {
		status = "completed"
	}
	_, err = tx.ExecContext(ctx, `UPDATE practice_sessions SET current_step=$2,revision=$3,feedback=$4,first_answers_correct=$5,questions_answered=$6,status=$7,completed_at=$8,updated_at=$9 WHERE id=$1`, id, st.Index, st.Revision, feedback, st.FirstAnswersCorrect, st.QuestionsAnswered, status, st.CompletedAt, now)
	if err != nil {
		return nil, err
	}
	if err = record(ctx, tx, st, "action", key, a.ClientActionID, fp, a, meaning, correct, now); err != nil {
		return nil, err
	}
	if resolved && source != nil {
		_, err = tx.ExecContext(ctx, `INSERT INTO practice_mistake_resolutions(user_id,source_kind,source_id,meaning_id,session_id,resolved_at)VALUES($1,$2,$3,$4,$5,$6)ON CONFLICT DO NOTHING`, u, source.Kind, source.ID, meaning, id, now)
		if err != nil {
			return nil, err
		}
	}
	return st, tx.Commit()
}
func record(ctx context.Context, tx *sql.Tx, st *State, operation, key, client, fp string, action, meaning, correct any, now time.Time) error {
	body, err := json.Marshal(action)
	if err != nil {
		return err
	}
	result, err := json.Marshal(project(*st))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO practice_actions(id,session_id,user_id,operation,idempotency_key,client_action_id,fingerprint,action,result,meaning_id,correct,created_at)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, uuid.New(), st.ID, st.UserID, operation, key, client, fp, string(body), string(result), meaning, correct, now)
	return err
}
