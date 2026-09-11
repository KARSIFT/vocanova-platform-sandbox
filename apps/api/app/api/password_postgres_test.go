package api

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/password"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/email"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPasswordAuthPostgreSQL exercises password proof, identity, and session
// transactions against the disposable migrated PostgreSQL database used by CI.
func TestPasswordAuthPostgreSQL(t *testing.T) {
	db := newControlledSignupDisposablePostgres(t)
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	c := &clock.Fixed{T: now}
	fake := &email.Fake{}
	allowed := true
	newService := func(environment string) *password.Service {
		return password.NewService(db, fake, c, auth.NewFixedWindowRateLimiter(c, time.Hour, 100), "https://test.example.com", environment, 30*24*time.Hour, func() bool { return true }, func(string) bool { return allowed }, func(string) bool { return true })
	}
	svc := newService("test")
	ctx := context.Background()
	lastToken := func(t *testing.T) string {
		t.Helper()
		msg, ok := fake.Last()
		require.True(t, ok)
		u, err := url.Parse(strings.TrimSpace(msg.BodyText))
		require.NoError(t, err)
		token := u.Query().Get("token")
		require.NotEmpty(t, token)
		return token
	}
	insertGoogleUser := func(t *testing.T, address string) uuid.UUID {
		t.Helper()
		id := uuid.New()
		require.NoError(t, db.QueryRowContext(ctx, `INSERT INTO users(id,email,display_name,status,email_verified_at,created_at,updated_at) VALUES($1,$2,'Google learner','active',$3,$3,$3) RETURNING id`, id, address, c.Now()).Scan(&id))
		_, err := db.ExecContext(ctx, `INSERT INTO external_identities(id,user_id,provider,provider_subject,provider_email,provider_email_verified,created_at,updated_at) VALUES($1,$2,'google',$3,$4,true,$5,$5)`, uuid.New(), id, "google-"+id.String(), address, c.Now())
		require.NoError(t, err)
		return id
	}

	// A registration creates both identity and credential only after proof. Its
	// replay cannot create another account or attach a credential to an existing
	// one, and the resulting login creates a real server session.
	require.NoError(t, svc.RequestSignup(ctx, "127.0.0.1", "learner@example.com", "correct horse battery staple", "Learner"))
	registration := lastToken(t)
	require.NoError(t, svc.VerifySignup(ctx, "127.0.0.1", registration))
	require.ErrorIs(t, svc.VerifySignup(ctx, "127.0.0.1", registration), password.ErrInvalidToken)
	user, session, raw, err := svc.Login(ctx, "127.0.0.1", "learner@example.com", "correct horse battery staple")
	require.NoError(t, err)
	require.NotEmpty(t, raw)
	require.Equal(t, user.ID, session.UserID)
	require.True(t, user.HasPassword)
	before := len(fake.Sent)
	require.NoError(t, svc.RequestSignup(ctx, "127.0.0.1", "learner@example.com", "must never replace this password", "Other"))
	assert.Len(t, fake.Sent, before, "existing-account signup is a generic no-op")

	t.Run("registration proof honours expiry environment and signup policy", func(t *testing.T) {
		require.NoError(t, svc.RequestSignup(ctx, "127.0.0.2", "expired@example.com", "correct horse battery staple", "Expired"))
		expired := lastToken(t)
		c.T = c.Now().Add(16 * time.Minute)
		require.ErrorIs(t, svc.VerifySignup(ctx, "127.0.0.2", expired), password.ErrInvalidToken)
		c.T = now

		require.NoError(t, svc.RequestSignup(ctx, "127.0.0.3", "other-environment@example.com", "correct horse battery staple", "Other"))
		crossEnvironment := lastToken(t)
		require.ErrorIs(t, newService("other").VerifySignup(ctx, "127.0.0.3", crossEnvironment), password.ErrInvalidToken)
		allowed = false
		require.ErrorIs(t, svc.VerifySignup(ctx, "127.0.0.3", crossEnvironment), password.ErrInvalidToken)
		allowed = true
		require.ErrorIs(t, svc.Reset(ctx, "127.0.0.3", crossEnvironment, "another sufficiently long password"), password.ErrInvalidToken)
	})

	t.Run("reset creates a password for Google identities and revokes all access", func(t *testing.T) {
		googleID := insertGoogleUser(t, "google@example.com")
		require.NoError(t, svc.RequestReset(ctx, "127.0.0.4", "google@example.com"))
		first := lastToken(t)
		require.NoError(t, svc.RequestReset(ctx, "127.0.0.5", "google@example.com"))
		second := lastToken(t)
		magicToken, magicHash, err := auth.NewTokenAndHash()
		require.NoError(t, err)
		_ = magicToken
		_, err = db.ExecContext(ctx, `INSERT INTO magic_links(id,user_id,email,token_hash,environment,created_at,expires_at) VALUES($1,$2,$3,$4,'test',$5,$6)`, uuid.New(), googleID, "google@example.com", magicHash, c.Now(), c.Now().Add(15*time.Minute))
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `INSERT INTO sessions(id,user_id,token_hash,created_at,expires_at) VALUES($1,$2,$3,$4,$5)`, uuid.New(), googleID, mustTokenHash(t), c.Now(), c.Now().Add(time.Hour))
		require.NoError(t, err)
		require.NoError(t, svc.Reset(ctx, "127.0.0.5", second, "a newer sufficiently long password"))
		require.ErrorIs(t, svc.Reset(ctx, "127.0.0.4", first, "a newer sufficiently long password"), password.ErrInvalidToken)
		googleUser, _, _, err := svc.Login(ctx, "127.0.0.4", "google@example.com", "a newer sufficiently long password")
		require.NoError(t, err)
		require.Equal(t, googleID, googleUser.ID)
		var identities, revokedResets, consumedResets, revokedSessions, revokedMagic int
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM external_identities WHERE user_id=$1 AND provider='google'`, googleID).Scan(&identities))
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM password_reset_links WHERE user_id=$1 AND revoked_at IS NOT NULL`, googleID).Scan(&revokedResets))
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM password_reset_links WHERE user_id=$1 AND consumed_at IS NOT NULL`, googleID).Scan(&consumedResets))
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM sessions WHERE user_id=$1 AND revoked_at IS NOT NULL`, googleID).Scan(&revokedSessions))
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM magic_links WHERE user_id=$1 AND revoked_at IS NOT NULL`, googleID).Scan(&revokedMagic))
		assert.Equal(t, 1, identities)
		assert.Equal(t, 1, consumedResets)
		assert.GreaterOrEqual(t, revokedResets, 1)
		assert.GreaterOrEqual(t, revokedSessions, 1)
		assert.GreaterOrEqual(t, revokedMagic, 1)
	})

	t.Run("reset rejects changed or deleted identities", func(t *testing.T) {
		changedID := insertGoogleUser(t, "changed@example.com")
		require.NoError(t, svc.RequestReset(ctx, "127.0.0.6", "changed@example.com"))
		changed := lastToken(t)
		_, err := db.ExecContext(ctx, `UPDATE users SET email='changed-new@example.com',updated_at=$2 WHERE id=$1`, changedID, c.Now())
		require.NoError(t, err)
		require.ErrorIs(t, svc.Reset(ctx, "127.0.0.6", changed, "a newer sufficiently long password"), password.ErrInvalidToken)

		deletedID := insertGoogleUser(t, "deleted@example.com")
		require.NoError(t, svc.RequestReset(ctx, "127.0.0.7", "deleted@example.com"))
		deleted := lastToken(t)
		_, err = db.ExecContext(ctx, `UPDATE users SET status='deleted',deleted_at=$2,updated_at=$2 WHERE id=$1`, deletedID, c.Now())
		require.NoError(t, err)
		require.ErrorIs(t, svc.Reset(ctx, "127.0.0.7", deleted, "a newer sufficiently long password"), password.ErrInvalidToken)
	})

	t.Run("same reset proof is consumed exactly once under concurrency", func(t *testing.T) {
		insertGoogleUser(t, "concurrent@example.com")
		require.NoError(t, svc.RequestReset(ctx, "127.0.0.8", "concurrent@example.com"))
		token := lastToken(t)
		results := make(chan error, 2)
		var wg sync.WaitGroup
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				results <- svc.Reset(ctx, "127.0.0.8", token, "a newer sufficiently long password")
			}()
		}
		wg.Wait()
		close(results)
		var successes, invalid int
		for err := range results {
			if err == nil {
				successes++
			} else if err == password.ErrInvalidToken {
				invalid++
			} else {
				t.Fatalf("unexpected concurrent reset error: %v", err)
			}
		}
		assert.Equal(t, 1, successes)
		assert.Equal(t, 1, invalid)
	})
	t.Run("cleanup purges expired proof hashes", func(t *testing.T) {
		require.NoError(t, svc.RequestSignup(ctx, "127.0.0.9", "cleanup@example.com", "correct horse battery staple", "Cleanup"))
		c.T = c.Now().Add(16 * time.Minute)
		require.NoError(t, password.NewCleaner(db, c).Cleanup(ctx))
		var registrations int
		require.NoError(t, db.QueryRow(`SELECT count(*) FROM password_registration_links WHERE lower(email)=lower($1)`, "cleanup@example.com").Scan(&registrations))
		assert.Zero(t, registrations)
	})
}

func mustTokenHash(t *testing.T) []byte {
	t.Helper()
	_, hash, err := auth.NewTokenAndHash()
	require.NoError(t, err)
	return hash
}
