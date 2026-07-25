package reviews

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// PostgreSQLRepository implements Repository against the VOC-027 review schema.
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository creates a repository backed by db.
func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

func (r *PostgreSQLRepository) ListDueWords(ctx context.Context, req ListDueWordsRequest) (*ListDueWordsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var cursorNextReviewAt sql.NullTime
	var cursorID uuid.UUID
	if req.AfterCursor != "" {
		c, err := decodeDueCursor(req.AfterCursor)
		if err != nil {
			return nil, ErrInvalidCursor
		}
		cursorID = c.ID
		if !c.NextReviewAt.IsZero() {
			cursorNextReviewAt = sql.NullTime{Time: c.NextReviewAt, Valid: true}
		}
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT
		   uw.id, uw.meaning_id, cw.id, cw.text, cw.normalized_text,
		   wm.part_of_speech, wm.short_definition, uw.status, uw.source, uw.review_step,
		   uw.next_review_at,
		   COUNT(*) OVER() AS total_count
		 FROM user_words uw
		 JOIN word_meanings wm ON wm.id = uw.meaning_id
		 JOIN canonical_words cw ON cw.id = wm.word_id
		 WHERE uw.user_id = $1
		   AND uw.deleted_at IS NULL
		   AND uw.status IN ('new', 'learning', 'reviewing')
		   AND (uw.next_review_at IS NULL OR uw.next_review_at <= NOW())
		   AND ($2::timestamptz IS NULL OR
		        COALESCE(uw.next_review_at, '-infinity'::timestamptz) > $2 OR
		        (COALESCE(uw.next_review_at, '-infinity'::timestamptz) = $2 AND uw.id > $3))
		 ORDER BY COALESCE(uw.next_review_at, '-infinity'::timestamptz) ASC, uw.id ASC
		 LIMIT $4`,
		req.UserID, cursorNextReviewAt, cursorID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list due words: %w", err)
	}
	defer rows.Close()

	var items []DueWord
	var totalCount int
	var last DueWord
	for rows.Next() {
		var d DueWord
		var normalizedText string
		var nextReviewAt sql.NullTime
		if err := rows.Scan(
			&d.UserWordID, &d.MeaningID, &d.WordID, &d.WordText, &normalizedText,
			&d.PartOfSpeech, &d.ShortDefinition, &d.Status, &d.Source, &d.ReviewStep,
			&nextReviewAt, &totalCount,
		); err != nil {
			return nil, fmt.Errorf("scan due word: %w", err)
		}
		d.WordSlug = wordSlug(normalizedText)
		if nextReviewAt.Valid {
			d.NextReviewAt = &nextReviewAt.Time
		}
		items = append(items, d)
		last = d
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list due words rows: %w", err)
	}

	resp := &ListDueWordsResponse{Items: items, TotalCount: totalCount}
	if len(items) == limit {
		cursor := dueCursor{ID: last.UserWordID}
		if last.NextReviewAt != nil {
			cursor.NextReviewAt = *last.NextReviewAt
		}
		resp.NextCursor = encodeDueCursor(cursor)
	}
	return resp, nil
}

func wordSlug(normalized string) string {
	out := ""
	for _, r := range normalized {
		if r == ' ' {
			out += "-"
		} else {
			out += string(r)
		}
	}
	return out
}
