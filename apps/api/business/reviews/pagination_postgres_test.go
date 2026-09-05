package reviews

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestPostgresDuePaginationLookahead(t *testing.T) {
	for _, returned := range []int{2, 3} {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		userID := uuid.New()
		rows := sqlmock.NewRows([]string{"user_word_id", "meaning_id", "word_id", "text", "normalized_text", "part_of_speech", "short_definition", "status", "source", "review_step", "next_review_at", "total_count"})
		for i := 1; i <= returned; i++ {
			rows.AddRow(uuid.UUID{15: byte(i)}, uuid.New(), uuid.New(), "work", "work", "verb", "do a job", "new", "discover", 0, nil, 6)
		}
		mock.ExpectQuery("SELECT \\* FROM").WithArgs(userID, nil, uuid.Nil, 3).WillReturnRows(rows)
		resp, err := NewPostgreSQLRepository(db, nil).ListDueWords(context.Background(), ListDueWordsRequest{UserID: userID, Limit: 2})
		if err != nil {
			t.Fatal(err)
		}
		if len(resp.Items) != 2 || resp.TotalCount != 6 {
			t.Fatalf("unexpected page: %+v", resp)
		}
		if (resp.NextCursor != "") != (returned > 2) {
			t.Fatalf("lookahead=%d, cursor=%q", returned, resp.NextCursor)
		}
		if returned > 2 {
			cursor, err := decodeDueCursor(resp.NextCursor)
			if err != nil {
				t.Fatal(err)
			}
			if cursor.ID != (uuid.UUID{15: 2}) {
				t.Fatalf("cursor skipped a row: %+v", cursor)
			}
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPostgresDuePaginationExhaustedCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	userID, lastID := uuid.New(), uuid.New()
	mock.ExpectQuery("SELECT \\* FROM").WithArgs(userID, nil, lastID, 3).WillReturnRows(sqlmock.NewRows([]string{"user_word_id"}))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM user_words").WithArgs(userID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(6))
	resp, err := NewPostgreSQLRepository(db, nil).ListDueWords(context.Background(), ListDueWordsRequest{UserID: userID, Limit: 2, AfterCursor: encodeDueCursor(dueCursor{ID: lastID})})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 0 || resp.NextCursor != "" || resp.TotalCount != 6 {
		t.Fatalf("unexpected exhausted result: %+v", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
