package reviews

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Run against a disposable PostgreSQL instance with VOCANOVA_TEST_POSTGRES_DSN.
// Fixtures use connection-local temporary tables and never touch persistent data.
func TestDuePaginationPostgreSQL(t *testing.T) {
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN is unset; real PostgreSQL pagination test unavailable")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `
	CREATE TEMP TABLE canonical_words (id uuid, text text, normalized_text text);
	CREATE TEMP TABLE word_meanings (id uuid, word_id uuid, part_of_speech text, short_definition text);
	CREATE TEMP TABLE user_words (id uuid, user_id uuid, meaning_id uuid, status text, source text, review_step int, next_review_at timestamptz, deleted_at timestamptz);`)
	if err != nil {
		t.Fatal(err)
	}
	userID, otherID, meaningID, wordID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.ExecContext(ctx, `INSERT INTO canonical_words VALUES ($1, 'work', 'work')`, wordID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO word_meanings VALUES ($1, $2, 'verb', 'do a job')`, meaningID, wordID); err != nil {
		t.Fatal(err)
	}
	for _, mixed := range []bool{false, true} {
		t.Run(map[bool]string{false: "all_null", true: "mixed_tied_dates"}[mixed], func(t *testing.T) {
			if _, err := db.ExecContext(ctx, `TRUNCATE pg_temp.user_words`); err != nil {
				t.Fatal(err)
			}
			past := time.Now().Add(-time.Hour)
			for i := 1; i <= 7; i++ {
				owner := userID
				if i == 7 {
					owner = otherID
				}
				var scheduled any
				if mixed && i > 2 {
					scheduled = past
				}
				if _, err := db.ExecContext(ctx, `INSERT INTO user_words VALUES ($1,$2,$3,'new','discover',0,$4,NULL)`, uuid.UUID{15: byte(i)}, owner, meaningID, scheduled); err != nil {
					t.Fatal(err)
				}
			}
			repo := NewPostgreSQLRepository(db, nil)
			cursor := ""
			var last DueWord
			for page := 0; page < 3; page++ {
				resp, err := repo.ListDueWords(ctx, ListDueWordsRequest{UserID: userID, Limit: 2, AfterCursor: cursor})
				if err != nil {
					t.Fatal(err)
				}
				if len(resp.Items) != 2 || resp.TotalCount != 6 {
					t.Fatalf("page %d: %+v", page, resp)
				}
				for i, item := range resp.Items {
					if item.UserWordID != (uuid.UUID{15: byte(page*2 + i + 1)}) {
						t.Fatalf("unexpected item %s", item.UserWordID)
					}
				}
				if (resp.NextCursor != "") != (page < 2) {
					t.Fatalf("page %d continuation %q", page, resp.NextCursor)
				}
				cursor = resp.NextCursor
				last = resp.Items[1]
			}
			terminal := dueCursor{ID: last.UserWordID}
			if last.NextReviewAt != nil {
				terminal.NextReviewAt = *last.NextReviewAt
			}
			resp, err := repo.ListDueWords(ctx, ListDueWordsRequest{UserID: userID, Limit: 2, AfterCursor: encodeDueCursor(terminal)})
			if err != nil {
				t.Fatal(err)
			}
			if len(resp.Items) != 0 || resp.TotalCount != 6 || resp.NextCursor != "" {
				t.Fatalf("terminal page: %+v", resp)
			}
		})
	}
}
