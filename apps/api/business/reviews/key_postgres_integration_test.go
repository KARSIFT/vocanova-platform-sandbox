package reviews

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func reviewKeyDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset; real PostgreSQL unavailable")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	schema := "review_keys_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, err)
	})
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	t.Cleanup(func() { _ = db.Close() })
	paths, err := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, paths)
	for _, path := range paths {
		if strings.HasSuffix(path, ".down.sql") {
			continue
		}
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = db.Exec(string(migration))
		require.NoError(t, err, "migration %s", path)
	}
	return db
}

func seedReviewKeyRequest(t *testing.T, db *sql.DB, now time.Time) SubmitReviewRequest {
	t.Helper()
	userID, wordID, meaningID, userWordID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	_, err := db.Exec(`INSERT INTO users (id, email, created_at, updated_at) VALUES ($1,$2,$3,$3)`, userID, userID.String()+"@example.test", now)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO canonical_words (id,text,normalized_text,status,created_at,updated_at) VALUES ($1,$2,$2,'active',$3,$3)`, wordID, wordID.String(), now)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO word_meanings (id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES ($1,$2,'noun','A fixture.',1,'active',$3,$3)`, meaningID, wordID, now)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO user_words (id,user_id,meaning_id,source,added_at,created_at,updated_at) VALUES ($1,$2,$3,'manual',$4,$4,$4)`, userWordID, userID, meaningID, now)
	require.NoError(t, err)
	return SubmitReviewRequest{UserID: userID, UserWordID: userWordID, MeaningID: meaningID,
		PromptType: PromptTypeSelfCheck, Result: ResultCorrect, Rating: RatingGood,
		AnsweredAt: now, ClientAttemptID: "attempt", IdempotencyKey: "header-key"}
}

func TestReviewKeyTransactionPostgreSQL(t *testing.T) {
	db := reviewKeyDB(t)
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	c := &clock.Fixed{T: time.Now().UTC().Truncate(time.Microsecond)}
	gam := gamification.NewService(gamification.NewRepository(db))
	repo := NewPostgreSQLRepository(db, c, WithGamificationService(gam), WithMissionsService(missions.NewService(missions.NewRepository(db), gam)))
	// No external key store is needed for correctness: the repository owns the
	// claim. A memory store here also ensures tests reach the SQL replay path.
	svc := NewService(repo, nil, c)
	_, err := db.ExecContext(ctx, `CREATE FUNCTION delay_review_claim() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_sleep(0.1); RETURN NEW; END $$; CREATE TRIGGER delay_review_claim BEFORE INSERT ON idempotency_keys FOR EACH ROW EXECUTE FUNCTION delay_review_claim()`)
	require.NoError(t, err)

	for _, tc := range []struct {
		name                    string
		matching, differentWord bool
	}{
		{name: "different attempts"},
		{name: "different words", differentWord: true},
		{name: "matching attempts", matching: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := seedReviewKeyRequest(t, db, c.T)
			other := req
			if tc.differentWord {
				other = seedReviewKeyRequest(t, db, c.T)
				_, err := db.ExecContext(ctx, `UPDATE user_words SET user_id=$1 WHERE id=$2`, req.UserID, other.UserWordID)
				require.NoError(t, err)
				other.UserID = req.UserID
			}
			type outcome struct {
				attempt *ReviewAttempt
				err     error
			}
			start := make(chan struct{})
			results := make(chan outcome, 2)
			for i := range 2 {
				go func() {
					<-start
					r := req
					if !tc.matching && i == 1 {
						r = other
						r.ClientAttemptID = "other-attempt"
					}
					a, err := svc.SubmitReview(ctx, r)
					results <- outcome{a, err}
				}()
			}
			close(start)
			a, b := <-results, <-results
			if tc.matching {
				require.NoError(t, a.err)
				require.NoError(t, b.err)
				require.Equal(t, a.attempt.ID, b.attempt.ID)
			} else {
				require.True(t, (a.err == nil && errors.Is(b.err, ErrIdempotencyConflict)) || (b.err == nil && errors.Is(a.err, ErrIdempotencyConflict)), "results: %v / %v", a.err, b.err)
			}
			for _, table := range []string{"review_attempts", "idempotency_keys", "confidence_point_ledger"} {
				var count int
				require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", req.UserID).Scan(&count))
				require.Equal(t, 1, count, table)
			}
			var step, total, points, reviews int
			require.NoError(t, db.QueryRowContext(ctx, `SELECT sum(review_step),sum(total_review_count) FROM user_words WHERE user_id=$1`, req.UserID).Scan(&step, &total))
			require.Equal(t, 1, step)
			require.Equal(t, 1, total)
			require.NoError(t, db.QueryRowContext(ctx, `SELECT confidence_points_earned,reviews_attempted FROM daily_activity_summaries WHERE user_id=$1`, req.UserID).Scan(&points, &reviews))
			require.Equal(t, 5, points)
			require.Equal(t, 1, reviews)
		})
	}

	t.Run("claim failure cannot publish effects", func(t *testing.T) {
		req := seedReviewKeyRequest(t, db, c.T)
		_, err := db.ExecContext(ctx, `ALTER TABLE idempotency_keys ADD CONSTRAINT reject_fixture_claim CHECK (user_id <> '`+req.UserID.String()+`')`)
		require.NoError(t, err)
		_, err = svc.SubmitReview(ctx, req)
		require.ErrorContains(t, err, "claim review idempotency")
		for _, table := range []string{"review_attempts", "idempotency_keys", "confidence_point_ledger", "daily_activity_summaries", "daily_mission_snapshots"} {
			var count int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", req.UserID).Scan(&count))
			require.Zero(t, count, table)
		}
		_, err = db.ExecContext(ctx, `ALTER TABLE idempotency_keys DROP CONSTRAINT reject_fixture_claim`)
		require.NoError(t, err)
		_, err = svc.SubmitReview(ctx, req)
		require.NoError(t, err)
	})

	t.Run("late effect failure rolls back key and all earlier writes", func(t *testing.T) {
		req := seedReviewKeyRequest(t, db, c.T)
		_, err := db.ExecContext(ctx, `ALTER TABLE confidence_point_ledger ADD CONSTRAINT reject_fixture_reward CHECK (user_id <> '`+req.UserID.String()+`')`)
		require.NoError(t, err)
		_, err = svc.SubmitReview(ctx, req)
		require.ErrorContains(t, err, "grant review point")
		for _, table := range []string{"review_attempts", "idempotency_keys", "confidence_point_ledger", "daily_activity_summaries", "daily_mission_snapshots"} {
			var count int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table+" WHERE user_id=$1", req.UserID).Scan(&count))
			require.Zero(t, count, table)
		}
		var total int
		require.NoError(t, db.QueryRowContext(ctx, `SELECT total_review_count FROM user_words WHERE id=$1`, req.UserWordID).Scan(&total))
		require.Zero(t, total)
		_, err = db.ExecContext(ctx, `ALTER TABLE confidence_point_ledger DROP CONSTRAINT reject_fixture_reward`)
		require.NoError(t, err)
		_, err = svc.SubmitReview(ctx, req)
		require.NoError(t, err, "failed claim must not block a retry")
	})

	t.Run("expiry does not slide and client attempt guard survives", func(t *testing.T) {
		req := seedReviewKeyRequest(t, db, c.T)
		first, err := svc.SubmitReview(ctx, req)
		require.NoError(t, err)
		createdAt := c.T
		c.T = createdAt.Add(23 * time.Hour)
		_, err = svc.SubmitReview(ctx, req)
		require.NoError(t, err)
		var stored time.Time
		require.NoError(t, db.QueryRowContext(ctx, `SELECT created_at FROM idempotency_keys WHERE user_id=$1`, req.UserID).Scan(&stored))
		require.True(t, stored.Equal(createdAt))
		changed := req
		changed.ClientAttemptID = "new-attempt"
		_, err = svc.SubmitReview(ctx, changed)
		require.ErrorIs(t, err, ErrIdempotencyConflict)
		c.T = createdAt.Add(24 * time.Hour)
		second, err := svc.SubmitReview(ctx, changed)
		require.NoError(t, err)
		require.NotEqual(t, first.ID, second.ID)
		req.IdempotencyKey = "new-header-key"
		replayed, err := svc.SubmitReview(ctx, req)
		require.NoError(t, err)
		require.Equal(t, first.ID, replayed.ID)
	})
}
