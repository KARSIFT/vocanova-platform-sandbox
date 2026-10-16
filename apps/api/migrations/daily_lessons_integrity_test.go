package migrations_test

import (
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestDailyLessonsMigrationMaintainsOwnershipAndUniqueCredit(t *testing.T) {
	b, e := os.ReadFile("20261017100000_daily_lessons.sql")
	require.NoError(t, e)
	s := string(b)
	for _, invariant := range []string{"UNIQUE(id,user_id,meaning_id)", "PRIMARY KEY(user_id,meaning_id,local_date)", "FOREIGN KEY(review_attempt_id,user_id,meaning_id) REFERENCES review_attempts(id,user_id,meaning_id)", "UNIQUE(review_attempt_id)", "UNIQUE(id,user_id)", "FOREIGN KEY(session_id,user_id) REFERENCES daily_lesson_sessions(id,user_id)", "UNIQUE(user_id,operation,idempotency_key)", "UNIQUE(session_id,operation,client_action_id)", "daily_lesson_one_active", "WHERE status='in_progress'", "CHECK(revision>=0)", "CHECK((status='completed')=(completed_at IS NOT NULL))"} {
		require.Contains(t, s, invariant)
	}
	down, e := os.ReadFile("20261017100000_daily_lessons.down.sql.example")
	require.NoError(t, e)
	require.Less(t, strings.Index(string(down), "DROP TABLE daily_lesson_actions"), strings.Index(string(down), "DROP TABLE daily_lesson_sessions"))
}
