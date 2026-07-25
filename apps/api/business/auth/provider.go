package auth

import (
	"context"
	"fmt"
	"sync"
)

// FakeOAuthProvider is a deterministic provider for service and HTTP tests.
// It never calls a real identity provider and never stores real credentials.
type FakeOAuthProvider struct {
	mu       sync.Mutex
	identity *OAuthIdentity
}

// NewFakeOAuthProvider returns a fake provider that returns identity for every
// valid code/state/redirectURI.
func NewFakeOAuthProvider(identity *OAuthIdentity) *FakeOAuthProvider {
	return &FakeOAuthProvider{identity: identity}
}

// SetIdentity updates the identity returned by Verify.
func (p *FakeOAuthProvider) SetIdentity(identity *OAuthIdentity) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.identity = identity
}

// AuthURL returns a placeholder URL that includes the state and redirect URI.
func (p *FakeOAuthProvider) AuthURL(state, redirectURI string) string {
	return fmt.Sprintf("https://fake-oauth.example.com/auth?state=%s&redirect_uri=%s", state, redirectURI)
}

// Verify returns the configured identity if code and state are non-empty.
func (p *FakeOAuthProvider) Verify(ctx context.Context, code, state, redirectURI string) (*OAuthIdentity, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if code == "" || state == "" {
		return nil, fmt.Errorf("missing code or state")
	}
	if p.identity == nil {
		return nil, fmt.Errorf("no identity configured")
	}
	copy := *p.identity
	return &copy, nil
}
