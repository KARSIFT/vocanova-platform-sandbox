package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/google/uuid"
)

// DataExportRepository returns the learner-visible, personal-data projection.
// It deliberately excludes sessions, credentials, raw provider prompts, abuse
// signals, and internal report classifications.
type DataExportRepository interface {
	ExportPersonalData(ctx context.Context, userID uuid.UUID) (json.RawMessage, error)
}

const dataExportOperation = "personal_data_export"

func dataExportFingerprint(userID uuid.UUID) string {
	return fmt.Sprintf("personal_data_export|%s", userID)
}

var (
	ErrDataExportIdempotencyKeyRequired = errors.New("idempotency key required")
	ErrDataExportIdempotencyConflict    = errors.New("idempotency key conflict")
	ErrDataExportRateLimited            = errors.New("personal data export rate limited")
)

// ExportPersonalData produces an immediate JSON export for the authenticated
// requester. This MVP intentionally has no queued job, retained download, or
// email delivery: the data only travels over the current authenticated HTTPS
// response. Replaying a key is safe and re-reads the current export rather
// than retaining a second copy of personal data server-side.
func (s *Service) ExportPersonalData(ctx context.Context, userID, clientIP, sessionToken, idempotencyKey string) (json.RawMessage, error) {
	uid, err := uuid.Parse(userID)
	if err != nil || uid == uuid.Nil {
		return nil, errors.New("user id required")
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, ErrDataExportIdempotencyKeyRequired
	}
	if s.idem == nil {
		return nil, errors.New("idempotency store not configured")
	}
	repo, ok := s.repo.(DataExportRepository)
	if !ok {
		return nil, errors.New("personal data export not configured")
	}

	fingerprint := dataExportFingerprint(uid)
	status, err := s.idem.Check(ctx, uid, dataExportOperation, idempotencyKey, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("idempotency check: %w", err)
	}
	if status == IdempotencyConflict {
		return nil, ErrDataExportIdempotencyConflict
	}
	if allowed, err := s.limiter.Allow(ctx, auth.KeyForIP("dataexport.request", clientIP)); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	} else if !allowed {
		return nil, ErrDataExportRateLimited
	}
	if allowed, err := s.limiter.Allow(ctx, auth.KeyForSession("dataexport.request", sessionToken)); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	} else if !allowed {
		return nil, ErrDataExportRateLimited
	}
	if _, err := s.auth.GetUserByID(ctx, uid); err != nil {
		return nil, ErrUserNotFound
	}

	payload, err := repo.ExportPersonalData(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("export personal data: %w", err)
	}
	if !json.Valid(payload) {
		return nil, errors.New("invalid personal data export payload")
	}
	if status == IdempotencyAbsent {
		if err := s.idem.Record(ctx, uid, dataExportOperation, idempotencyKey, fingerprint); err != nil {
			return nil, fmt.Errorf("record idempotency: %w", err)
		}
	}
	return payload, nil
}
