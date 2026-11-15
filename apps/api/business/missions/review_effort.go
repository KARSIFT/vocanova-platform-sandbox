package missions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/google/uuid"
	"time"
)

// CompleteReviewGoalTx preserves one-time mission rewards and streak behavior.
// Caller holds the user lock and has already persisted effort and activity.
func (s *Service) CompleteReviewGoalTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, resolved gamification.ResolvedSettings, snap *DailyMissionSnapshot, now time.Time) error {
	// If the review target is met for the first time today, transition
	// the snapshot to completed, record the +10 daily-mission-completion
	// award, and let the streak reconciliation know the day completed
	// (so it advances the streak and may earn a grace day).
	missionCompletedNow := false
	if snap.ReviewsCompleted >= snap.ReviewTarget && snap.Status == StatusOpen {
		completed, err := s.MarkSnapshotCompleted(ctx, tx, userID, snap.LocalDate, now)
		if err != nil {
			return fmt.Errorf("mark snapshot completed: %w", err)
		}
		if completed {
			missionCompletedNow = true
			balance, err := s.gamification.CurrentBalanceTx(ctx, tx, userID)
			if err != nil {
				return fmt.Errorf("get latest point balance: %w", err)
			}
			localDateKey := snap.LocalDate.Format("2006-01-02")
			newBalance, _, _, err := s.gamification.GrantPoint(
				ctx, tx, userID,
				gamification.RewardKindDailyMissionDone, nil,
				gamification.DailyMissionCompletedKey(userID.String(), localDateKey),
				balance, now, nil,
			)
			if err != nil {
				return fmt.Errorf("grant daily mission point: %w", err)
			}
			if err := s.RecordConfidencePointChange(
				ctx, tx, userID, snap.LocalDate, resolved.Timezone, newBalance-balance,
			); err != nil {
				return fmt.Errorf("increment points earned: %w", err)
			}
		}
	}

	// Streak reconciliation: lazy, no queue/cron (DOC-06 §15), computed
	// from the recent snapshot history. Reads happen on the caller's tx
	// (consistent read) and any grace-day ledger writes happen here too.
	snaps, err := s.fetchStreakSnapshotsTx(ctx, tx, userID)
	if err != nil {
		return err
	}
	reconciliation, err := s.gamification.ReconcileAndAdvance(
		ctx, tx, userID, now, resolved.Timezone,
		snaps, missionCompletedNow,
	)
	if err != nil {
		return fmt.Errorf("reconcile streak: %w", err)
	}
	if reconciliation.GraceDayUsed != nil && reconciliation.GraceDayUsedID != nil && reconciliation.YesterdayProtectedLocalDate != nil {
		protected, err := s.MarkSnapshotProtected(
			ctx, tx, userID, *reconciliation.YesterdayProtectedLocalDate, *reconciliation.GraceDayUsedID, resolved.Timezone,
			snap.ReviewTarget,
		)
		if err != nil {
			return fmt.Errorf("mark grace-protected snapshot: %w", err)
		}
		if !protected {
			return errors.New("mark grace-protected snapshot: missed snapshot was not updated")
		}
	}
	return nil
}

// ReconcileReviewEffortTx credits only today's immutable, server-confirmed answer
// receipts. Client-supplied metadata/source are insufficient. No prior date is
// updated, no activity or rating XP is replayed, and completion is idempotent.
func (s *Service) ReconcileReviewEffortTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, resolved gamification.ResolvedSettings, now time.Time) (*DailyMissionSnapshot, error) {
	snap, err := s.EnsureTodaySnapshot(ctx, tx, userID, resolved, now)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO daily_review_effort(user_id,word_id,meaning_id,local_date,review_attempt_id,daily_session_id,created_at)
 SELECT DISTINCT ON(wm.word_id) a.user_id,wm.word_id,a.meaning_id,$2,a.id,d.session_id,a.created_at
 FROM review_attempts a JOIN word_meanings wm ON wm.id=a.meaning_id
 JOIN daily_lesson_actions d ON d.user_id=a.user_id AND d.operation='action'
 AND a.client_attempt_id='daily:'||encode(sha256(convert_to(replace(replace(replace(replace(replace(to_json(d.action->>'clientActionId')::text,'&','\u0026'),'<','\u003c'),'>','\u003e'),chr(8232),'\u2028'),chr(8233),'\u2029'),'UTF8')),'hex')
 AND a.created_at=d.created_at
 JOIN daily_lesson_sessions lesson ON lesson.id=d.session_id AND lesson.user_id=a.user_id
 WHERE a.user_id=$1 AND a.result IN('correct','incorrect') AND NOT a.was_hint_used
 AND a.created_at>=($2::date::timestamp AT TIME ZONE $3) AND a.created_at<(($2::date+1)::timestamp AT TIME ZONE $3)
 AND d.action->>'action'='answer' AND COALESCE((d.action->>'wasHintUsed')::boolean,false)=false
 AND d.result->'feedback'->>'meaningId'=a.meaning_id::text
 AND d.result->'feedback'->>'selectedChoiceId'=d.action->>'choiceId'
 AND d.result->'feedback'->>'skipped'='false'
 AND EXISTS(SELECT 1 FROM jsonb_array_elements(lesson.state->'Words') w WHERE w->>'MeaningID'=a.meaning_id::text AND w->>'UserWordID'=a.user_word_id::text)
 ORDER BY wm.word_id,a.created_at,a.id
 ON CONFLICT DO NOTHING`, userID, snap.LocalDate, snap.Timezone)
	if err != nil {
		return nil, fmt.Errorf("reconcile review effort: %w", err)
	}
	err = tx.QueryRowContext(ctx, `UPDATE daily_mission_snapshots SET reviews_completed=LEAST(review_target,(SELECT count(*) FROM daily_review_effort WHERE user_id=$1 AND local_date=$2)),updated_at=$3 WHERE user_id=$1 AND local_date=$2 RETURNING reviews_completed`, userID, snap.LocalDate, now).Scan(&snap.ReviewsCompleted)
	if err != nil {
		return nil, err
	}
	if err = s.CompleteReviewGoalTx(ctx, tx, userID, resolved, snap, now); err != nil {
		return nil, err
	}
	return s.missions.GetDailyMissionSnapshotInTx(ctx, tx, userID, snap.LocalDate)
}

// fetchStreakSnapshotsTx reads the recent daily_mission_snapshots history
// for the user (a 14-day window is sufficient for streak reconciliation)
// inside the caller's tx so the read is consistent with the writes just
// performed by the P4 wiring block.
func (s *Service) fetchStreakSnapshotsTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID) ([]gamification.StreakSnapshot, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT local_date, status, completed_at, grace_applied, grace_day_id
		 FROM daily_mission_snapshots
		 WHERE user_id = $1
		 ORDER BY local_date DESC
		 LIMIT $2`,
		userID, 14,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch streak snapshots: %w", err)
	}
	defer rows.Close()
	var out []gamification.StreakSnapshot
	for rows.Next() {
		var s gamification.StreakSnapshot
		var completedAt sql.NullTime
		var graceDayID *uuid.UUID
		if err := rows.Scan(&s.LocalDate, &s.Status, &completedAt, &s.GraceApplied, &graceDayID); err != nil {
			return nil, fmt.Errorf("scan streak snapshot: %w", err)
		}
		if completedAt.Valid {
			t := completedAt.Time
			s.CompletedAt = &t
		}
		if graceDayID != nil {
			id := graceDayID.String()
			s.GraceDayID = &id
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan streak snapshots rows: %w", err)
	}
	return out, nil
}
