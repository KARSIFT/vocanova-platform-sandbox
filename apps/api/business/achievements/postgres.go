package achievements

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db} }

// Completion receipts are transactionally committed with the final session state.
// They are never rewritten on replay. Review timestamps come from server-created
// history, not learner-supplied answered_at. No saved-word joins can revoke a badge.
func (r *PostgreSQLRepository) Metrics(ctx context.Context, user uuid.UUID) (map[string]Metric, error) {
	rows, err := r.db.QueryContext(ctx, `WITH owner AS(
 SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL
),lesson_events AS(
 SELECT ls.lesson_key event_id,min(la.created_at) occurred_at
 FROM owner u JOIN lesson_sessions ls ON ls.user_id=u.id JOIN lesson_actions la ON la.session_id=ls.id AND la.user_id=u.id
 WHERE ls.status='completed' AND ls.completed_at IS NOT NULL AND la.operation='action' AND la.action->>'action'='continue' AND la.result->>'status'='completed'
 GROUP BY ls.lesson_key
),practice_events AS(
 SELECT ps.id::text event_id,min(pa.created_at) occurred_at,
 (ps.mode IN('typed_recall','mistakes') AND ps.questions_answered=ps.total_steps AND ps.first_answers_correct=ps.total_steps) independent
 FROM owner u JOIN practice_sessions ps ON ps.user_id=u.id JOIN practice_actions pa ON pa.session_id=ps.id AND pa.user_id=u.id
 WHERE ps.status='completed' AND ps.completed_at IS NOT NULL AND pa.operation='action' AND pa.action->>'action'='continue' AND pa.result->>'status'='completed'
 GROUP BY ps.id
),writing_events AS(
 SELECT s.id::text event_id,min(a.completed_at) occurred_at
 FROM owner u JOIN learner_sentences s ON s.user_id=u.id JOIN ai_feedback_attempts a ON a.learner_sentence_id=s.id
 WHERE s.deleted_at IS NULL AND a.status='succeeded' AND a.completed_at IS NOT NULL AND a.feedback_json->>'status' IN('correct','needs_improvement','incorrect')
 GROUP BY s.id
),events AS(
 SELECT 'lessons' kind,event_id,occurred_at FROM lesson_events
 UNION ALL SELECT 'practice',event_id,occurred_at FROM practice_events
 UNION ALL SELECT 'independent',event_id,occurred_at FROM practice_events WHERE independent
 UNION ALL SELECT 'reviews',r.id::text,r.created_at FROM owner u JOIN review_attempts r ON r.user_id=u.id WHERE r.attempt_type='review' AND r.prompt_type IN('multiple_choice','self_check') AND r.result IN('correct','incorrect')
 UNION ALL SELECT 'writing',event_id,occurred_at FROM writing_events
),ranked AS(
 SELECT *,row_number() OVER(PARTITION BY kind ORDER BY occurred_at,event_id) ordinal FROM events
),metrics AS(
 SELECT kind,count(*) count,
 min(occurred_at) FILTER(WHERE ordinal=1) first_at,
 min(occurred_at) FILTER(WHERE ordinal=3) third_at,
 min(occurred_at) FILTER(WHERE ordinal=5) fifth_at,
 min(occurred_at) FILTER(WHERE ordinal=7) seventh_at,
 min(occurred_at) FILTER(WHERE ordinal=10) tenth_at,
 min(occurred_at) FILTER(WHERE ordinal=50) fiftieth_at
 FROM ranked GROUP BY kind
)
 SELECT owner.id,metrics.kind,metrics.count,first_at,third_at,fifth_at,seventh_at,tenth_at,fiftieth_at FROM owner LEFT JOIN metrics ON true`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]Metric{}
	found := false
	for rows.Next() {
		found = true
		var id uuid.UUID
		var kind sql.NullString
		var count sql.NullInt64
		var times [6]sql.NullTime
		if err = rows.Scan(&id, &kind, &count, &times[0], &times[1], &times[2], &times[3], &times[4], &times[5]); err != nil {
			return nil, err
		}
		if !kind.Valid {
			continue
		}
		m := Metric{Count: int(count.Int64), Thresholds: map[int]time.Time{}}
		for i, n := range []int{1, 3, 5, 7, 10, 50} {
			if times[i].Valid {
				m.Thresholds[n] = times[i].Time
			}
		}
		out[kind.String] = m
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNotFound
	}
	return out, nil
}
