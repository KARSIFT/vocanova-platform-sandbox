package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/danielgtaylor/huma/v2"
)

// StreakDTO is the shared streak projection that backs both
// GET /api/v1/daily-mission and GET /api/v1/progress so Home and Progress
// can never disagree (DOC-12 §5 P4 gate "cross-capability consistency").
type StreakDTO struct {
	CurrentStreakCount int    `json:"currentStreakCount" doc:"Current consecutive local-day completion count"`
	LongestStreakCount int    `json:"longestStreakCount" doc:"All-time longest streak count"`
	Status             string `json:"status" enum:"active,at_risk,broken" doc:"Streak state"`
	GraceDayBalance    int    `json:"graceDayBalance" doc:"Available grace-day balance"`
}

// DailyMissionDTO is the public projection of today's daily mission
// (specification.md item 5).
type DailyMissionDTO struct {
	LocalDate                  string     `json:"localDate" format:"date" doc:"Local calendar date for the mission"`
	Timezone                   string     `json:"timezone" doc:"Resolved IANA timezone for the mission"`
	ReviewTarget               int        `json:"reviewTarget" doc:"Required review count for the mission"`
	ReviewsCompleted           int        `json:"reviewsCompleted" doc:"Reviews completed against today's target"`
	NewWordTarget              *int       `json:"newWordTarget,omitempty" doc:"Optional new-word goal (D03: disabled at launch)"`
	NewWordsCompleted          *int       `json:"newWordsCompleted,omitempty" doc:"Optional new-word counter (D03: disabled at launch)"`
	SentencePracticeTarget     *int       `json:"sentencePracticeTarget,omitempty" doc:"Optional sentence-practice goal (D03: disabled at launch)"`
	SentencePracticesCompleted *int       `json:"sentencePracticesCompleted,omitempty" doc:"Optional sentence-practice counter (D03: disabled at launch)"`
	PolicyVersion              string     `json:"policyVersion" doc:"Snapshot policy version (e.g. p4-mission-policy-v1)"`
	Status                     string     `json:"status" enum:"open,completed,missed,protected" doc:"Mission state"`
	CompletedAt                *time.Time `json:"completedAt,omitempty" format:"date-time" doc:"When the mission was first completed (status=completed)"`
	GraceApplied               bool       `json:"graceApplied" doc:"Whether a grace day protected today's streak"`
	Streak                     StreakDTO  `json:"streak" doc:"Shared streak object (matches GET /api/v1/progress)"`
}

// ProgressDTO is the public projection of the requester's overall progress
// (specification.md item 6).
type ProgressDTO struct {
	ConfidencePointsBalance int                `json:"confidencePointsBalance" doc:"Current Confidence Points balance from the ledger"`
	Streak                  StreakDTO          `json:"streak" doc:"Shared streak object (matches GET /api/v1/daily-mission)"`
	CompletionHistory       []CompletionDayDTO `json:"completionHistory" doc:"Bounded 7-day completion history"`
}

// CompletionDayDTO is one day in the bounded 7-day completion history.
type CompletionDayDTO struct {
	LocalDate string `json:"localDate" format:"date" doc:"Local calendar date for the day"`
	Completed bool   `json:"completed" doc:"Whether the mission was completed or protected that local day"`
}

// GetDailyMissionInput requests today's daily mission for the authenticated
// requester. The optional `timezone` query parameter supplies a client-side
// IANA timezone that the server validates and uses only when the requester
// has no stored user_settings row (VOC-030-D01).
type GetDailyMissionInput struct {
	Timezone string `query:"timezone" doc:"Optional client-supplied IANA timezone"`
}

// GetDailyMissionOutput returns the daily-mission projection.
type GetDailyMissionOutput struct {
	Body DailyMissionDTO
}

// GetProgressInput requests the requester's overall progress. The optional
// `timezone` query parameter follows the same D01 chain as
// GetDailyMission.
type GetProgressInput struct {
	Timezone string `query:"timezone" doc:"Optional client-supplied IANA timezone"`
}

// GetProgressOutput returns the progress projection.
type GetProgressOutput struct {
	Body ProgressDTO
}

