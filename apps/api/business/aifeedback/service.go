package aifeedback

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/google/uuid"
)

// ServiceConfig is the runtime configuration for the AI feedback service.
type ServiceConfig struct {
	Provider  string
	Model     string
	Release   string
	RateLimit RateLimitConfig
	OpenCode  OpenCodeConfig
	Gate      GenerationGate
	Metrics   MetricsRecorder
}

// DefaultServiceConfig returns the default P3 configuration.
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		Provider:  ProviderMock,
		Model:     "mock",
		Release:   "unknown",
		RateLimit: DefaultRateLimitConfig(),
		OpenCode:  DefaultOpenCodeConfig(),
		Gate:      NewAlwaysEnabledGate(),
		Metrics:   NewNoopMetricsRecorder(),
	}
}

// OpenCodeServiceConfig returns a service configuration wired for the OpenCode
// production provider. BaseURL and APIKey must still be supplied from
// backend-only secrets.
func OpenCodeServiceConfig() ServiceConfig {
	cfg := DefaultServiceConfig()
	cfg.Provider = ProviderOpenCode
	cfg.Model = DefaultOpenCodeModel
	cfg.OpenCode.Model = DefaultOpenCodeModel
	return cfg
}

// Service orchestrates the sentence-feedback lifecycle.
type Service struct {
	repo            Repository
	provider        FeedbackProvider
	safety          SafetyClassifier
	rateLimiter     RateLimiter
	idem            learning.IdempotencyStore
	mission         MissionUpdater
	telemetry       TelemetryRecorder
	taskBuilder     TaskBuilder
	outputValidator OutputValidator
	clock           clock.Clock
	config          ServiceConfig
}

const operationSentenceFeedback = "ai_feedback_request"

// NewService creates a new AI feedback service.
func NewService(
	repo Repository,
	provider FeedbackProvider,
	safety SafetyClassifier,
	rateLimiter RateLimiter,
	idem learning.IdempotencyStore,
	mission MissionUpdater,
	telemetry TelemetryRecorder,
	taskBuilder TaskBuilder,
	outputValidator OutputValidator,
	c clock.Clock,
	config ServiceConfig,
) *Service {
	if c == nil {
		c = clock.Real{}
	}
	if rateLimiter == nil {
		rateLimiter = NewMemoryRateLimiter(config.RateLimit, c)
	}
	if safety == nil {
		safety = NewCompositeSafetyClassifier(NewDefaultLocalAbuseChecker(), nil)
	}
	if mission == nil {
		mission = NewStubMissionUpdater()
	}
	if telemetry == nil {
		telemetry = NewNoopTelemetryRecorder()
	}
	if taskBuilder == nil {
		taskBuilder = NewDefaultTaskBuilder()
	}
	if outputValidator == nil {
		outputValidator = NewDefaultOutputValidator()
	}
	if config.Gate == nil {
		config.Gate = NewAlwaysEnabledGate()
	}
	if config.Metrics == nil {
		config.Metrics = NewNoopMetricsRecorder()
	}
	return &Service{
		repo:            repo,
		provider:        provider,
		safety:          safety,
		rateLimiter:     rateLimiter,
		idem:            idem,
		mission:         mission,
		telemetry:       telemetry,
		taskBuilder:     taskBuilder,
		outputValidator: outputValidator,
		clock:           c,
		config:          config,
	}
}

