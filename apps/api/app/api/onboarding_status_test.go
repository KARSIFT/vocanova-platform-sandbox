package api

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type onboardingProfileReaderFunc func(context.Context, uuid.UUID) (*users.OnboardingProfile, error)

func (f onboardingProfileReaderFunc) GetOnboarding(ctx context.Context, id uuid.UUID) (*users.OnboardingProfile, error) {
	return f(ctx, id)
}

func TestOnboardingStatusLookupPreservesAuthoritativeStatus(t *testing.T) {
	for _, tc := range []struct {
		name    string
		profile *users.OnboardingProfile
		err     error
		want    string
	}{
		{"grandfathered completed without answers", &users.OnboardingProfile{Status: "completed"}, users.ErrOnboardingNotFound, "completed"},
		{"in progress without answers", &users.OnboardingProfile{Status: "in_progress"}, users.ErrOnboardingNotFound, "in_progress"},
		{"wrapped missing answers", &users.OnboardingProfile{Status: "completed"}, fmt.Errorf("lookup: %w", users.ErrOnboardingNotFound), "completed"},
		{"new learner without answers", &users.OnboardingProfile{Status: "not_started"}, users.ErrOnboardingNotFound, "not_started"},
		{"completed profile", &users.OnboardingProfile{Status: "completed"}, nil, "completed"},
		{"in progress profile", &users.OnboardingProfile{Status: "in_progress"}, nil, "in_progress"},
		{"new learner profile", &users.OnboardingProfile{Status: "not_started"}, nil, "not_started"},
		{"missing user", nil, users.ErrUserNotFound, "not_started"},
		{"missing profile", nil, users.ErrOnboardingNotFound, "not_started"},
		{"nil profile without error", nil, nil, "not_started"},
		{"empty status", &users.OnboardingProfile{}, nil, "not_started"},
		{"unknown status", &users.OnboardingProfile{Status: "unexpected"}, nil, "not_started"},
		{"unknown status without answers", &users.OnboardingProfile{Status: "unexpected"}, users.ErrOnboardingNotFound, "not_started"},
		{"real error despite completed status", &users.OnboardingProfile{Status: "completed"}, errors.New("database unavailable"), "not_started"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := uuid.New()
			ctx := context.Background()
			lookup := newOnboardingStatusLookup(onboardingProfileReaderFunc(func(gotCtx context.Context, gotID uuid.UUID) (*users.OnboardingProfile, error) {
				require.Equal(t, ctx, gotCtx)
				require.Equal(t, id, gotID)
				return tc.profile, tc.err
			}))
			require.NotPanics(t, func() {
				status, err := lookup(ctx, id)
				require.NoError(t, err)
				require.Equal(t, tc.want, status)
			})
		})
	}
}
