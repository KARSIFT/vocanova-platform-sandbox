package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPSender is a Sender implementation that POSTs each message as
// JSON to a transactional-email provider's HTTP API. It is
// provider-agnostic: any provider that accepts an
// `Authorization: Bearer <key>` JSON POST with a
// {"from","to","subject","text","html"} body works (Resend,
// SendGrid v3, Postmark's token-auth mode, etc.). The
// provider-agnostic shape is intentional - T14 fixes the
// "no real email sender exists" gap (VOC-032-D10) by adding the
// first HTTP-based real implementation, not by binding to a single
// vendor's exact wire format. A future, narrower follow-up can
// add a provider-specific request shape if the founder picks a
// vendor whose API needs something this generic sender does not
// produce.
//
// HTTPSender is the production-wiring path. Fake{} remains in place
// for unit tests and for the "no credential configured" fallback -
// the production wiring in apps/api/app/api/production.go uses
// HTTPSender only when the credential env var is set, and falls
// back to Fake{} otherwise (and always when
// EMAIL_MAGIC_LINK_ENABLED is false, per DOC-11 §3's kill switch).
//
// HTTPSender is safe for concurrent use; the only mutable state is
// the *http.Client, which net/http documents as safe for concurrent
// use across requests.
type HTTPSender struct {
	URL    string
	APIKey string
	From   string
	Client *http.Client
}

// HTTPSenderConfig captures the configuration NewHTTPSender needs.
// Callers that need a custom transport (e.g. a fake transport for
// tests) should construct the HTTPSender directly rather than
// calling this constructor, so the test can substitute Client.
type HTTPSenderConfig struct {
	URL     string
	APIKey  string
	From    string
	Timeout time.Duration
}

// NewHTTPSender returns an HTTPSender with a default *http.Client
// whose timeout is set from cfg. Required fields (URL, APIKey, From)
// are validated up front; an empty cfg returns a descriptive error
// rather than letting the misconfiguration surface at first send.
func NewHTTPSender(cfg HTTPSenderConfig) (*HTTPSender, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, errors.New("email: HTTPSenderConfig.URL is required")
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("email: HTTPSenderConfig.APIKey is required")
	}
	if strings.TrimSpace(cfg.From) == "" {
		return nil, errors.New("email: HTTPSenderConfig.From is required")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &HTTPSender{
		URL:    cfg.URL,
		APIKey: cfg.APIKey,
		From:   cfg.From,
		Client: &http.Client{Timeout: timeout},
	}, nil
}

// httpPayload is the provider-agnostic JSON body HTTPSender POSTs.
// Field order is stable for any test or operator that diffs the
// rendered wire format; the JSON encoder sorts by struct order.
type httpPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	HTML    string   `json:"html"`
}

// Send POSTs msg to the configured URL as JSON, authenticated with
// the configured API key as a Bearer token. The From address comes
// from the sender's configuration (per-deployment, not per-message).
// Send does not retry; the caller (e.g. the auth service) decides
// retry policy. Send never logs the API key and never includes it
// in any returned error message.
func (s *HTTPSender) Send(ctx context.Context, msg Message) error {
	if len(msg.To) == 0 {
		return errors.New("email: Send called with no recipients")
	}
	if strings.TrimSpace(msg.Subject) == "" {
		return errors.New("email: Send called with empty subject")
	}
	if strings.TrimSpace(msg.BodyText) == "" && strings.TrimSpace(msg.BodyHTML) == "" {
		return errors.New("email: Send called with no body")
	}

	recipients := make([]string, 0, len(msg.To))
	for _, addr := range msg.To {
		if strings.TrimSpace(addr.Email) == "" {
			return errors.New("email: Send called with recipient missing email")
		}
		recipients = append(recipients, addr.Email)
	}

	body, err := json.Marshal(httpPayload{
		From:    s.From,
		To:      recipients,
		Subject: msg.Subject,
		Text:    msg.BodyText,
		HTML:    msg.BodyHTML,
	})
	if err != nil {
		return fmt.Errorf("email: marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email: build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "vocanova-api/1.0")

	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("email: send request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("email: provider returned status %d", resp.StatusCode)
	}
	return nil
}
