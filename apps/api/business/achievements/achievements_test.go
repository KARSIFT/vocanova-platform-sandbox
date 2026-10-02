package achievements

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type metricsRepo map[string]Metric

func (r metricsRepo) Metrics(context.Context, uuid.UUID) (map[string]Metric, error) { return r, nil }
func TestAchievementsFixedCatalogHonestThresholds(t *testing.T) {
	at := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	svc := NewService(metricsRepo{"lessons": {Count: 4, Thresholds: map[int]time.Time{1: at, 3: at.Add(time.Hour)}}, "practice": {Count: 500, Thresholds: map[int]time.Time{5: at}}, "reviews": {Count: 50, Thresholds: map[int]time.Time{10: at}}, "writing": {Count: 1, Thresholds: map[int]time.Time{1: at}}})
	out, err := svc.List(t.Context(), uuid.New())
	require.NoError(t, err)
	require.Equal(t, "1", out.CatalogVersion)
	require.Len(t, out.Items, 8)
	seen := map[string]bool{}
	for _, a := range out.Items {
		require.False(t, seen[a.ID])
		seen[a.ID] = true
		require.LessOrEqual(t, a.Current, a.Target)
		if a.Earned {
			require.NotNil(t, a.EarnedAt)
		} else {
			require.Nil(t, a.EarnedAt)
		}
	}
	require.Equal(t, "guided-first", out.Items[0].ID)
	require.True(t, out.Items[0].Earned)
	require.Equal(t, at.Add(time.Hour), *out.Items[1].EarnedAt)
	require.False(t, out.Items[2].Earned)
	require.Equal(t, 4, out.Items[2].Current)
	require.False(t, out.Items[3].Earned, "participation is not unaided recall")
	require.Equal(t, 5, out.Items[4].Current)
	require.True(t, out.Items[4].Earned)
	require.False(t, out.Items[6].Earned, "a count without its confirmed threshold event cannot fabricate an earned timestamp")
}
func TestAchievementsEmptyAndAnonymous(t *testing.T) {
	svc := NewService(metricsRepo{})
	_, err := svc.List(t.Context(), uuid.Nil)
	require.ErrorIs(t, err, ErrNotFound)
	out, err := svc.List(t.Context(), uuid.New())
	require.NoError(t, err)
	require.Len(t, out.Items, 8)
	for _, a := range out.Items {
		require.Zero(t, a.Current)
		require.False(t, a.Earned)
		require.Nil(t, a.EarnedAt)
	}
}
