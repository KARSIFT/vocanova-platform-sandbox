package wordknowledge

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db: db} }

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func get(ctx context.Context, q queryer, u, m uuid.UUID) (*State, error) {
	st := &State{MeaningID: m}
	err := q.QueryRowContext(ctx, `SELECT COALESCE(k.self_reported_known,false),COALESCE(k.note,''),k.updated_at FROM word_meanings wm CROSS JOIN users u LEFT JOIN user_word_knowledge k ON k.meaning_id=wm.id AND k.user_id=u.id WHERE wm.id=$2 AND u.id=$1 AND u.status='active' AND u.deleted_at IS NULL`, u, m).Scan(&st.SelfReportedKnown, &st.Note, &st.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return st, err
}
func (r *PostgreSQLRepository) Get(ctx context.Context, u, m uuid.UUID) (*State, error) {
	return get(ctx, r.db, u, m)
}
func (r *PostgreSQLRepository) Write(ctx context.Context, req WriteRequest, now time.Time) (*State, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Serialize mutations and recheck the user's active state under lock.
	var id uuid.UUID
	err = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL FOR UPDATE`, req.UserID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	fp := fingerprint(req)
	var prior string
	err = tx.QueryRowContext(ctx, `SELECT fingerprint FROM word_knowledge_actions WHERE user_id=$1 AND idempotency_key=$2`, req.UserID, req.IdempotencyKey).Scan(&prior)
	if err == nil {
		if prior != fp {
			return nil, ErrConflict
		}
		// Return current state on replay, never resurrect older note text or assessment.
		st, e := get(ctx, tx, req.UserID, req.MeaningID)
		if e != nil {
			return nil, e
		}
		return st, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if _, err = get(ctx, tx, req.UserID, req.MeaningID); err != nil {
		return nil, err
	}
	if req.Delete {
		_, err = tx.ExecContext(ctx, `DELETE FROM user_word_knowledge WHERE user_id=$1 AND meaning_id=$2`, req.UserID, req.MeaningID)
	} else if req.AssessmentOnly {
		// Preserve the note atomically, including edits made after the check loaded.
		_, err = tx.ExecContext(ctx, `INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at) VALUES($1,$2,$3,'',$4) ON CONFLICT(user_id,meaning_id) DO UPDATE SET self_reported_known=EXCLUDED.self_reported_known,updated_at=EXCLUDED.updated_at`, req.UserID, req.MeaningID, req.SelfReportedKnown, now)
	} else {
		_, err = tx.ExecContext(ctx, `INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at) VALUES($1,$2,$3,$4,$5) ON CONFLICT(user_id,meaning_id) DO UPDATE SET self_reported_known=EXCLUDED.self_reported_known,note=EXCLUDED.note,updated_at=EXCLUDED.updated_at`, req.UserID, req.MeaningID, req.SelfReportedKnown, req.Note, now)
	}
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO word_knowledge_actions(user_id,idempotency_key,fingerprint,created_at) VALUES($1,$2,$3,$4)`, req.UserID, req.IdempotencyKey, fp, now)
	if err != nil {
		return nil, err
	}
	st, err := get(ctx, tx, req.UserID, req.MeaningID)
	if err != nil {
		return nil, err
	}
	return st, tx.Commit()
}
func (r *PostgreSQLRepository) KnownStates(ctx context.Context, u uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := map[uuid.UUID]bool{}
	if len(ids) == 0 {
		return out, nil
	}
	values := make([]string, len(ids))
	for i, id := range ids {
		values[i] = id.String()
	}
	rows, err := r.db.QueryContext(ctx, `SELECT k.meaning_id FROM user_word_knowledge k JOIN users u ON u.id=k.user_id WHERE k.user_id=$1 AND k.self_reported_known AND k.meaning_id=ANY($2::uuid[]) AND u.status='active' AND u.deleted_at IS NULL`, u, pq.Array(values))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
func (r *PostgreSQLRepository) CountKnown(ctx context.Context, u uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM user_word_knowledge k JOIN users u ON u.id=k.user_id WHERE k.user_id=$1 AND k.self_reported_known AND u.status='active' AND u.deleted_at IS NULL`, u).Scan(&count)
	return count, err
}
