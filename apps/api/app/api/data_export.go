package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/accounts"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/auth"
	"github.com/danielgtaylor/huma/v2"
)

// CreatePersonalDataExportInput has no body: identity is always derived from
// the authenticated session, and the key makes an accidental double-submit
// safe. The response is a portable JSON document downloaded by Settings.
type CreatePersonalDataExportInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" doc:"User-scoped idempotency key (DOC-07)"`
}

type CreatePersonalDataExportOutput struct {
	CacheControl string `header:"Cache-Control"`
	Pragma       string `header:"Pragma"`
	Body         PersonalDataExportDTO
}

// PersonalDataExportDTO is an explicit learner-visible export contract. It is
// deliberately not a map or a database row: adding a database column cannot
// make it downloadable without a conscious DTO and SQL-projection change.
type PersonalDataExportDTO struct {
	SchemaVersion           string                            `json:"schemaVersion"`
	ExportedAt              string                            `json:"exportedAt,omitempty" format:"date-time"`
	Profile                 PersonalDataProfileDTO            `json:"profile"`
	Settings                PersonalDataSettingsDTO           `json:"settings"`
	OnboardingProfile       *PersonalDataOnboardingDTO        `json:"onboardingProfile"`
	SavedWords              []PersonalDataSavedWordDTO        `json:"savedWords"`
	ReviewHistory           []PersonalDataReviewDTO           `json:"reviewHistory"`
	SentenceFeedbackHistory []PersonalDataSentenceFeedbackDTO `json:"sentenceFeedbackHistory"`
	DailyMissions           []PersonalDataMissionDTO          `json:"dailyMissions"`
	DailyActivity           []PersonalDataActivityDTO         `json:"dailyActivity"`
	ConfidencePointLedger   []PersonalDataConfidenceLedgerDTO `json:"confidencePointLedger"`
	GraceDayLedger          []PersonalDataGraceLedgerDTO      `json:"graceDayLedger"`
	StreakState             *PersonalDataStreakDTO            `json:"streakState"`
}

