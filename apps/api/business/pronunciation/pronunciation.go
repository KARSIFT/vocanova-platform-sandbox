// Package pronunciation provides bounded, cached audio for public canonical
// teaching text. Learner writing never crosses this provider boundary.
package pronunciation

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/clock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
)

const Model = "@cf/deepgram/aura-2-en"
const Voice = "luna"
const DailyCharacterLimit = 1000
const DailyGenerationLimit = 20
const MaxAudioBytes = 1024 * 1024

var (
	ErrInvalid     = errors.New("invalid pronunciation text")
	ErrNotFound    = errors.New("canonical pronunciation text not found")
	ErrUnavailable = errors.New("pronunciation unavailable")
	ErrRateLimited = errors.New("pronunciation limit reached")
	ErrBusy        = errors.New("pronunciation generation pending")
)

type Audio struct {
	AudioBase64     string `json:"audioBase64"`
	MimeType        string `json:"mimeType"`
	Voice           string `json:"voice"`
	CacheHit        bool   `json:"cacheHit"`
	MeaningID       string `json:"meaningId,omitempty"`
	MeaningRevision string `json:"meaningRevision,omitempty"`
	PartOfSpeech    string `json:"partOfSpeech,omitempty"`
}

// Utterance is resolved from active public catalog content, never client guidance.
type Utterance struct {
	MeaningID    string
	WordText     string
	Text         string
	Kind         string
	PartOfSpeech string
	Definition   string
	Context      string
	Locale       string
}
type ProviderIdentity struct{ Provider, Model, Voice, InstructionVersion string }
type Provider interface {
	Generate(context.Context, Utterance) ([]byte, error)
	Identity() ProviderIdentity
}
type reservation struct {
	audio      []byte
	generation uuid.UUID
}
type repository interface {
	ResolveUtterance(context.Context, uuid.UUID, string, string) (Utterance, error)
	Reserve(context.Context, string, int, time.Time) (reservation, error)
	Complete(context.Context, string, uuid.UUID, []byte, time.Time) error
}
type Service struct {
	repo     repository
	provider Provider
	enabled  bool
	clock    clock.Clock
	limiter  aifeedback.RateLimiter
}

func NewService(db *sql.DB, p Provider, enabled bool, c clock.Clock) *Service {
	return newService(&postgresRepository{db}, p, enabled, c)
}
func newService(repo repository, p Provider, enabled bool, c clock.Clock) *Service {
	if c == nil {
		c = clock.Real{}
	}
	return &Service{repo: repo, provider: p, enabled: enabled && p != nil, clock: c, limiter: aifeedback.NewMemoryRateLimiter(aifeedback.RateLimitConfig{MaxActivePerLearner: 1, MaxPerMinute: 60, MaxPerDay: 1000}, c)}
}
func (s *Service) Enabled() bool { return s != nil && s.enabled }
func cacheKey(text string) string {
	return utteranceCacheKey(ProviderIdentity{Provider: "cloudflare", Model: Model, Voice: Voice, InstructionVersion: "1"}, Utterance{Text: text})
}
func utteranceCacheKey(identity ProviderIdentity, utterance Utterance) string {
	data, _ := json.Marshal(struct {
		Identity  ProviderIdentity
		Utterance Utterance
	}{identity, utterance})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (s *Service) Speak(ctx context.Context, user uuid.UUID, text string) (*Audio, error) {
	return s.SpeakMeaning(ctx, user, text, "")
}
func (s *Service) SpeakMeaning(ctx context.Context, user uuid.UUID, text, meaningID string) (*Audio, error) {
	return s.SpeakMeaningRevision(ctx, user, text, meaningID, "")
}
func (s *Service) SpeakMeaningRevision(ctx context.Context, user uuid.UUID, text, meaningID, revision string) (*Audio, error) {
	if user == uuid.Nil {
		return nil, ErrNotFound
	}
	if !s.Enabled() {
		return nil, ErrUnavailable
	}
	if text == "" || text != strings.TrimSpace(text) || !utf8.ValidString(text) || strings.ContainsRune(text, 0) || utf8.RuneCountInString(text) > 500 {
		return nil, ErrInvalid
	}
	if revision != "" {
		decoded, err := hex.DecodeString(revision)
		if meaningID == "" || err != nil || len(decoded) != sha256.Size || revision != strings.ToLower(revision) {
			return nil, ErrInvalid
		}
	}
	if meaningID != "" {
		id, err := uuid.Parse(meaningID)
		if err != nil || id == uuid.Nil {
			return nil, ErrInvalid
		}
		meaningID = id.String()
	}
	if err := s.limiter.Allow(ctx, user); err != nil {
		return nil, ErrRateLimited
	}
	defer s.limiter.Release(user)
	utterance, err := s.repo.ResolveUtterance(ctx, user, text, meaningID)
	if err != nil {
		return nil, err
	}
	liveRevision := meaningidentity.Revision(utterance.MeaningID, utterance.WordText, utterance.PartOfSpeech, utterance.Definition)
	if revision != "" && (liveRevision == "" || revision != liveRevision) {
		return nil, ErrNotFound
	}
	identity := s.provider.Identity()
	// Cloudflare's legacy text-only adapter cannot honor selected grammatical
	// meaning. Never charge or present it as sense-specific pronunciation.
	if identity.Provider == "cloudflare" && meaningID != "" {
		return nil, ErrUnavailable
	}
	key := utteranceCacheKey(identity, utterance)
	reserved, err := s.repo.Reserve(ctx, key, utf8.RuneCountInString(text), s.clock.Now().UTC())
	if err != nil {
		return nil, err
	}
	if len(reserved.audio) > 0 {
		return audioResponse(reserved.audio, true, identity, utterance), nil
	}
	// Once a paid request is reserved, finish its bounded work even if a mobile
	// learner changes cards. An aborted HTTP response must not waste the cache.
	providerCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 12*time.Second)
	data, providerErr := s.provider.Generate(providerCtx, utterance)
	cancel()
	if providerErr != nil || !validMP3(data) {
		data = nil
	}
	finishCtx, finishCancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	err = s.repo.Complete(finishCtx, key, reserved.generation, data, s.clock.Now().UTC())
	finishCancel()
	if err != nil {
		return nil, ErrUnavailable
	}
	if len(data) == 0 {
		return nil, ErrUnavailable
	}
	return audioResponse(data, false, identity, utterance), nil
}
func audioResponse(data []byte, hit bool, identity ProviderIdentity, utterance Utterance) *Audio {
	return &Audio{AudioBase64: base64.StdEncoding.EncodeToString(data), MimeType: "audio/mpeg", Voice: identity.Voice, CacheHit: hit, MeaningID: utterance.MeaningID, MeaningRevision: meaningidentity.Revision(utterance.MeaningID, utterance.WordText, utterance.PartOfSpeech, utterance.Definition), PartOfSpeech: utterance.PartOfSpeech}
}
func validMP3(data []byte) bool {
	return len(data) >= 3 && len(data) <= MaxAudioBytes && (string(data[:3]) == "ID3" || (data[0] == 0xff && data[1]&0xe0 == 0xe0))
}
