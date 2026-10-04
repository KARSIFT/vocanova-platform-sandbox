package wordlists

import (
	"context"
	"database/sql"
	"errors"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/practice"
	"github.com/google/uuid"
	"strings"
	"time"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db} }

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func active(ctx context.Context, q queryer, u uuid.UUID, lock bool) error {
	suffix := ""
	if lock {
		suffix = " FOR UPDATE"
	}
	var id uuid.UUID
	err := q.QueryRowContext(ctx, `SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL`+suffix, u).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
func get(ctx context.Context, q queryer, u, id uuid.UUID) (*WordListDetail, error) {
	st := &WordListDetail{Members: []WordListMember{}}
	err := q.QueryRowContext(ctx, `SELECT l.id,l.name,l.revision,l.created_at,l.updated_at FROM user_word_lists l JOIN users u ON u.id=l.user_id WHERE l.id=$2 AND l.user_id=$1 AND l.deleted_at IS NULL AND u.status='active' AND u.deleted_at IS NULL`, u, id).Scan(&st.ID, &st.Name, &st.Revision, &st.CreatedAt, &st.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	rows, err := q.QueryContext(ctx, `SELECT m.meaning_id,cw.id,cw.text,cw.normalized_text,wm.short_definition,wm.part_of_speech,m.created_at,(wm.status='active' AND cw.status='active') FROM user_word_list_members m JOIN word_meanings wm ON wm.id=m.meaning_id JOIN canonical_words cw ON cw.id=wm.word_id WHERE m.list_id=$1 ORDER BY lower(cw.text),wm.meaning_order,wm.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m WordListMember
		var enabled bool
		if err = rows.Scan(&m.MeaningID, &m.WordID, &m.WordText, &m.WordSlug, &m.ShortDefinition, &m.PartOfSpeech, &m.AddedAt, &enabled); err != nil {
			return nil, err
		}
		m.WordSlug = strings.ReplaceAll(m.WordSlug, " ", "-")
		m.PracticeAvailable = enabled && practice.SupportsMeaning(m.MeaningID, m.WordText, m.ShortDefinition)
		st.Members = append(st.Members, m)
		if m.PracticeAvailable {
			st.UsableMemberCount++
		}
	}
	st.MemberCount = len(st.Members)
	return st, rows.Err()
}
func (r *PostgreSQLRepository) Get(ctx context.Context, u, id uuid.UUID) (*WordListDetail, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	st, err := get(ctx, tx, u, id)
	if err != nil {
		return nil, err
	}
	return st, tx.Commit()
}
func (r *PostgreSQLRepository) List(ctx context.Context, u uuid.UUID) (*WordListsResponse, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = active(ctx, tx, u, false); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM user_word_lists WHERE user_id=$1 AND deleted_at IS NULL ORDER BY lower(name),id`, u)
	if err != nil {
		return nil, err
	}
	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	out := &WordListsResponse{Items: []WordListSummary{}}
	for _, id := range ids {
		st, e := get(ctx, tx, u, id)
		if e != nil {
			return nil, e
		}
		out.Items = append(out.Items, st.WordListSummary)
	}
	return out, tx.Commit()
}
func (r *PostgreSQLRepository) Write(ctx context.Context, req WriteRequest, now time.Time) (*WordListDetail, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = active(ctx, tx, req.UserID, true); err != nil {
		return nil, err
	}
	fp := fingerprint(req)
	var prior string
	err = tx.QueryRowContext(ctx, `SELECT fingerprint FROM word_list_actions WHERE user_id=$1 AND idempotency_key=$2`, req.UserID, req.IdempotencyKey).Scan(&prior)
	if err == nil {
		if prior != fp {
			return nil, ErrConflict
		}
		if req.Operation == "delete" {
			return nil, tx.Commit()
		}
		st, e := get(ctx, tx, req.UserID, req.ListID)
		if e != nil {
			return nil, e
		}
		return st, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	var revision int
	var deleted *time.Time
	err = tx.QueryRowContext(ctx, `SELECT revision,deleted_at FROM user_word_lists WHERE id=$1 AND user_id=$2 FOR UPDATE`, req.ListID, req.UserID).Scan(&revision, &deleted)
	fresh := errors.Is(err, sql.ErrNoRows)
	if fresh {
		if req.Operation != "put" || req.ExpectedRevision != 0 {
			return nil, ErrNotFound
		}
		var count int
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM user_word_lists WHERE user_id=$1 AND deleted_at IS NULL`, req.UserID).Scan(&count); err != nil {
			return nil, err
		}
		if count >= MaxLists {
			return nil, ErrLimit
		}
		// A deleted or foreign ID cannot be overwritten by caller-chosen UUID creation.
		result, e := tx.ExecContext(ctx, `INSERT INTO user_word_lists(id,user_id,name,revision,created_at,updated_at)VALUES($1,$2,$3,1,$4,$4) ON CONFLICT(id) DO NOTHING`, req.ListID, req.UserID, req.Name, now)
		if e != nil {
			return nil, e
		}
		n, e := result.RowsAffected()
		if e != nil {
			return nil, e
		}
		if n != 1 {
			return nil, ErrConflict
		}
	} else {
		if err != nil {
			return nil, err
		}
		if deleted != nil {
			return nil, ErrNotFound
		}
		if revision != req.ExpectedRevision {
			return nil, ErrConflict
		}
		switch req.Operation {
		case "put":
			_, err = tx.ExecContext(ctx, `UPDATE user_word_lists SET name=$2 WHERE id=$1`, req.ListID, req.Name)
		case "delete":
			_, err = tx.ExecContext(ctx, `DELETE FROM user_word_list_members WHERE list_id=$1`, req.ListID)
			if err == nil {
				_, err = tx.ExecContext(ctx, `UPDATE user_word_lists SET deleted_at=$2 WHERE id=$1`, req.ListID, now)
			}
		case "add":
			var found uuid.UUID
			err = tx.QueryRowContext(ctx, `SELECT wm.id FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id WHERE wm.id=$1 AND wm.status='active' AND cw.status='active'`, req.MeaningID).Scan(&found)
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			if err != nil {
				return nil, err
			}
			var count int
			var exists bool
			err = tx.QueryRowContext(ctx, `SELECT count(*),COALESCE(bool_or(meaning_id=$2),false) FROM user_word_list_members WHERE list_id=$1`, req.ListID, req.MeaningID).Scan(&count, &exists)
			if err != nil {
				return nil, err
			}
			if count >= MaxMembers && !exists {
				return nil, ErrLimit
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO user_word_list_members(list_id,meaning_id,created_at)VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, req.ListID, req.MeaningID, now)
		case "remove":
			_, err = tx.ExecContext(ctx, `DELETE FROM user_word_list_members WHERE list_id=$1 AND meaning_id=$2`, req.ListID, req.MeaningID)
		}
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, `UPDATE user_word_lists SET revision=revision+1,updated_at=$2 WHERE id=$1`, req.ListID, now)
		if err != nil {
			return nil, err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO word_list_actions(user_id,idempotency_key,fingerprint,list_id,operation,created_at)VALUES($1,$2,$3,$4,$5,$6)`, req.UserID, req.IdempotencyKey, fp, req.ListID, req.Operation, now)
	if err != nil {
		return nil, err
	}
	if req.Operation == "delete" {
		return nil, tx.Commit()
	}
	st, err := get(ctx, tx, req.UserID, req.ListID)
	if err != nil {
		return nil, err
	}
	return st, tx.Commit()
}
