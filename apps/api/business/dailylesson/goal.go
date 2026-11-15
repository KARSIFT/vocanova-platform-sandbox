package dailylesson

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"time"
)

func (s *Service) goal(ctx context.Context, tx *sql.Tx, u uuid.UUID, now time.Time) (*DailyReviewGoal, error) {
	snap, err := s.reviews.ReconcileDailyGoalTx(ctx, tx, u, now)
	if err != nil {
		return nil, err
	}
	out := &DailyReviewGoal{LocalDate: snap.LocalDate.Format("2006-01-02"), TimeZone: snap.Timezone, Target: snap.ReviewTarget}
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM daily_review_effort WHERE user_id=$1 AND local_date=$2`, u, snap.LocalDate).Scan(&out.Reviewed); err != nil {
		return nil, err
	}
	out.Remaining = max(0, out.Target-out.Reviewed)
	out.Complete = out.Remaining == 0
	var next sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT min(uw.next_review_at) FROM user_words uw JOIN word_meanings wm ON wm.id=uw.meaning_id JOIN canonical_words cw ON cw.id=wm.word_id WHERE uw.user_id=$1 AND uw.deleted_at IS NULL AND uw.status IN('new','learning','reviewing') AND wm.status='active' AND cw.status='active' AND uw.next_review_at>$2 AND NOT EXISTS(SELECT 1 FROM user_word_knowledge k WHERE k.user_id=uw.user_id AND k.meaning_id=uw.meaning_id AND k.self_reported_known)`, u, now).Scan(&next)
	if err != nil {
		return nil, err
	}
	if next.Valid {
		out.NextReviewAt = &next.Time
	}
	return out, nil
}
