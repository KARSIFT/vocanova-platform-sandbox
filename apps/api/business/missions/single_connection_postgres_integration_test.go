//go:build integration

package missions

import (
	"context"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// The explicitly supplied test database gets an isolated, fully migrated schema.
// This proves actual PostgreSQL transaction visibility and bounded-pool behavior,
// rather than only matching the SQL strings through a mock driver.
func TestPostgreSQLSingleConnectionMissionAndSentenceRecovery(t *testing.T) {
	db := newMigratedPostgresFromEnv(t)
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	userID := insertTestUser(t, db)
	gam := gamification.NewService(gamification.NewRepository(db))
	service := NewService(NewRepository(db), gam)
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	first, err := service.GetDailyMissionView(ctx, userID, "UTC", now)
	require.NoError(t, err)
	require.Equal(t, 0, first.ReviewsCompleted)
	second, err := service.GetDailyMissionView(ctx, userID, "UTC", now)
	require.NoError(t, err)
	require.Equal(t, first.LocalDate, second.LocalDate)
	updater := NewMissionUpdater(service, gam)
	sentenceID, attemptID := uuid.New(), uuid.New()
	settings := gamification.ResolvedSettings{Timezone: "UTC", DailyReviewTarget: 20}
	completed, err := updater.UpdateForSentence(ctx, userID, sentenceID, attemptID, settings, now, false)
	require.NoError(t, err)
	require.False(t, completed)
	_, err = updater.UpdateForSentence(ctx, userID, sentenceID, attemptID, settings, now, false)
	require.NoError(t, err)
	points, err := gam.CurrentBalance(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, gamification.RewardSentenceSubmitted+gamification.RewardAIFeedbackGot, points)
	_, err = service.GetDailyMissionView(ctx, userID, "UTC", now.AddDate(0, 0, 1))
	require.NoError(t, err)
}
