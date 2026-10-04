package stories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"time"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db} }

const columns = "s.id,s.user_id,s.snapshot,s.current_step,s.revision,s.feedback,s.first_answers_correct,s.questions_answered,s.created_at,s.updated_at,s.completed_at"

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
	if st.Snapshot.ContentVersion != "original-dialogues-v1" || st.Snapshot.GradingVersion != "curated-choice-v1" || len(st.Snapshot.Steps) == 0 || st.Index < 0 || st.Index > len(st.Snapshot.Steps) {
		return nil, ErrContentUnavailable
	}
	// Supported historical snapshots stay valid even when catalog text changes.
	return &st, nil
}
func (r *PostgreSQLRepository) Active(ctx context.Context, u uuid.UUID) error {
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, "SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL", u).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
func lockUser(ctx context.Context, tx *sql.Tx, u uuid.UUID) error {
	var id uuid.UUID
	err := tx.QueryRowContext(ctx, "SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL FOR UPDATE", u).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
func (r *PostgreSQLRepository) List(ctx context.Context, u uuid.UUID) ([]State, error) {
	if err := r.Active(ctx, u); err != nil {
		return nil, err
	}
	// Select one newest attempt per story in PostgreSQL before loading any
	// snapshots. Repeated attempts cannot grow this response or decoding work.
	rows, err := r.db.QueryContext(ctx, "SELECT DISTINCT ON (s.story_key) "+columns+" FROM story_sessions s JOIN users u ON u.id=s.user_id WHERE s.user_id=$1 AND u.status='active' AND u.deleted_at IS NULL ORDER BY s.story_key,s.updated_at DESC,s.id DESC", u)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []State{}
	for rows.Next() {
		st, e := scan(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *st)
	}
	return out, rows.Err()
}
func (r *PostgreSQLRepository) Get(ctx context.Context, u, id uuid.UUID) (*State, error) {
	return scan(r.db.QueryRowContext(ctx, "SELECT "+columns+" FROM story_sessions s JOIN users u ON u.id=s.user_id WHERE s.id=$1 AND s.user_id=$2 AND u.status='active' AND u.deleted_at IS NULL", id, u))
}
func (r *PostgreSQLRepository) Start(ctx context.Context, u uuid.UUID, req StoryStartRequest, key string, now time.Time) (*State, error) {
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
	err = tx.QueryRowContext(ctx, "SELECT session_id,fingerprint FROM story_actions WHERE user_id=$1 AND operation='start' AND idempotency_key=$2", u, key).Scan(&id, &prior)
	if err == nil {
		if prior != fp {
			return nil, ErrConflict
		}
		st, e := scan(tx.QueryRowContext(ctx, "SELECT "+columns+" FROM story_sessions s WHERE s.id=$1 AND s.user_id=$2", id, u))
		if e != nil {
			return nil, e
		}
		return st, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	snap, err := build(req.StoryKey)
	if err != nil {
		return nil, err
	}
	id = uuid.New()
	st := &State{ID: id, UserID: u, Snapshot: snap, CreatedAt: now, UpdatedAt: now}
	raw, err := json.Marshal(snap)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO story_sessions(id,user_id,story_key,content_version,grading_version,snapshot,total_steps,created_at,updated_at)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$8)", id, u, req.StoryKey, ContentVersion, GradingVersion, string(raw), len(snap.Steps), now)
	if err != nil {
		return nil, err
	}
	if err = record(ctx, tx, st, "start", key, key, fp, req, now); err != nil {
		return nil, err
	}
	return st, tx.Commit()
}
func (r *PostgreSQLRepository) Act(ctx context.Context, u, id uuid.UUID, a StoryAction, key string, now time.Time) (*State, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = lockUser(ctx, tx, u); err != nil {
		return nil, err
	}
	st, err := scan(tx.QueryRowContext(ctx, "SELECT "+columns+" FROM story_sessions s WHERE s.id=$1 AND s.user_id=$2 FOR UPDATE", id, u))
	if err != nil {
		return nil, err
	}
	fp := fingerprint(struct {
		ID     uuid.UUID
		Action StoryAction
	}{id, a})
	rows, err := tx.QueryContext(ctx, "SELECT session_id,idempotency_key,client_action_id,fingerprint FROM story_actions WHERE user_id=$1 AND operation='action' AND(idempotency_key=$2 OR(session_id=$3 AND client_action_id=$4))", u, key, id, a.ClientActionID)
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
	if err = apply(st, a, now); err != nil {
		return nil, err
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
	_, err = tx.ExecContext(ctx, "UPDATE story_sessions SET current_step=$2,revision=$3,feedback=$4,first_answers_correct=$5,questions_answered=$6,status=$7,completed_at=$8,updated_at=$9 WHERE id=$1", id, st.Index, st.Revision, feedback, st.FirstAnswersCorrect, st.QuestionsAnswered, status, st.CompletedAt, now)
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
	_, err = tx.ExecContext(ctx, "INSERT INTO story_actions(id,session_id,user_id,operation,idempotency_key,client_action_id,fingerprint,action,result,created_at)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)", uuid.New(), st.ID, st.UserID, operation, key, client, fp, string(body), string(result), now)
	return err
}