type PersonalDataProfileDTO struct {
	ID               string  `json:"id"`
	Email            *string `json:"email"`
	DisplayName      *string `json:"displayName"`
	AvatarURL        *string `json:"avatarUrl"`
	OnboardingStatus string  `json:"onboardingStatus"`
	EmailVerifiedAt  *string `json:"emailVerifiedAt"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}
type PersonalDataSettingsDTO struct {
	Timezone               string  `json:"timezone"`
	DailyReviewTarget      int     `json:"dailyReviewTarget"`
	ReviewIntervalPreset   string  `json:"reviewIntervalPreset"`
	NotificationsEnabled   bool    `json:"notificationsEnabled"`
	MarketingEmailsEnabled bool    `json:"marketingEmailsEnabled"`
	AppLanguage            string  `json:"appLanguage"`
	CreatedAt              *string `json:"createdAt"`
	UpdatedAt              *string `json:"updatedAt"`
}
type PersonalDataOnboardingDTO struct {
	_                 struct{} `nullable:"true"`
	EnglishLevel      string   `json:"englishLevel"`
	NativeLanguage    string   `json:"nativeLanguage"`
	LearningGoal      string   `json:"learningGoal"`
	MainUseCase       string   `json:"mainUseCase"`
	DailyReviewTarget int      `json:"dailyReviewTarget"`
	CompletedAt       string   `json:"completedAt"`
	CreatedAt         string   `json:"createdAt"`
	UpdatedAt         string   `json:"updatedAt"`
}
type PersonalDataSavedWordDTO struct {
	ID                        string  `json:"id"`
	MeaningID                 string  `json:"meaningId"`
	WordID                    string  `json:"wordId"`
	WordText                  string  `json:"wordText"`
	PartOfSpeech              string  `json:"partOfSpeech"`
	ShortDefinition           string  `json:"shortDefinition"`
	Status                    string  `json:"status"`
	Source                    string  `json:"source"`
	ReviewStep                int     `json:"reviewStep"`
	NextReviewAt              *string `json:"nextReviewAt"`
	LastReviewedAt            *string `json:"lastReviewedAt"`
	LastResult                *string `json:"lastResult"`
	LastRating                *string `json:"lastRating"`
	ConsecutiveCorrectCount   int     `json:"consecutiveCorrectCount"`
	ConsecutiveIncorrectCount int     `json:"consecutiveIncorrectCount"`
	TotalReviewCount          int     `json:"totalReviewCount"`
	CorrectReviewCount        int     `json:"correctReviewCount"`
	AddedAt                   string  `json:"addedAt"`
	MasteredAt                *string `json:"masteredAt"`
	IgnoredAt                 *string `json:"ignoredAt"`
	CreatedAt                 string  `json:"createdAt"`
	UpdatedAt                 string  `json:"updatedAt"`
}
type PersonalDataReviewDTO struct {
	ID                      string  `json:"id"`
	UserWordID              string  `json:"userWordId"`
	MeaningID               string  `json:"meaningId"`
	AttemptType             string  `json:"attemptType"`
	PromptType              string  `json:"promptType"`
	Result                  string  `json:"result"`
	Rating                  *string `json:"rating"`
	ReviewStepBefore        int     `json:"reviewStepBefore"`
	ReviewStepAfter         int     `json:"reviewStepAfter"`
	AnsweredAt              string  `json:"answeredAt"`
	ResponseTimeMS          int     `json:"responseTimeMs"`
	SelectedOptionMeaningID *string `json:"selectedOptionMeaningId"`
	TypedAnswer             *string `json:"typedAnswer"`
	WasHintUsed             bool    `json:"wasHintUsed"`
	Source                  string  `json:"source"`
	CreatedAt               string  `json:"createdAt"`
	UpdatedAt               string  `json:"updatedAt"`
}
type PersonalDataFeedbackDTO struct {
	_                       struct{} `nullable:"true"`
	Status                  *string  `json:"status"`
	TargetWordUsedCorrectly *bool    `json:"targetWordUsedCorrectly"`
	CorrectedSentence       *string  `json:"correctedSentence"`
	Explanation             *string  `json:"explanation"`
	ImprovementTip          *string  `json:"improvementTip"`
}
type PersonalDataFeedbackAttemptDTO struct {
	Status       string                   `json:"status"`
	Feedback     *PersonalDataFeedbackDTO `json:"feedback"`
	FeedbackText *string                  `json:"feedbackText"`
	StartedAt    *string                  `json:"startedAt"`
	CompletedAt  *string                  `json:"completedAt"`
	CreatedAt    string                   `json:"createdAt"`
	UpdatedAt    string                   `json:"updatedAt"`
}
type PersonalDataSentenceDTO struct {
	ID           string  `json:"id"`
	MeaningID    *string `json:"meaningId"`
	UserWordID   *string `json:"userWordId"`
	SentenceText string  `json:"sentenceText"`
	Source       string  `json:"source"`
	Status       string  `json:"status"`
	SubmittedAt  string  `json:"submittedAt"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}
