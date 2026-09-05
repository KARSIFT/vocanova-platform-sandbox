package accounts

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestReportPurgeParticipatesInDeletionTransaction(t *testing.T) {
	for _, failDelete := range []bool{false, true} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()
		uid := uuid.New()
		mock.ExpectBegin()
		deletion := mock.ExpectExec("DELETE FROM ai_feedback_quality_review_reports WHERE user_id = \\$1").WithArgs(uid)
		if failDelete {
			deletion.WillReturnError(errors.New("delete failed"))
		} else {
			deletion.WillReturnResult(sqlmock.NewResult(0, 3))
			// Any later failure rolls back removal of reports as well.
			mock.ExpectExec("DELETE FROM ai_feedback_attempts").WithArgs(uid).WillReturnError(errors.New("later failure"))
		}
		mock.ExpectRollback()
		_, err = NewPostgreSQLRepository(db).AnonymizeUserData(t.Context(), uid)
		require.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	}
}
