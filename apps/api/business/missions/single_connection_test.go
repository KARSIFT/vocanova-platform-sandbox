package missions

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestExistingSnapshotReadUsesTransactionConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	userID, snapshotID := uuid.New(), uuid.New()
	today := fixedDay()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, user_id, local_date, timezone, review_target, reviews_completed").
		WithArgs(userID, today).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "local_date", "timezone", "review_target", "reviews_completed",
			"new_word_target", "new_words_completed", "sentence_practice_target",
			"sentence_practices_completed", "policy_version", "status", "completed_at",
			"grace_applied", "grace_day_id",
		}).AddRow(snapshotID, userID, today, "UTC", 20, 3, nil, nil, nil, nil,
			gamification.MissionPolicyVersion, "open", nil, false, nil))
	mock.ExpectCommit()
	svc := NewService(NewRepository(db), nil)
	snapshot, err := svc.ensureTodaySnapshotAndReconcile(ctx, userID,
		gamification.ResolvedSettings{Timezone: "UTC", DailyReviewTarget: 20}, today, fixedNow())
	require.NoError(t, err)
	require.Equal(t, snapshotID.String(), snapshot.ID)
	require.Equal(t, 3, snapshot.ReviewsCompleted)
	require.NoError(t, mock.ExpectationsWereMet())
}
