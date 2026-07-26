package learning

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgreSQLRepository implements Repository against the P1 user_words schema.
// Optional gamification dependency enables P4 reward wiring inside
// transactions (DOC-06 §3).
type PostgreSQLRepository struct {
	db           *sql.DB
	gamification *gamification.Service
}

// NewPostgreSQLRepository creates a repository backed by db with optional
// gamification integration (nil if not yet wired in).
func NewPostgreSQLRepository(db *sql.DB, opts ...interface{}) *PostgreSQLRepository {
	repo := &PostgreSQLRepository{db: db}
	for _, opt := range opts {
		switch v := opt.(type) {
		case *gamification.Service:
			repo.gamification = v
		}
	}
	return repo
}

func (r *PostgreSQLRepository) SaveUserWord(ctx context.Context, req SaveUserWordRequest, now time.Time) (*SavedMeaning, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var meaningID uuid.UUID
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM word_meanings WHERE id = $1 AND status = 'active'`,
		req.MeaningID,
	).Scan(&meaningID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMeaningNotFound
		}
		return nil, fmt.Errorf("lookup meaning: %w", err)
	}

	var existingID uuid.UUID
	var existingDeletedAt sql.NullTime
	err = tx.QueryRowContext(ctx,
		`SELECT id, deleted_at FROM user_words WHERE user_id = $1 AND meaning_id = $2 ORDER BY deleted_at DESC NULLS FIRST, created_at DESC LIMIT 1 FOR UPDATE`,
		req.UserID, req.MeaningID,
	).Scan(&existingID, &existingDeletedAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("lookup user word: %w", err)
	}

	if err == nil && !existingDeletedAt.Valid {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit existing: %w", err)
		}
		return r.savedMeaningByUserMeaning(ctx, nil, req.UserID, req.MeaningID)
	}

	if err == nil && existingDeletedAt.Valid {
		if _, err := tx.ExecContext(ctx,
			`UPDATE user_words
			 SET deleted_at = NULL, status = 'new', source = $1, review_step = 0,
			     added_at = $2, updated_at = $2
			 WHERE id = $3`,
			req.Source, now, existingID,
		); err != nil {
			return nil, fmt.Errorf("restore user word: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit restore: %w", err)
		}
		return r.savedMeaningByUserMeaning(ctx, nil, req.UserID, req.MeaningID)
	}

	id := uuid.New()
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO user_words (
			id, user_id, meaning_id, status, source, review_step,
			next_review_at, last_reviewed_at, last_result, last_rating,
			consecutive_correct_count, consecutive_incorrect_count, total_review_count,
			correct_review_count, added_at, mastered_at, ignored_at, deleted_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, 'new', $4, 0,
			NULL, NULL, NULL, NULL,
			0, 0, 0,
			0, $5, NULL, NULL, NULL,
			$5, $5
		) RETURNING id`,
		id, req.UserID, req.MeaningID, req.Source, now,
	).Scan(&id); err != nil {
		return nil, fmt.Errorf("insert user word: %w", err)
	}

	// P4 reward wiring: record the +2 point award for word addition inside the
	// existing transaction. This is idempotent via the confidence_point_ledger's
	// (user_id, idempotency_key) unique index. D03 keeps the optional new-word
	// mission goal disabled, so we don't call IncrementWordsAdded.
	if r.gamification != nil {
		// Read current balance inside the transaction to ensure isolation.
		currentBalance, err := r.getLatestPointBalanceTx(ctx, tx, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("get current balance: %w", err)
		}
		if _, _, err := r.gamification.GrantPoint(
			ctx, tx, req.UserID,
			gamification.RewardKindAddWord,
			&id,
			gamification.UserWordAddedKey(id.String()),
			currentBalance,
			now,
			nil,
		); err != nil {
			return nil, fmt.Errorf("grant add-word point: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit insert: %w", err)
	}
	return r.savedMeaningByID(ctx, nil, id)
}

func (r *PostgreSQLRepository) UnsaveUserWord(ctx context.Context, userID, meaningID uuid.UUID, now time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE user_words
		 SET deleted_at = $1, updated_at = $1
		 WHERE user_id = $2 AND meaning_id = $3 AND deleted_at IS NULL`,
		now, userID, meaningID,
	)
	if err != nil {
		return fmt.Errorf("unsave user word: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrUserWordNotFound
	}
	return nil
}

func (r *PostgreSQLRepository) GetSavedMeaning(ctx context.Context, userID, meaningID uuid.UUID) (*SavedMeaning, error) {
	return r.savedMeaningByUserMeaning(ctx, nil, userID, meaningID)
}

func (r *PostgreSQLRepository) ListSavedWords(ctx context.Context, req ListSavedWordsRequest) (*ListSavedWordsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var cursorTime sql.NullTime
	var cursorID uuid.UUID
	if req.AfterCursor != "" {
		c, err := decodeSavedCursor(req.AfterCursor)
		if err != nil {
			return nil, ErrInvalidCursor
		}
		cursorTime = sql.NullTime{Time: c.AddedAt, Valid: true}
		cursorID = c.ID
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT uw.id, uw.meaning_id, cw.id, cw.text, cw.normalized_text,
			wm.part_of_speech, wm.short_definition, uw.status, uw.source, uw.added_at
		 FROM user_words uw
		 JOIN word_meanings wm ON wm.id = uw.meaning_id
		 JOIN canonical_words cw ON cw.id = wm.word_id
		 WHERE uw.user_id = $1 AND uw.deleted_at IS NULL
		   AND ($2::timestamptz IS NULL OR
				uw.added_at < $2 OR
				(uw.added_at = $2 AND uw.id < $3))
		 ORDER BY uw.added_at DESC, uw.id DESC
		 LIMIT $4`,
		req.UserID, cursorTime, cursorID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list saved words: %w", err)
	}
	defer rows.Close()

	var items []SavedMeaning
	var last SavedMeaning
	for rows.Next() {
		var m SavedMeaning
		var normalizedText string
		if err := rows.Scan(&m.UserWordID, &m.MeaningID, &m.WordID, &m.WordText, &normalizedText,
			&m.PartOfSpeech, &m.ShortDefinition, &m.Status, &m.Source, &m.AddedAt); err != nil {
			return nil, fmt.Errorf("scan saved word: %w", err)
		}
		m.WordSlug = wordSlug(normalizedText)
		m.Saved = true
		items = append(items, m)
		last = m
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list saved words rows: %w", err)
	}

	resp := &ListSavedWordsResponse{Items: items}
	if len(items) == limit {
		resp.NextCursor = encodeSavedCursor(savedCursor{AddedAt: last.AddedAt, ID: last.UserWordID})
	}
	return resp, nil
}

