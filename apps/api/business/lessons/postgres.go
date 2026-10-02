package lessons

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

const stateColumns = `id,user_id,snapshot,current_step,revision,feedback,first_answers_correct,questions_answered,completed_at`

type scanner interface{ Scan(...any) error }

func scanState(row scanner) (*State, error) {
	var st State
	var snapshot, feedback []byte
	err := row.Scan(&st.ID, &st.UserID, &snapshot, &st.Index, &st.Revision, &feedback, &st.FirstAnswersCorrect, &st.QuestionsAnswered, &st.CompletedAt)
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
	if len(st.Snapshot.Steps) == 0 || st.Index < 0 || st.Index > len(st.Snapshot.Steps) {
		return nil, errors.New("invalid stored lesson state")
	}
	return &st, nil
}
func (r *PostgreSQLRepository) List(ctx context.Context, u uuid.UUID) ([]State, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+stateColumns+` FROM lesson_sessions WHERE user_id=$1 ORDER BY created_at,id`, u)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []State{}
	for rows.Next() {
		st, e := scanState(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *st)
	}
	return out, rows.Err()
}
func (r *PostgreSQLRepository) Get(ctx context.Context, u, id uuid.UUID) (*State, error) {
	return scanState(r.db.QueryRowContext(ctx, `SELECT `+stateColumns+` FROM lesson_sessions WHERE id=$1 AND user_id=$2`, id, u))
}
func lockUser(ctx context.Context, tx *sql.Tx, u uuid.UUID) error {
	var id uuid.UUID
	err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL FOR UPDATE`, u).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
func (r *PostgreSQLRepository) Start(ctx context.Context, u uuid.UUID, d Definition, key string, now time.Time) (*State, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = lockUser(ctx, tx, u); err != nil {
		return nil, err
	}
	fp := fingerprint(d.Key)
	var existingID uuid.UUID
	var existingFP string
	err = tx.QueryRowContext(ctx, `SELECT session_id,fingerprint FROM lesson_actions WHERE user_id=$1 AND operation='start' AND idempotency_key=$2`, u, key).Scan(&existingID, &existingFP)
	if err == nil {
		if existingFP != fp {
			return nil, ErrConflict
		}
		st, e := scanState(tx.QueryRowContext(ctx, `SELECT `+stateColumns+` FROM lesson_sessions WHERE id=$1 AND user_id=$2`, existingID, u))
		if e != nil {
			return nil, e
		}
		return st, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	st, err := scanState(tx.QueryRowContext(ctx, `SELECT `+stateColumns+` FROM lesson_sessions WHERE user_id=$1 AND lesson_key=$2`, u, d.Key))
	if errors.Is(err, ErrNotFound) {
		words, e := loadWords(ctx, tx, d)
		if e != nil {
			return nil, e
		}
		id := uuid.New()
		snapshot, e := buildSnapshot(d, words, id.String())
		if e != nil {
			return nil, e
		}
		st = &State{ID: id, UserID: u, Snapshot: snapshot}
		raw, e := json.Marshal(snapshot)
		if e != nil {
			return nil, e
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO lesson_sessions(id,user_id,lesson_key,lesson_version,snapshot,total_steps,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$7)`, id, u, d.Key, d.Version, string(raw), len(snapshot.Steps), now)
	} else if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if err = record(ctx, tx, st, "start", key, key, fp, map[string]string{"lessonKey": d.Key}, now); err != nil {
		return nil, err
	}
	return st, tx.Commit()
}
func loadWords(ctx context.Context, tx *sql.Tx, d Definition) ([]Word, error) {
	ids := make([]string, len(d.Words))
	for i, w := range d.Words {
		ids[i] = w.MeaningID
	}
	rows, err := tx.QueryContext(ctx, `SELECT wm.id,cw.text,wm.part_of_speech,wm.short_definition,
 COALESCE((SELECT example_text FROM word_examples WHERE meaning_id=wm.id AND status='active' ORDER BY example_order,id LIMIT 1),''),
 COALESCE((SELECT note_text FROM usage_notes WHERE meaning_id=wm.id AND status='active' ORDER BY note_order,id LIMIT 1),'')
 FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id
 WHERE wm.id=ANY($1::uuid[]) AND wm.status='active' AND cw.status='active'
 AND EXISTS(SELECT 1 FROM journey_words jw JOIN journey_situations js ON js.id=jw.journey_situation_id WHERE jw.meaning_id=wm.id AND js.slug=$2 AND js.status='active')`, pq.Array(ids), d.SituationSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byID := map[string]Word{}
	for rows.Next() {
		var w Word
		if err = rows.Scan(&w.MeaningID, &w.WordText, &w.PartOfSpeech, &w.Definition, &w.Example, &w.UsageNote); err != nil {
			return nil, err
		}
		w.WordSlug = strings.ReplaceAll(strings.ToLower(w.WordText), " ", "-")
		byID[w.MeaningID] = w
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	words := make([]Word, len(d.Words))
	for i, ref := range d.Words {
		w, ok := byID[ref.MeaningID]
		if !ok {
			return nil, ErrContentUnavailable
		}
		words[i] = w
	}
	return words, nil
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
	st, err := scanState(tx.QueryRowContext(ctx, `SELECT `+stateColumns+` FROM lesson_sessions WHERE id=$1 AND user_id=$2 FOR UPDATE`, id, u))
	if err != nil {
		return nil, err
	}
	fp := fingerprint(struct {
		Session string
		Action  Action
	}{id.String(), a})
	rows, err := tx.QueryContext(ctx, `SELECT session_id,idempotency_key,client_action_id,fingerprint FROM lesson_actions WHERE user_id=$1 AND operation='action' AND (idempotency_key=$2 OR (session_id=$3 AND client_action_id=$4))`, u, key, id, a.ClientActionID)
	if err != nil {
		return nil, err
	}
	replay := false
	for rows.Next() {
		var sid uuid.UUID
		var oldKey, client, oldFP string
		if e := rows.Scan(&sid, &oldKey, &client, &oldFP); e != nil {
			rows.Close()
			return nil, e
		}
		if sid != id || oldKey != key || client != a.ClientActionID || oldFP != fp {
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
	// Return CURRENT state on a replay, never rewind to the old receipt revision.
	if replay {
		return st, tx.Commit()
	}
	if err = apply(st, a, now); err != nil {
		return nil, err
	}
	var feedback any
	if st.Feedback != nil {
		b, e := json.Marshal(st.Feedback)
		if e != nil {
			return nil, e
		}
		feedback = string(b)
	}
	status := "in_progress"
	if st.CompletedAt != nil {
		status = "completed"
	}
	_, err = tx.ExecContext(ctx, `UPDATE lesson_sessions SET current_step=$2,revision=$3,feedback=$4,first_answers_correct=$5,questions_answered=$6,status=$7,completed_at=$8,updated_at=$9 WHERE id=$1`, id, st.Index, st.Revision, feedback, st.FirstAnswersCorrect, st.QuestionsAnswered, status, st.CompletedAt, now)
	if err != nil {
		return nil, err
	}
	if err = record(ctx, tx, st, "action", key, a.ClientActionID, fp, a, now); err != nil {
		return nil, err
	}
	return st, tx.Commit()
}
func record(ctx context.Context, tx *sql.Tx, st *State, operation, key, client, fp string, action any, now time.Time) error {
	body, err := json.Marshal(action)
	if err != nil {
		return err
	}
	result, err := json.Marshal(project(*st))
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO lesson_actions(id,session_id,user_id,operation,idempotency_key,client_action_id,fingerprint,action,result,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, uuid.New(), st.ID, st.UserID, operation, key, client, fp, string(body), string(result), now)
	if err != nil {
		return fmt.Errorf("record lesson action: %w", err)
	}
	return nil
}
