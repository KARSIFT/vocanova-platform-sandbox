package aifeedback

import (
	"context"
	"strings"
)

// SafetyResult is the internal outcome of the safety/moderation pipeline. It
// extends the provider moderation outcome with a crisis resource message for
// self-harm interventions. The Outcome field is one of the Safety* constants.
type SafetyResult struct {
	Outcome               string
	Reason                string
	CrisisResourceMessage string
}

// SafetyClassifier is the seam between the orchestrator and the content
// moderation layer. In T01 it may be implemented by the mock provider; in T03 it
// is expanded with local abuse checks and outcome mapping.
type SafetyClassifier interface {
	Classify(ctx context.Context, input ModerationInput) (*SafetyResult, error)
}

// LocalAbuseChecker performs lightweight, deterministic local abuse checks
// before any provider moderation call. It must run quickly, never call an
// external model, and never treat prompt-injection attempts as abuse.
type LocalAbuseChecker interface {
	Check(ctx context.Context, input ModerationInput) *SafetyResult
}

// DefaultLocalAbuseChecker is the lightweight local abuse check implementation.
// It uses conservative word lists and explicit allow-patterns to avoid blocking
// legitimate discussion of difficult subjects or prompt-injection attempts.
type DefaultLocalAbuseChecker struct {
	blockedChecks  []patternCheck
	selfHarmChecks []patternCheck
	allowPatterns  []string
}

// NewDefaultLocalAbuseChecker creates a local abuse checker with conservative
// built-in patterns.
func NewDefaultLocalAbuseChecker() *DefaultLocalAbuseChecker {
	return &DefaultLocalAbuseChecker{
		blockedChecks: []patternCheck{
			{category: "weapon_instruction", patterns: []string{"how to make a bomb", "build a bomb", "make a weapon", "create a weapon"}},
			{category: "dangerous_substance", patterns: []string{"how to make poison", "how to make drugs", "how to make dangerous substances"}},
			{category: "sexual_exploitation_minor", patterns: []string{"sexual exploitation of minors", "child sexual abuse", "sexualize minors", "minor pornography"}},
			{category: "credible_threat", patterns: []string{"i will kill you", "i am going to kill", "i'm going to kill", "i will hurt you", "i'm going to hurt"}},
			{category: "hateful_incitement", patterns: []string{"kill all", "death to", "exterminate"}},
			{category: "harassment", patterns: []string{"dox them", "post their address", "post their phone number"}},
		},
		selfHarmChecks: []patternCheck{
			{category: "self_harm_urge", patterns: []string{"kill myself", "kill himself", "kill herself", "kill themself", "kill themselves", "want to die", "i want to die", "i want to end my life", "commit suicide", "self-harm", "self harm"}},
		},
		allowPatterns: []string{
			"work with people who self-harm",
			"work with people who self harm",
			"help people who self-harm",
			"help people who self harm",
			"support people who self-harm",
			"support people who self harm",
		},
	}
}

// Check returns a SafetyResult if the input matches a local abuse pattern.
// It returns nil if no local pattern matches, allowing the provider moderation
// path to run. Prompt-injection attempts and legitimate discussion of difficult
// subjects are intentionally not blocked here.
func (c *DefaultLocalAbuseChecker) Check(ctx context.Context, input ModerationInput) *SafetyResult {
	normalized := strings.ToLower(strings.TrimSpace(input.SentenceText))
	if normalized == "" {
		return nil
	}

	for _, allowed := range c.allowPatterns {
		if strings.Contains(normalized, allowed) {
			return nil
		}
	}

	for _, check := range c.selfHarmChecks {
		if check.matches(normalized) {
			return &SafetyResult{
				Outcome:               SafetySelfHarmIntervention,
				Reason:                "local " + check.category,
				CrisisResourceMessage: CrisisResourceText,
			}
		}
	}

	for _, check := range c.blockedChecks {
		if check.matches(normalized) {
			return &SafetyResult{
				Outcome: SafetyBlocked,
				Reason:  "local " + check.category,
			}
		}
	}

	return nil
}

type patternCheck struct {
	category string
	patterns []string
}

func (pc patternCheck) matches(text string) bool {
	for _, p := range pc.patterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

// CompositeSafetyClassifier runs deterministic local abuse checks first, then
// calls the provider moderation layer when required. Provider errors are mapped
// to SafetyModerationUnavailable so the service can return a safe retryable
// failure.
type CompositeSafetyClassifier struct {
	local    LocalAbuseChecker
	provider ModerationProvider
}

// NewCompositeSafetyClassifier creates a safety classifier that combines local
// checks and provider moderation.
func NewCompositeSafetyClassifier(local LocalAbuseChecker, provider ModerationProvider) *CompositeSafetyClassifier {
	if local == nil {
		local = NewDefaultLocalAbuseChecker()
	}
	return &CompositeSafetyClassifier{
		local:    local,
		provider: provider,
	}
}

// Classify runs local checks first. If no local match is found, it delegates to
// the provider moderation provider. Provider errors are mapped to
// SafetyModerationUnavailable.
func (c *CompositeSafetyClassifier) Classify(ctx context.Context, input ModerationInput) (*SafetyResult, error) {
	if local := c.local.Check(ctx, input); local != nil {
		return local, nil
	}

	if c.provider == nil {
		return &SafetyResult{Outcome: SafetyModerationUnavailable, Reason: "no moderation provider configured"}, nil
	}

	result, err := c.provider.Classify(ctx, input)
	if err != nil {
		return &SafetyResult{Outcome: SafetyModerationUnavailable, Reason: "provider moderation unavailable"}, nil
	}
	if result == nil {
		return &SafetyResult{Outcome: SafetyAllowed, Reason: "provider returned nil"}, nil
	}

	out := &SafetyResult{
		Outcome: result.Outcome,
		Reason:  result.Reason,
	}
	if result.Outcome == SafetySelfHarmIntervention {
		out.CrisisResourceMessage = CrisisResourceText
	}
	return out, nil
}

// ProviderSafetyClassifier wraps a ModerationProvider to satisfy SafetyClassifier.
type ProviderSafetyClassifier struct {
	provider ModerationProvider
}

// NewProviderSafetyClassifier wraps a moderation provider. It converts the
// ModerationProvider result into a SafetyResult. Provider errors are mapped to
// SafetyModerationUnavailable.
func NewProviderSafetyClassifier(p ModerationProvider) *ProviderSafetyClassifier {
	return &ProviderSafetyClassifier{provider: p}
}

// Classify delegates to the underlying moderation provider and maps the result.
func (c *ProviderSafetyClassifier) Classify(ctx context.Context, input ModerationInput) (*SafetyResult, error) {
	if c.provider == nil {
		return &SafetyResult{Outcome: SafetyModerationUnavailable, Reason: "no moderation provider configured"}, nil
	}

	result, err := c.provider.Classify(ctx, input)
	if err != nil {
		return &SafetyResult{Outcome: SafetyModerationUnavailable, Reason: "provider moderation unavailable"}, nil
	}
	if result == nil {
		return &SafetyResult{Outcome: SafetyAllowed, Reason: "provider returned nil"}, nil
	}

	out := &SafetyResult{
		Outcome: result.Outcome,
		Reason:  result.Reason,
	}
	if result.Outcome == SafetySelfHarmIntervention {
		out.CrisisResourceMessage = CrisisResourceText
	}
	return out, nil
}
