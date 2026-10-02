package api

import (
	"context"
	"errors"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/users"
	"github.com/google/uuid"
)

type onboardingProfileReader interface {
	GetOnboarding(context.Context, uuid.UUID) (*users.OnboardingProfile, error)
}

// newOnboardingStatusLookup shares the current-user gate across API wiring.
func newOnboardingStatusLookup(reader onboardingProfileReader) OnboardingStatusLookup {
	return func(ctx context.Context, userID uuid.UUID) (string, error) {
		profile, err := reader.GetOnboarding(ctx, userID)
		if profile == nil || (err != nil && !errors.Is(err, users.ErrOnboardingNotFound)) {
			return users.OnboardingStatusNotStarted, nil
		}
		// Grandfathered users can have authoritative status without stored
		// answers. Missing answers alone must not send them back to onboarding.
		switch profile.Status {
		case users.OnboardingStatusNotStarted, users.OnboardingStatusInProgress, users.OnboardingStatusCompleted:
			return profile.Status, nil
		default:
			return users.OnboardingStatusNotStarted, nil
		}
	}
}
