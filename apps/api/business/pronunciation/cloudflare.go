package pronunciation

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CloudflareConfig struct {
	APIToken, AccountID, BaseURL string
	Timeout                      time.Duration
}
type cloudflareProvider struct {
	config CloudflareConfig
	client *http.Client
}

func NewCloudflareProvider(config CloudflareConfig) Provider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.cloudflare.com/client/v4"
	}
	if config.Timeout <= 0 || config.Timeout > 12*time.Second {
		config.Timeout = 12 * time.Second
	}
	return &cloudflareProvider{config: config, client: &http.Client{Timeout: config.Timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (p *cloudflareProvider) Identity() ProviderIdentity {
	return ProviderIdentity{Provider: "cloudflare", Model: Model, Voice: Voice, InstructionVersion: "1"}
}
func (p *cloudflareProvider) Generate(ctx context.Context, utterance Utterance) ([]byte, error) {
	if strings.TrimSpace(p.config.APIToken) == "" || strings.TrimSpace(p.config.AccountID) == "" {
		return nil, ErrUnavailable
	}
	body, err := json.Marshal(map[string]string{"text": utterance.Text, "speaker": Voice, "encoding": "mp3"})
	if err != nil {
		return nil, ErrUnavailable
	}
	endpoint := strings.TrimRight(p.config.BaseURL, "/") + "/accounts/" + url.PathEscape(p.config.AccountID) + "/ai/run/" + Model
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, ErrUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+p.config.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnavailable
	}
	media, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil || (media != "audio/mpeg" && media != "audio/mp3" && media != "application/octet-stream") {
		return nil, ErrUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, MaxAudioBytes+1))
	if err != nil || !validMP3(data) {
		return nil, ErrUnavailable
	}
	return data, nil
}