func (r *PostgreSQLRepository) IsSaved(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := make(map[uuid.UUID]bool, len(meaningIDs))
	for _, id := range meaningIDs {
		out[id] = false
	}
	if len(meaningIDs) == 0 {
		return out, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT meaning_id FROM user_words WHERE user_id = $1 AND meaning_id = ANY($2::uuid[]) AND deleted_at IS NULL`,
		userID, pq.Array(meaningIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("is saved: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan is saved: %w", err)
		}
		out[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("is saved rows: %w", err)
	}
	return out, nil
}

// SavedUserWordIDs implements SavedStateReader for the learning repository.
func (r *PostgreSQLRepository) SavedUserWordIDs(ctx context.Context, userID uuid.UUID, meaningIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	out := make(map[uuid.UUID]uuid.UUID, len(meaningIDs))
	if len(meaningIDs) == 0 {
		return out, nil
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT meaning_id, id FROM user_words WHERE user_id = $1 AND meaning_id = ANY($2::uuid[]) AND deleted_at IS NULL`,
		userID, pq.Array(meaningIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("saved user word ids: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var meaningID, userWordID uuid.UUID
		if err := rows.Scan(&meaningID, &userWordID); err != nil {
			return nil, fmt.Errorf("scan saved user word id: %w", err)
		}
		out[meaningID] = userWordID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("saved user word id rows: %w", err)
	}
	return out, nil
}

func (r *PostgreSQLRepository) savedMeaningByID(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*SavedMeaning, error) {
	var row *sql.Row
	q := `SELECT uw.id, uw.meaning_id, cw.id, cw.text, cw.normalized_text,
		     wm.part_of_speech, wm.short_definition, uw.status, uw.source, uw.added_at
		  FROM user_words uw
		  JOIN word_meanings wm ON wm.id = uw.meaning_id
		  JOIN canonical_words cw ON cw.id = wm.word_id
		  WHERE uw.id = $1`
	if tx != nil {
		row = tx.QueryRowContext(ctx, q, id)
	} else {
		row = r.db.QueryRowContext(ctx, q, id)
	}
	return r.scanSavedMeaning(row)
}

func (r *PostgreSQLRepository) savedMeaningByUserMeaning(ctx context.Context, tx *sql.Tx, userID, meaningID uuid.UUID) (*SavedMeaning, error) {
	var row *sql.Row
	q := `SELECT uw.id, uw.meaning_id, cw.id, cw.text, cw.normalized_text,
		     wm.part_of_speech, wm.short_definition, uw.status, uw.source, uw.added_at
		  FROM user_words uw
		  JOIN word_meanings wm ON wm.id = uw.meaning_id
		  JOIN canonical_words cw ON cw.id = wm.word_id
		  WHERE uw.user_id = $1 AND uw.meaning_id = $2 AND uw.deleted_at IS NULL`
	if tx != nil {
		row = tx.QueryRowContext(ctx, q, userID, meaningID)
	} else {
		row = r.db.QueryRowContext(ctx, q, userID, meaningID)
	}
	return r.scanSavedMeaning(row)
}

func (r *PostgreSQLRepository) scanSavedMeaning(row *sql.Row) (*SavedMeaning, error) {
	var m SavedMeaning
	var normalizedText string
	err := row.Scan(&m.UserWordID, &m.MeaningID, &m.WordID, &m.WordText, &normalizedText,
		&m.PartOfSpeech, &m.ShortDefinition, &m.Status, &m.Source, &m.AddedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserWordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan saved meaning: %w", err)
	}
	m.WordSlug = wordSlug(normalizedText)
	m.Saved = true
	return &m, nil
}

// getLatestPointBalanceTx reads the current confidence points balance for a user
// inside the given transaction to ensure isolation from concurrent updates.
// Returns 0 if no balance entries exist.
func (r *PostgreSQLRepository) getLatestPointBalanceTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (int, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT COALESCE(balance_after, 0) FROM confidence_point_ledger
		 WHERE user_id = $1
		 ORDER BY occurred_at DESC, id DESC
		 LIMIT 1`,
		userID,
	)
	var balance int
	if err := row.Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("fetch latest point balance: %w", err)
	}
	return balance, nil
}
