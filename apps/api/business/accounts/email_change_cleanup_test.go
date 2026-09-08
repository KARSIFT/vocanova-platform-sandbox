package accounts

import (
	"context"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestServiceCleanupExpiredEmailChangeLinksRemovesOnlyInactiveLinks(t *testing.T) {
	now := testNow()
	repo := NewMemoryRepository()
	userID := uuid.New()
	create := func(label string, expiresAt time.Time) *EmailChangeLink {
		t.Helper()
		link, err := repo.CreateEmailChangeLink(context.Background(), userID, label+"@example.test", []byte(label+"-012345678901234567890123456"), "test", now.Add(-time.Minute), expiresAt)
		require.NoError(t, err)
		return link
	}

	expired := create("expired", now)
	consumed := create("consumed", now.Add(time.Minute))
	revoked := create("revoked", now.Add(time.Minute))
	require.NoError(t, repo.ConsumeEmailChangeLink(context.Background(), consumed.ID, now))
	_, err := repo.RevokeAllEmailChangeLinksForUser(context.Background(), userID, now)
	require.NoError(t, err)

	// Recreate the active link after the user-wide revoke so it is the sole
	// valid link that must survive the cleanup pass.
	active := create("active-after-revoke", now.Add(time.Minute))

	svc := NewService(repo, nil, nil, nil, &clock.Fixed{T: now}, nil, Config{})
	deleted, err := svc.CleanupExpiredEmailChangeLinks(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, int64(3), deleted)

	links := repo.LinksForUser(userID)
	require.Len(t, links, 1)
	require.Equal(t, active.ID, links[0].ID)
	require.NotEqual(t, expired.ID, links[0].ID)
	require.NotEqual(t, consumed.ID, links[0].ID)
	require.NotEqual(t, revoked.ID, links[0].ID)
}
