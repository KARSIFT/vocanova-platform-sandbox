package aifeedback

import "context"

// SafetyClassifier is the seam between the orchestrator and the content
// moderation layer. In T01 it may be implemented by the mock provider; in T03 it
// is expanded with local abuse checks and outcome mapping.
type SafetyClassifier interface {
	Classify(ctx context.Context, input ModerationInput) (*ModerationResult, error)
}

// ProviderSafetyClassifier wraps a ModerationProvider to satisfy SafetyClassifier.
type ProviderSafetyClassifier struct {
	provider ModerationProvider
}

// NewProviderSafetyClassifier wraps a moderation provider.
func NewProviderSafetyClassifier(p ModerationProvider) *ProviderSafetyClassifier {
	return &ProviderSafetyClassifier{provider: p}
}

// Classify delegates to the underlying moderation provider.
func (c *ProviderSafetyClassifier) Classify(ctx context.Context, input ModerationInput) (*ModerationResult, error) {
	return c.provider.Classify(ctx, input)
}
