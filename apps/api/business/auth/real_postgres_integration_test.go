package auth

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// blockingCreateUserRepository makes two complete magic-link consumes reach
// the first-user INSERT together. PostgreSQL remains the persistence layer;
// the wrapper only makes the otherwise timing-sensitive interleaving reliable.
type blockingCreateUserRepository struct {
	Repository
	entered chan struct{}
	release chan struct{}
	once    sync.Once
}

func (r *blockingCreateUserRepository) CreateUser(ctx context.Context, email string, verifiedAt *time.Time) (*User, error) {
	r.entered <- struct{}{}
	<-r.release
	return r.Repository.CreateUser(ctx, email, verifiedAt)
}

// TestConsumeMagicLinkConcurrentFirstLoginPostgreSQL proves two independently
// issued links for a new mailbox both complete without leaking a database
// uniqueness failure. It uses only a disposable database supplied by the caller.
func TestConsumeMagicLinkConcurrentFirstLoginPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is not set")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.PingContext(context.Background()))

	emailAddr := "first-login-race-" + time.Now().UTC().Format("20060102150405.000000000") + "@example.test"
	baseRepo := NewPostgreSQLRepository(db)
	repo := &blockingCreateUserRepository{
		Repository: baseRepo,
		entered:    make(chan struct{}, 2),
		release:    make(chan struct{}),
	}
	fake := &email.Fake{}
	now := time.Now().UTC()
	svc := NewService(repo, fake, nil, &clock.Fixed{T: now}, NewFixedWindowRateLimiter(&clock.Fixed{T: now}, time.Hour, 100), testConfig())
	ctx := context.Background()

	require.NoError(t, svc.RequestMagicLink(ctx, "1.2.3.4", emailAddr))
	require.NoError(t, svc.RequestMagicLink(ctx, "1.2.3.5", emailAddr))
	require.Len(t, fake.Sent, 2)
	tokens := []string{
		extractTokenFromURL(t, fake.Sent[0].BodyText),
		extractTokenFromURL(t, fake.Sent[1].BodyText),
	}

	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, token := range tokens {
		wg.Add(1)
		go func(clientIP, token string) {
			defer wg.Done()
			_, _, _, err := svc.ConsumeMagicLink(ctx, clientIP, token, emailAddr)
			errs <- err
		}("1.2.3.6", token)
	}
	for range tokens {
		<-repo.entered
	}
	repo.once.Do(func() { close(repo.release) })
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	var users, consumedLinks, sessions int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE lower(email) = lower($1) AND deleted_at IS NULL`, emailAddr).Scan(&users))
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM magic_links WHERE lower(email) = lower($1) AND consumed_at IS NOT NULL`, emailAddr).Scan(&consumedLinks))
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM sessions s JOIN users u ON u.id = s.user_id WHERE lower(u.email) = lower($1)`, emailAddr).Scan(&sessions))
	require.Equal(t, 1, users)
	require.Equal(t, 2, consumedLinks)
	require.Equal(t, 2, sessions)
}
