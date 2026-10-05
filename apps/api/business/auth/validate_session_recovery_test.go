package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type sessionValidationRepository struct {
	*MemoryRepository
	sessionErr, userErr error
}

func (r *sessionValidationRepository) GetSessionByTokenHash(ctx context.Context, hash []byte) (*Session, error) {
	if r.sessionErr != nil {
		return nil, r.sessionErr
	}
	return r.MemoryRepository.GetSessionByTokenHash(ctx, hash)
}

func (r *sessionValidationRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if r.userErr != nil {
		return nil, r.userErr
	}
	return r.MemoryRepository.GetUserByID(ctx, id)
}

func TestValidateSessionPreservesStorageFailureAndRecoversSameToken(t *testing.T) {
	for _, stage := range []string{"session", "user"} {
		for _, failure := range []error{errors.New("synthetic storage unavailable"), context.DeadlineExceeded, context.Canceled} {
			t.Run(stage+"/"+failure.Error(), func(t *testing.T) {
				svc, memory, _, _ := testService(t)
				user, err := memory.CreateUser(t.Context(), "validation@example.com", nil)
				require.NoError(t, err)
				_, _, token, err := svc.issueSession(t.Context(), user)
				require.NoError(t, err)
				repo := &sessionValidationRepository{MemoryRepository: memory}
				svc.repo = repo
				if stage == "session" {
					repo.sessionErr = fmt.Errorf("repository: %w", failure)
				} else {
					repo.userErr = fmt.Errorf("repository: %w", failure)
				}
				got, err := svc.ValidateSession(t.Context(), token)
				require.Nil(t, got)
				require.ErrorIs(t, err, failure)
				require.False(t, errors.Is(err, ErrAuthenticationRequired))
				repo.sessionErr, repo.userErr = nil, nil
				got, err = svc.ValidateSession(t.Context(), token)
				require.NoError(t, err)
				require.Equal(t, user.ID, got.ID)
			})
		}
	}
}

func TestValidateSessionStillRejectsWrappedMissingRecords(t *testing.T) {
	for _, stage := range []string{"session", "user"} {
		t.Run(stage, func(t *testing.T) {
			svc, memory, _, _ := testService(t)
			user, err := memory.CreateUser(t.Context(), "missing-record@example.com", nil)
			require.NoError(t, err)
			_, _, token, err := svc.issueSession(t.Context(), user)
			require.NoError(t, err)
			repo := &sessionValidationRepository{MemoryRepository: memory}
			svc.repo = repo
			if stage == "session" {
				repo.sessionErr = fmt.Errorf("lookup: %w", ErrSessionNotFound)
			} else {
				repo.userErr = fmt.Errorf("lookup: %w", ErrUserNotFound)
			}
			got, err := svc.ValidateSession(t.Context(), token)
			require.Nil(t, got)
			require.ErrorIs(t, err, ErrAuthenticationRequired)
		})
	}
}

func TestAuthRepositoriesDistinguishMissingUserFromReadFailure(t *testing.T) {
	for _, by := range []string{"id", "email"} {
		t.Run(by, func(t *testing.T) {
			id := uuid.New()
			memory := NewMemoryRepository()
			if by == "id" {
				_, err := memory.GetUserByID(t.Context(), id)
				require.ErrorIs(t, err, ErrUserNotFound)
			} else {
				_, err := memory.GetUserByEmail(t.Context(), "absent@example.com")
				require.ErrorIs(t, err, ErrUserNotFound)
			}
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			repo := NewPostgreSQLRepository(db)
			lookup := func() error {
				if by == "id" {
					_, err := repo.GetUserByID(t.Context(), id)
					return err
				}
				_, err := repo.GetUserByEmail(t.Context(), "absent@example.com")
				return err
			}
			mock.ExpectQuery("SELECT id, email").WillReturnRows(sqlmock.NewRows([]string{"id"}))
			require.ErrorIs(t, lookup(), ErrUserNotFound)
			failure := errors.New("synthetic query failure")
			mock.ExpectQuery("SELECT id, email").WillReturnError(failure)
			err = lookup()
			require.ErrorIs(t, err, failure)
			require.False(t, errors.Is(err, ErrUserNotFound))
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