// SubmitSentenceFeedback runs the full DOC-09 §17 lifecycle against the
// configured provider. It returns a SentenceFeedbackResult for success, failure,
// validation, rate-limit, or safety outcomes.
func (s *Service) SubmitSentenceFeedback(ctx context.Context, req SubmitSentenceFeedbackRequest) (*SentenceFeedbackResult, error) {
	if req.UserID == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if req.SentenceText == "" {
		return s.validationResult(req.SentenceText, ValidationCodeInvalidInput), nil
	}
	if req.AttemptID == uuid.Nil {
		return s.validationResult(req.SentenceText, ValidationCodeAttemptNotEligible), nil
	}
	if req.IdempotencyKey == "" {
		return nil, errors.New("idempotency key required")
	}

	if err := s.config.Gate.Check(ctx, req.UserID); err != nil {
		if errors.Is(err, ErrAIGenerationDisabled) {
			s.recordTelemetry(ctx, req.UserID, nil, "generation_disabled", 0, "")
			return &SentenceFeedbackResult{
				OriginalSentence: req.SentenceText,
				ErrorCode:        ErrorCodeAIGenerationDisabled,
				CanRetry:         true,
			}, nil
		}
		s.recordTelemetry(ctx, req.UserID, nil, "global_rate_limited", 0, "")
		return &SentenceFeedbackResult{
			OriginalSentence: req.SentenceText,
			ErrorCode:        ErrorCodeRateLimited,
			CanRetry:         true,
		}, nil
	}

	if err := s.rateLimiter.Allow(ctx, req.UserID); err != nil {
		defer s.rateLimiter.Release(req.UserID)
		return &SentenceFeedbackResult{
			OriginalSentence: req.SentenceText,
			ErrorCode:        ErrorCodeRateLimited,
			CanRetry:         true,
		}, nil
	}
	defer s.rateLimiter.Release(req.UserID)

	start := s.clock.Now()

	target, err := s.repo.LoadTarget(ctx, LoadTargetRequest{UserID: req.UserID, Source: req.Source, AttemptID: req.AttemptID})
	if err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			s.recordTelemetry(ctx, req.UserID, target, "attempt_not_eligible", 0, "")
			return s.validationResult(req.SentenceText, ValidationCodeAttemptNotEligible), nil
		}
		return nil, fmt.Errorf("load target: %w", err)
	}

	validation := ValidateSentence(req.SentenceText, target)
	if !validation.Valid {
		s.recordTelemetry(ctx, req.UserID, target, "validation_failed", 0, "")
		return s.validationResult(req.SentenceText, validation.Code), nil
	}

	requestHash := RequestHash(req.UserID, req.AttemptID, target.NormalizedWord, validation.Normalized, PromptVersionSentenceFeedbackV1)

	idempotencyStatus, err := s.idem.Check(ctx, req.UserID, operationSentenceFeedback, req.IdempotencyKey, requestHash)
	if err != nil {
		return nil, fmt.Errorf("idempotency check: %w", err)
	}
	if idempotencyStatus == learning.IdempotencyConflict {
		return &SentenceFeedbackResult{
			OriginalSentence: req.SentenceText,
			ErrorCode:        ErrorCodeIdempotencyConflict,
			CanRetry:         false,
		}, nil
	}

	existing, err := s.repo.GetFeedbackAttemptByRequestHash(ctx, requestHash)
	if err != nil {
		return nil, fmt.Errorf("dedup check: %w", err)
	}
	if existing != nil {
		return s.resultFromStored(existing, req.SentenceText), nil
	}

	moderation, err := s.safety.Classify(ctx, ModerationInput{
		SentenceText: validation.Normalized,
		TargetWord:   target.NormalizedWord,
		LearnerLevel: target.LearnerLevel,
	})
	if err != nil {
		return nil, fmt.Errorf("safety classification: %w", err)
	}
	if moderation == nil {
		moderation = &SafetyResult{Outcome: SafetyAllowed}
	}

	switch moderation.Outcome {
	case SafetyAllowed, SafetyAllowedSensitive:
		// proceed
	case SafetyBlocked:
		s.recordTelemetry(ctx, req.UserID, target, "safety_blocked", 0, "")
		return &SentenceFeedbackResult{
			OriginalSentence: req.SentenceText,
			ErrorCode:        ErrorCodeSafetyBlocked,
			CanRetry:         false,
		}, nil
	case SafetySelfHarmIntervention:
		s.recordTelemetry(ctx, req.UserID, target, "safety_self_harm", 0, "")
		return &SentenceFeedbackResult{
			OriginalSentence:      req.SentenceText,
			ErrorCode:             ErrorCodeSafetySelfHarm,
			CanRetry:              false,
			CrisisResourceMessage: moderation.CrisisResourceMessage,
		}, nil
	case SafetyModerationUnavailable:
		s.recordTelemetry(ctx, req.UserID, target, "safety_moderation_unavailable", 0, "")
		return &SentenceFeedbackResult{
			OriginalSentence: req.SentenceText,
			ErrorCode:        ErrorCodeSafetyModerationUnavailable,
			CanRetry:         true,
		}, nil
	default:
		return &SentenceFeedbackResult{
			OriginalSentence: req.SentenceText,
			ErrorCode:        ErrorCodeSafetyBlocked,
			CanRetry:         false,
		}, nil
	}

	now := s.clock.Now().UTC()
	pending, err := s.repo.CreatePendingAttempt(ctx, req, target, validation.Normalized, requestHash, s.config.Provider, s.config.Model, now)
	if err != nil {
		return nil, fmt.Errorf("create pending attempt: %w", err)
	}

	if err := s.idem.Record(ctx, req.UserID, operationSentenceFeedback, req.IdempotencyKey, requestHash); err != nil {
		return nil, fmt.Errorf("record idempotency: %w", err)
	}

	feedback, providerDuration, providerErr := s.generateWithRepair(ctx, target, validation.Normalized)

	if providerErr != nil {
		if err := s.repo.CompleteFeedbackAttempt(ctx, *pending, nil, ErrorCodeTemporaryFailure, providerErr.Error(), s.clock.Now().UTC()); err != nil {
			return nil, fmt.Errorf("finalize failed attempt: %w", err)
		}
		s.recordTelemetry(ctx, req.UserID, target, "provider_error", providerDuration.Milliseconds(), "")
		return &SentenceFeedbackResult{
			SentenceID:       pending.SentenceID,
			AttemptID:        pending.AttemptID,
			OriginalSentence: req.SentenceText,
			ErrorCode:        ErrorCodeTemporaryFailure,
			CanRetry:         true,
		}, nil
	}

	if err := s.outputValidator.Validate(feedback, target); err != nil {
		if err := s.repo.CompleteFeedbackAttempt(ctx, *pending, nil, ErrorCodeTemporaryFailure, err.Error(), s.clock.Now().UTC()); err != nil {
			return nil, fmt.Errorf("finalize invalid output attempt: %w", err)
		}
		s.recordTelemetry(ctx, req.UserID, target, "invalid_output", providerDuration.Milliseconds(), "")
		return &SentenceFeedbackResult{
			SentenceID:       pending.SentenceID,
			AttemptID:        pending.AttemptID,
			OriginalSentence: req.SentenceText,
			ErrorCode:        ErrorCodeTemporaryFailure,
			CanRetry:         true,
		}, nil
	}

	if err := s.repo.CompleteFeedbackAttempt(ctx, *pending, feedback, "", "", s.clock.Now().UTC()); err != nil {
		return nil, fmt.Errorf("finalize successful attempt: %w", err)
	}

	missionCompleted, err := s.mission.Update(ctx, req.UserID, pending.SentenceID)
	if err != nil {
		return nil, fmt.Errorf("mission update: %w", err)
	}

	duration := s.clock.Now().Sub(start)
	s.recordTelemetry(ctx, req.UserID, target, "success", duration.Milliseconds(), feedback.Status)

	result := &SentenceFeedbackResult{
		SentenceID:        pending.SentenceID,
		AttemptID:         pending.AttemptID,
		Status:            feedback.Status,
		OriginalSentence:  req.SentenceText,
		CorrectedSentence: feedback.CorrectedSentence,
		Explanation:       feedback.Explanation,
		ImprovementTip:    feedback.ImprovementTip,
		MissionCompleted:  missionCompleted,
		CanRetry:          false,
		Reported:          false,
	}
	return result, nil
}

