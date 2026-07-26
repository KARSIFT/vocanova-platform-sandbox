package missions

import "time"

// MissionStatus and Activity field types live here so the rest of the package
// can build strongly-typed writes. The constants are kept in sync with
// apps/api/business/gamification and the SQL CHECK constraints.
const (
	StatusOpen      = "open"
	StatusCompleted = "completed"
	StatusMissed    = "missed"
	StatusProtected = "protected"
)

// DailyMissionSnapshot is the canonical in-memory representation of a row in
// daily_mission_snapshots. The repository reads and writes this type.
type DailyMissionSnapshot struct {
	ID                         string
	UserID                     string
	LocalDate                  time.Time
	Timezone                   string
	ReviewTarget               int
	ReviewsCompleted           int
	NewWordTarget              *int
	NewWordsCompleted          *int
	SentencePracticeTarget     *int
	SentencePracticesCompleted *int
	PolicyVersion              string
	Status                     string
	CompletedAt                *time.Time
	GraceApplied               bool
	GraceDayID                 *string
}

// DailyActivitySummary is the canonical in-memory representation of a row in
// daily_activity_summaries.
type DailyActivitySummary struct {
	ID                     string
	UserID                 string
	LocalDate              time.Time
	Timezone               string
	ReviewsAttempted       int
	ReviewsCorrect         int
	ReviewsSkipped         int
	WordsDiscovered        int
	WordsAdded             int
	SentencesSubmitted     int
	AIFeedbackReceived     int
	ConfidencePointsEarned int
	ConfidencePointsSpent  int
}
