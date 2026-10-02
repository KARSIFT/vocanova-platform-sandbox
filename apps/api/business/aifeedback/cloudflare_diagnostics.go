package aifeedback

import (
	"context"
	"encoding/json"
	"errors"
	"net"
)

// CloudflareFailureCategory is a closed diagnostic vocabulary. Provider text
// must never be used as a category or included in the failure event.
type CloudflareFailureCategory uint8

const (
	CloudflareFailureUnknown CloudflareFailureCategory = iota
	CloudflareFailureAuthentication
	CloudflareFailureRequestRejected
	CloudflareFailureRateLimited
	CloudflareFailureUpstream
	CloudflareFailureTimeout
	CloudflareFailureCancelled
	CloudflareFailureTransport
	CloudflareFailureInvalidResponse
	CloudflareFailureRefusal
)

func (c CloudflareFailureCategory) String() string {
	switch c {
	case CloudflareFailureAuthentication:
		return "authentication"
	case CloudflareFailureRequestRejected:
		return "request_rejected"
	case CloudflareFailureRateLimited:
		return "rate_limited"
	case CloudflareFailureUpstream:
		return "upstream_failure"
	case CloudflareFailureTimeout:
		return "timeout"
	case CloudflareFailureCancelled:
		return "cancelled"
	case CloudflareFailureTransport:
		return "transport_failure"
	case CloudflareFailureInvalidResponse:
		return "invalid_response"
	case CloudflareFailureRefusal:
		return "refusal"
	default:
		return "unknown"
	}
}

// CloudflareModerationFailure contains only enums and numeric metadata.
// Zero HTTPStatus/ProviderCode means unavailable; no identity, URL, input,
// response text or credential is retained.
type CloudflareModerationFailure struct {
	Category     CloudflareFailureCategory
	HTTPStatus   int
	ProviderCode int
}

// Preserve existing error identities/retry behavior while retaining the
// network classification that mapNetworkError otherwise collapses to timeout.
type cloudflareNetworkError struct {
	cause      error
	category   CloudflareFailureCategory
	httpStatus int
}

func (e *cloudflareNetworkError) Error() string { return e.cause.Error() }
func (e *cloudflareNetworkError) Unwrap() error { return e.cause }

func cloudflareNetworkFailure(err error) *cloudflareNetworkError {
	category := CloudflareFailureTransport
	var networkError net.Error
	switch {
	case errors.Is(err, context.Canceled):
		category = CloudflareFailureCancelled
	case errors.Is(err, context.DeadlineExceeded):
		category = CloudflareFailureTimeout
	case errors.As(err, &networkError) && networkError.Timeout():
		category = CloudflareFailureTimeout
	}
	return &cloudflareNetworkError{cause: mapNetworkError(err), category: category}
}

func cloudflareModerationFailure(err error, status int, body []byte) CloudflareModerationFailure {
	event := CloudflareModerationFailure{Category: CloudflareFailureInvalidResponse}
	if status >= 100 && status <= 599 {
		event.HTTPStatus = status
	}
	var envelope struct {
		Errors []struct {
			Code int `json:"code"`
		} `json:"errors"`
	}
	if json.Unmarshal(body, &envelope) == nil && len(envelope.Errors) > 0 {
		code := envelope.Errors[0].Code
		if code > 0 && code <= 2147483647 {
			event.ProviderCode = code
		}
	}
	var networkError *cloudflareNetworkError
	switch {
	case errors.As(err, &networkError):
		event.Category = networkError.category
		if networkError.httpStatus >= 100 && networkError.httpStatus <= 599 {
			event.HTTPStatus = networkError.httpStatus
		}
	case status == 401 || status == 403:
		event.Category = CloudflareFailureAuthentication
	case status == 429:
		event.Category = CloudflareFailureRateLimited
	case status >= 500 && status <= 599:
		event.Category = CloudflareFailureUpstream
	case status >= 400 && status <= 499:
		event.Category = CloudflareFailureRequestRejected
	case errors.Is(err, ErrProviderAuth):
		event.Category = CloudflareFailureAuthentication
	case errors.Is(err, ErrProviderInvalidInput):
		event.Category = CloudflareFailureRequestRejected
	case errors.Is(err, ErrProviderRefusal):
		event.Category = CloudflareFailureRefusal
	case errors.Is(err, context.Canceled):
		event.Category = CloudflareFailureCancelled
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, ErrProviderTimeout):
		event.Category = CloudflareFailureTimeout
	}
	return event
}
