package aifeedback

import (
	"context"
	"errors"
	"time"
)

const EvaluationScopeAdapterOnly = "adapter_only"

// EvaluationCaseEvidence preserves observations for every input, including
// failures and unscored fixtures. Nil measurements mean unmeasured, never zero.
// ProviderAttempts counts logical adapter calls, not hidden HTTP retries.
type EvaluationCaseEvidence struct {
	InputIndex                int
	Case                      EvaluationCase
	Outcome                   string
	Validation                ValidationResult
	ProviderAttempts          int
	ProviderReturned          bool
	ProviderErrorCode         string
	StructuredOutputValid     *bool
	StructuredOutputErrorCode string
	Feedback                  map[string]any
	ElapsedMillis             int64
	ProviderElapsedMillis     int64
	TransportAttempts         *int
	RepairAttempts            *int
	SafetyChecked             bool
	PersistenceConfirmed      *bool
	ExpectedOutcomeMatched    *bool
	StatusMatched             *bool
	HumanReviewStatus         string
}

// EvaluationResult is the outcome of running the dataset against a provider.
type EvaluationResult struct {
	DatasetVersion  string
	Scope           string
	CaseEvidence    []EvaluationCaseEvidence
	Total           int
	Validated       int
	ProviderCalled  int
	MatchedStatus   int
	MismatchedCases []EvaluationMismatch
	ByCategory      map[string]int
	ByStatus        map[string]int
}

// EvaluationMismatch records a case whose provider status differed from the
// expected status.
type EvaluationMismatch struct {
	Case             EvaluationCase
	GotStatus        string
	ValidationFailed bool
}

// EvaluationRun captures the metadata that must be recorded for every material
// AI change (DOC-09 §23).
type EvaluationRun struct {
	DatasetVersion   string
	GoldenSetVersion string
	PromptVersion    string
	SchemaVersion    string
	Provider         string
	Model            string
	Config           string
	Commit           string
	Timestamp        string
	Result           EvaluationResult
	LatencyMs        int64
	CostCents        int
	CriticalFailures []string
	Reviewer         string
}

// RunEvaluation observes deterministic validation, a single logical adapter
// call, and the existing output validator. It does not run service moderation,
// repair or persistence, or observe an adapter's internal transport retries.
func RunEvaluation(ctx context.Context, provider FeedbackProvider, cases []EvaluationCase) EvaluationResult {
	result := EvaluationResult{
		DatasetVersion: DatasetVersion, Scope: EvaluationScopeAdapterOnly,
		CaseEvidence: make([]EvaluationCaseEvidence, 0, len(cases)),
		ByCategory:   make(map[string]int), ByStatus: make(map[string]int),
	}
	builder := NewDefaultTaskBuilder()
	for index, c := range cases {
		started := time.Now()
		result.Total++
		result.ByCategory[c.Category]++
		evidence := EvaluationCaseEvidence{InputIndex: index, Case: c, HumanReviewStatus: "not_reviewed"}
		target := &Target{
			NormalizedWord: c.TargetWord, WordType: c.WordType,
			PartOfSpeech: c.PartOfSpeech, LearnerLevel: c.LearnerLevel,
			ShortDefinition: c.TargetMeaning,
			AcceptedForms:   BuildAcceptedForms(c.TargetWord, c.WordType, c.PartOfSpeech),
		}
		evidence.Validation = ValidateSentence(c.Sentence, target)
		if !evidence.Validation.Valid {
			evidence.Outcome = "validation_failed"
		} else {
			result.Validated++
			evidence.Outcome = "provider_error"
			if ctx.Err() != nil {
				evidence.Outcome = "not_run"
				evidence.ProviderErrorCode = evaluationProviderErrorCode(ctx.Err())
			} else if provider == nil {
				evidence.ProviderErrorCode = "provider_unavailable"
			} else {
				task := builder.Build(target, evidence.Validation.Normalized)
				evidence.ProviderAttempts = 1
				result.ProviderCalled++
				callStarted := time.Now()
				feedback, err := provider.GenerateFeedback(ctx, task)
				evidence.ProviderElapsedMillis = time.Since(callStarted).Milliseconds()
				// Preserve only named feedback fields, never raw provider envelopes.
				if feedback != nil {
					evidence.ProviderReturned = true
					evidence.Feedback = feedback.StructuredJSON()
				}
				if err != nil {
					evidence.ProviderErrorCode = evaluationProviderErrorCode(err)
				} else {
					valid := NewDefaultOutputValidator().Validate(feedback, target) == nil
					evidence.StructuredOutputValid = &valid
					if valid {
						evidence.Outcome = "feedback"
					} else {
						evidence.Outcome = "invalid_output"
						evidence.StructuredOutputErrorCode = "output_validation_failed"
					}
				}
			}
		}
		expectedOutcome := c.ExpectedOutcome
		if expectedOutcome == "" {
			expectedOutcome = "feedback"
		}
		// This adapter-only runner cannot measure an expected service safety
		// intervention. Keep the observed adapter outcome without inventing one.
		if expectedOutcome != "safety_intercept" {
			matched := evidence.Outcome == expectedOutcome
			evidence.ExpectedOutcomeMatched = &matched
		}
		gotStatus := evidence.Outcome
		if evidence.Outcome == "feedback" {
			gotStatus, _ = evidence.Feedback["status"].(string)
		}
		result.ByStatus[gotStatus]++
		if c.ExpectedStatus != "" && expectedOutcome == "feedback" && c.ScoringExclusionReason == "" {
			matched := evidence.Outcome == "feedback" && gotStatus == c.ExpectedStatus
			evidence.StatusMatched = &matched
			if matched {
				result.MatchedStatus++
			} else {
				result.MismatchedCases = append(result.MismatchedCases, EvaluationMismatch{
					Case: c, GotStatus: gotStatus, ValidationFailed: !evidence.Validation.Valid,
				})
			}
		}
		evidence.ElapsedMillis = time.Since(started).Milliseconds()
		result.CaseEvidence = append(result.CaseEvidence, evidence)
	}
	return result
}

// Error strings may include request URLs, account IDs, or provider bodies.
// Only stable categories cross the report boundary.
func evaluationProviderErrorCode(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, ErrProviderTimeout):
		return "timeout"
	case errors.Is(err, ErrProviderAuth):
		return "authentication"
	case errors.Is(err, ErrProviderRefusal):
		return "refusal"
	case errors.Is(err, ErrProviderInvalidInput), errors.Is(err, ErrMissingLearnerSentence):
		return "invalid_input"
	case errors.Is(err, ErrProviderInvalidResponse):
		return "invalid_response"
	case errors.Is(err, ErrRateLimited):
		return "rate_limited"
	default:
		return "provider_error"
	}
}

// RunMockEvaluation runs the initial dataset against the deterministic mock
// provider. It is safe for CI because it never calls a paid provider.
func RunMockEvaluation(ctx context.Context) EvaluationResult {
	return RunEvaluation(ctx, NewMockProvider(), InitialDataset())
}

// RunGoldenEvaluation runs the golden regression set against the deterministic
// mock provider.
func RunGoldenEvaluation(ctx context.Context) EvaluationResult {
	return RunEvaluation(ctx, NewMockProvider(), GoldenSet())
}
