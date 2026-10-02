package aifeedback

import (
	"context"
	"errors"
)

type OpenAIStage uint8

const (
	OpenAIStageFeedback OpenAIStage = iota + 1
	OpenAIStageModeration
)

func (s OpenAIStage) String() string {
	switch s {
	case OpenAIStageFeedback:
		return "feedback"
	case OpenAIStageModeration:
		return "moderation"
	default:
		return "unknown"
	}
}

type OpenAIFailureCategory uint8

const (
	OpenAIFailureUnknown OpenAIFailureCategory = iota
	OpenAIFailureAuthentication
	OpenAIFailureRequestRejected
	OpenAIFailureRateLimited
	OpenAIFailureUpstream
	OpenAIFailureDeadline
	OpenAIFailureCancelled
	OpenAIFailureUnavailable
	OpenAIFailureInvalidResponse
	OpenAIFailureRefusal
)

func (c OpenAIFailureCategory) String() string {
	switch c {
	case OpenAIFailureAuthentication:
		return "authentication"
	case OpenAIFailureRequestRejected:
		return "request_rejected"
	case OpenAIFailureRateLimited:
		return "rate_limited"
	case OpenAIFailureUpstream:
		return "upstream_failure"
	case OpenAIFailureDeadline:
		return "deadline"
	case OpenAIFailureCancelled:
		return "cancelled"
	case OpenAIFailureUnavailable:
		return "unavailable"
	case OpenAIFailureInvalidResponse:
		return "invalid_response"
	case OpenAIFailureRefusal:
		return "refusal"
	default:
		return "unknown"
	}
}

// OpenAIFailure deliberately contains no provider strings, identities or input.
// HTTPStatus is zero when unavailable. This is not token or billing telemetry.
type OpenAIFailure struct {
	Stage      OpenAIStage
	Category   OpenAIFailureCategory
	HTTPStatus int
}

func (p *OpenAIFeedbackProvider) recordFailure(stage OpenAIStage, status int, err error) {
	if err == nil || p.config.OnFailure == nil {
		return
	}
	event := OpenAIFailure{Stage: stage}
	if status >= 100 && status <= 599 {
		event.HTTPStatus = status
	}
	switch {
	case errors.Is(err, context.Canceled):
		event.Category = OpenAIFailureCancelled
	case errors.Is(err, context.DeadlineExceeded):
		event.Category = OpenAIFailureDeadline
	case errors.Is(err, ErrProviderAuth):
		event.Category = OpenAIFailureAuthentication
	case status == 429:
		event.Category = OpenAIFailureRateLimited
	case status >= 500 && status <= 599:
		event.Category = OpenAIFailureUpstream
	case errors.Is(err, ErrProviderInvalidInput):
		event.Category = OpenAIFailureRequestRejected
	case errors.Is(err, ErrProviderRefusal):
		event.Category = OpenAIFailureRefusal
	case errors.Is(err, ErrProviderInvalidResponse):
		event.Category = OpenAIFailureInvalidResponse
	case errors.Is(err, ErrProviderTimeout):
		event.Category = OpenAIFailureUnavailable
	}
	p.config.OnFailure(event)
}