// RegisterMissions registers the requester-scoped daily-mission and progress
// read routes. Both endpoints implicitly self-scope (no ID parameter exists
// that could be used to enumerate another learner) so the only
// authentication gate is RequireAuth. They are reads and require no CSRF
// token.
func RegisterMissions(api huma.API, svc *missions.Service) {
	huma.Register(api, huma.Operation{
		OperationID: "GetDailyMission",
		Method:      http.MethodGet,
		Path:        "/api/v1/daily-mission",
		Summary:     "Get today's daily mission for the authenticated requester",
		Tags:        []string{"Missions"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"200": {Description: "Today's daily mission projection"},
			"400": {Description: "Invalid client-supplied timezone"},
			"401": {Description: "Authentication is required"},
		},
	}, func(ctx context.Context, input *GetDailyMissionInput) (*GetDailyMissionOutput, error) {
		view, err := svc.GetDailyMissionView(ctx, RequesterUserID(ctx), input.Timezone, time.Now())
		if err != nil {
			return nil, mapMissionsError(err)
		}
		return &GetDailyMissionOutput{Body: dailyMissionViewToDTO(view)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "GetProgress",
		Method:      http.MethodGet,
		Path:        "/api/v1/progress",
		Summary:     "Get the authenticated requester's overall progress",
		Tags:        []string{"Missions"},
		Middlewares: []func(huma.Context, func(huma.Context)){RequireAuth()},
		Responses: map[string]*huma.Response{
			"200": {Description: "Progress projection including shared streak and 7-day completion history"},
			"400": {Description: "Invalid client-supplied timezone"},
			"401": {Description: "Authentication is required"},
		},
	}, func(ctx context.Context, input *GetProgressInput) (*GetProgressOutput, error) {
		view, err := svc.GetProgressView(ctx, RequesterUserID(ctx), input.Timezone, time.Now(), 7)
		if err != nil {
			return nil, mapMissionsError(err)
		}
		return &GetProgressOutput{Body: progressViewToDTO(view)}, nil
	})
}

func dailyMissionViewToDTO(v *missions.DailyMissionView) DailyMissionDTO {
	dto := DailyMissionDTO{
		LocalDate:        v.LocalDate.Format("2006-01-02"),
		Timezone:         v.Timezone,
		ReviewTarget:     v.ReviewTarget,
		ReviewsCompleted: v.ReviewsCompleted,
		PolicyVersion:    v.PolicyVersion,
		Status:           v.Status,
		CompletedAt:      v.CompletedAt,
		GraceApplied:     v.GraceApplied,
		Streak:           streakViewToDTO(v.Streak),
	}
	if v.NewWordTarget != nil {
		t := *v.NewWordTarget
		dto.NewWordTarget = &t
	}
	if v.NewWordsCompleted != nil {
		c := *v.NewWordsCompleted
		dto.NewWordsCompleted = &c
	}
	if v.SentencePracticeTarget != nil {
		t := *v.SentencePracticeTarget
		dto.SentencePracticeTarget = &t
	}
	if v.SentencePracticesCompleted != nil {
		c := *v.SentencePracticesCompleted
		dto.SentencePracticesCompleted = &c
	}
	return dto
}

func progressViewToDTO(v *missions.ProgressView) ProgressDTO {
	dto := ProgressDTO{
		ConfidencePointsBalance: v.ConfidencePointsBalance,
		Streak:                  streakViewToDTO(v.Streak),
		CompletionHistory:       make([]CompletionDayDTO, 0, len(v.CompletionHistory)),
	}
	for _, d := range v.CompletionHistory {
		dto.CompletionHistory = append(dto.CompletionHistory, CompletionDayDTO{
			LocalDate: d.LocalDate.Format("2006-01-02"),
			Completed: d.Completed,
		})
	}
	return dto
}

func streakViewToDTO(s missions.StreakView) StreakDTO {
	return StreakDTO{
		CurrentStreakCount: s.CurrentStreakCount,
		LongestStreakCount: s.LongestStreakCount,
		Status:             s.Status,
		GraceDayBalance:    s.GraceDayBalance,
	}
}

func mapMissionsError(err error) huma.StatusError {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, gamification.ErrInvalidTimezone):
		return huma.Error400BadRequest("invalid IANA timezone")
	default:
		return huma.Error500InternalServerError("internal error")
	}
}