type PersonalDataSentenceFeedbackDTO struct {
	Sentence         PersonalDataSentenceDTO          `json:"sentence"`
	FeedbackAttempts []PersonalDataFeedbackAttemptDTO `json:"feedbackAttempts"`
}
type PersonalDataMissionDTO struct {
	LocalDate                  string  `json:"localDate"`
	Timezone                   string  `json:"timezone"`
	ReviewTarget               int     `json:"reviewTarget"`
	ReviewsCompleted           int     `json:"reviewsCompleted"`
	NewWordTarget              *int    `json:"newWordTarget"`
	NewWordsCompleted          *int    `json:"newWordsCompleted"`
	SentencePracticeTarget     *int    `json:"sentencePracticeTarget"`
	SentencePracticesCompleted *int    `json:"sentencePracticesCompleted"`
	PolicyVersion              string  `json:"policyVersion"`
	Status                     string  `json:"status"`
	CompletedAt                *string `json:"completedAt"`
	GraceApplied               bool    `json:"graceApplied"`
	CreatedAt                  string  `json:"createdAt"`
	UpdatedAt                  string  `json:"updatedAt"`
}
type PersonalDataActivityDTO struct {
	LocalDate              string `json:"localDate"`
	Timezone               string `json:"timezone"`
	ReviewsAttempted       int    `json:"reviewsAttempted"`
	ReviewsCorrect         int    `json:"reviewsCorrect"`
	ReviewsSkipped         int    `json:"reviewsSkipped"`
	WordsDiscovered        int    `json:"wordsDiscovered"`
	WordsAdded             int    `json:"wordsAdded"`
	SentencesSubmitted     int    `json:"sentencesSubmitted"`
	AIFeedbackReceived     int    `json:"aiFeedbackReceived"`
	ConfidencePointsEarned int    `json:"confidencePointsEarned"`
	ConfidencePointsSpent  int    `json:"confidencePointsSpent"`
	CreatedAt              string `json:"createdAt"`
	UpdatedAt              string `json:"updatedAt"`
}
type PersonalDataConfidenceLedgerDTO struct {
	Amount       int     `json:"amount"`
	BalanceAfter int     `json:"balanceAfter"`
	Reason       string  `json:"reason"`
	SourceType   string  `json:"sourceType"`
	SourceID     *string `json:"sourceId"`
	OccurredAt   string  `json:"occurredAt"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}
type PersonalDataGraceLedgerDTO struct {
	Amount             int     `json:"amount"`
	BalanceAfter       int     `json:"balanceAfter"`
	Reason             string  `json:"reason"`
	SourceType         string  `json:"sourceType"`
	SourceID           *string `json:"sourceId"`
	AppliedToLocalDate string  `json:"appliedToLocalDate"`
	Timezone           string  `json:"timezone"`
	CreatedAt          string  `json:"createdAt"`
	UpdatedAt          string  `json:"updatedAt"`
}
type PersonalDataStreakDTO struct {
	_                      struct{} `nullable:"true"`
	CurrentStreakCount     int      `json:"currentStreakCount"`
	LongestStreakCount     int      `json:"longestStreakCount"`
	LastCompletedLocalDate *string  `json:"lastCompletedLocalDate"`
	LastActivityLocalDate  *string  `json:"lastActivityLocalDate"`
	Timezone               string   `json:"timezone"`
	Status                 string   `json:"status"`
	CreatedAt              string   `json:"createdAt"`
	UpdatedAt              string   `json:"updatedAt"`
}

func RegisterPersonalDataExports(api huma.API, svc *accounts.Service, authSvc *auth.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "CreatePersonalDataExport",
		Method:      http.MethodPost, Path: "/api/v1/personal-data-export",
		Summary:     "Download the requester's personal data as JSON",
		Description: "Synchronous MVP export. Includes learner-visible profile, settings, saved-word, review, practice, feedback, and progress history; excludes authentication secrets, hidden prompts, provider credentials, and internal abuse/report classifications.",
		Tags:        []string{"Account"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth(), CSRFMiddleware(authSvc)},
		Responses: map[string]*huma.Response{
			"401": {Description: "Authentication is required"}, "403": {Description: "Invalid CSRF token"},
			"404": {Description: "User not found"}, "409": {Description: "Idempotency-Key conflict"},
			"422": {Description: "Missing or invalid Idempotency-Key"}, "429": {Description: "Rate limited"},
		},
	}, func(ctx context.Context, input *CreatePersonalDataExportInput) (*CreatePersonalDataExportOutput, error) {
		c := authHumaContext(ctx)
		payload, err := svc.ExportPersonalData(ctx, RequesterUserID(ctx).String(), clientIPFromHuma(c), sessionTokenFromHuma(c, authSvc), input.IdempotencyKey)
		if err != nil {
			return nil, mapDataExportError(err)
		}
		var body PersonalDataExportDTO
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil, huma.Error500InternalServerError("internal error")
		}
		return &CreatePersonalDataExportOutput{CacheControl: "no-store", Pragma: "no-cache", Body: body}, nil
	})
}

func mapDataExportError(err error) huma.StatusError {
	switch {
	case errors.Is(err, accounts.ErrDataExportIdempotencyKeyRequired):
		return huma.Error400BadRequest("idempotency key required")
	case errors.Is(err, accounts.ErrDataExportIdempotencyConflict):
		return huma.Error409Conflict("idempotency key conflict")
	case errors.Is(err, accounts.ErrDataExportRateLimited):
		return huma.Error429TooManyRequests("personal data export rate limited")
	case errors.Is(err, accounts.ErrUserNotFound):
		return huma.Error404NotFound("user not found")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}