// ReportFeedback records a learner report for a feedback attempt. It verifies
// the attempt belongs to the authenticated learner, then emits a privacy-safe
// telemetry report. It does not change the stored result or mission completion.
func (s *Service) ReportFeedback(ctx context.Context, userID, attemptID uuid.UUID, reason, idempotencyKey string) error {
	if userID == uuid.Nil {
		return errors.New("user id required")
	}
	if attemptID == uuid.Nil {
		return ErrTargetNotFound
	}
	if !validReportReason(reason) {
		return ErrInvalidReportReason
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		return errors.New("idempotency key required")
	}
	owner, err := s.repo.GetFeedbackAttemptOwner(ctx, attemptID)
	if err != nil {
		return fmt.Errorf("lookup attempt owner: %w", err)
	}
	if owner != userID {
		return ErrTargetNotFound
	}
	fingerprint := fmt.Sprintf("%s|%s", attemptID, reason)
	const operation = "report_sentence_feedback"
	status, err := s.idem.Check(ctx, userID, operation, idempotencyKey, fingerprint)
	if err != nil {
		return fmt.Errorf("check report idempotency: %w", err)
	}
	if status == learning.IdempotencyConflict {
		return ErrReportIdempotencyConflict
	}
	created, err := s.repo.CreateQualityReviewReport(ctx, QualityReviewReport{
		ID: uuid.New(), AttemptID: attemptID, UserID: userID, Reason: reason,
		State: QualityReviewStateOpen, CreatedAt: s.clock.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("create quality review report: %w", err)
	}
	if err := s.idem.Record(ctx, userID, operation, idempotencyKey, fingerprint); err != nil {
		return fmt.Errorf("record report idempotency: %w", err)
	}
	if !created {
		return nil
	}
	s.telemetry.RecordReport(ctx, FeedbackReport{
		UserID:         userID,
		AttemptID:      attemptID,
		Reason:         reason,
		Classification: "",
	})
	s.config.Metrics.RecordReport(ctx, MetricsReportEvent{
		Classification: "",
		Provider:       s.config.Provider,
		Model:          s.config.Model,
		Release:        s.config.Release,
		Count:          1,
	})
	return nil
}

// generateWithRepair calls the provider once and, if the output fails validation,
// makes one constrained repair attempt (DOC-09 §10). The provider call is bounded
// by the DOC-09 §18 total backend target of 10 seconds; the adapter itself
// enforces an 8-second per-request timeout.
func (s *Service) generateWithRepair(ctx context.Context, target *Target, normalized string) (*ProviderFeedback, time.Duration, error) {
	providerCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	task := s.taskBuilder.Build(target, normalized)
	providerStart := s.clock.Now()
	feedback, err := s.provider.GenerateFeedback(providerCtx, task)
	providerDuration := s.clock.Now().Sub(providerStart)

	if err != nil {
		return nil, providerDuration, err
	}

	validationErr := s.outputValidator.Validate(feedback, target)
	if validationErr == nil {
		return feedback, providerDuration, nil
	}

	repairTask := s.taskBuilder.BuildRepair(task, validationErr.Error(), feedback.RawJSON)
	repairStart := s.clock.Now()
	feedback, err = s.provider.GenerateFeedback(providerCtx, repairTask)
	providerDuration += s.clock.Now().Sub(repairStart)

	if err != nil {
		return nil, providerDuration, err
	}

	if err := s.outputValidator.Validate(feedback, target); err != nil {
		return nil, providerDuration, fmt.Errorf("output validation failed after repair: %w", err)
	}
	return feedback, providerDuration, nil
}

func (s *Service) validationResult(original, code string) *SentenceFeedbackResult {
	return &SentenceFeedbackResult{
		OriginalSentence: original,
		ErrorCode:        code,
		CanRetry:         code != ValidationCodeAttemptNotEligible,
	}
}

func (s *Service) resultFromStored(attempt *StoredFeedbackAttempt, original string) *SentenceFeedbackResult {
	result := &SentenceFeedbackResult{
		SentenceID:       attempt.LearnerSentenceID,
		AttemptID:        attempt.ID,
		OriginalSentence: original,
		CanRetry:         false,
		Reported:         attempt.Reported,
	}

	switch attempt.Status {
	case AttemptStatusPending:
		result.ErrorCode = ErrorCodeTemporaryFailure
		result.CanRetry = true
	case AttemptStatusSucceeded:
		result.Status = stringValue(attempt.FeedbackJSON, "status")
		result.Explanation = attempt.FeedbackText
		if corrected, ok := attempt.FeedbackJSON["corrected_sentence"].(string); ok {
			result.CorrectedSentence = &corrected
		}
		if tip, ok := attempt.FeedbackJSON["improvement_tip"].(string); ok {
			result.ImprovementTip = &tip
		}
		result.MissionCompleted = false
	case AttemptStatusFailed:
		result.ErrorCode = attempt.ErrorCode
		if result.ErrorCode == "" {
			result.ErrorCode = ErrorCodeTemporaryFailure
		}
		result.ErrorMessage = attempt.ErrorMessage
		result.CanRetry = true
	}

	return result
}

func (s *Service) recordTelemetry(ctx context.Context, userID uuid.UUID, target *Target, outcome string, durationMs int64, learningStatus string) {
	ev := FeedbackEvent{
		UserID:        userID,
		PromptVersion: PromptVersionSentenceFeedbackV1,
		SchemaVersion: SchemaVersionFeedbackV1,
		Provider:      s.config.Provider,
		Model:         s.config.Model,
		Outcome:       outcome,
		DurationMs:    durationMs,
	}
	if target != nil {
		ev.LearningStatus = learningStatus
	}
	// Telemetry recording is best-effort and must not leak learner text.
	s.telemetry.Record(ctx, ev)

	// Metrics are privacy-safe and never contain learner text or user identity.
	s.config.Metrics.RecordFeedback(ctx, MetricsEvent{
		PromptVersion:  PromptVersionSentenceFeedbackV1,
		SchemaVersion:  SchemaVersionFeedbackV1,
		Provider:       s.config.Provider,
		Model:          s.config.Model,
		Release:        s.config.Release,
		Outcome:        outcome,
		DurationMs:     durationMs,
		LearningStatus: learningStatus,
		Count:          1,
	})
}

func stringValue(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
