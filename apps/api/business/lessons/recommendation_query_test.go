package lessons

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type recommendationReadFixture struct {
	focus  string
	known  []string
	states []State
}

func expectRecommendationReads(t *testing.T, mock sqlmock.Sqlmock, userID uuid.UUID, fixture recommendationReadFixture) {
	t.Helper()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE")).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"focus"}).AddRow(fixture.focus))
	states := sqlmock.NewRows(strings.Split(stateColumns, ","))
	for _, state := range fixture.states {
		snapshot, err := json.Marshal(state.Snapshot)
		require.NoError(t, err)
		states.AddRow(state.ID, userID, snapshot, state.Index, state.Revision, nil, state.FirstAnswersCorrect, state.QuestionsAnswered, state.CompletedAt)
	}
	mock.ExpectQuery("SELECT .* FROM lesson_sessions").WithArgs(userID).
		WillReturnRows(states)
	known := sqlmock.NewRows([]string{"meaning_id"})
	for _, id := range fixture.known {
		known.AddRow(id)
	}
	mock.ExpectQuery("SELECT meaning_id FROM user_word_knowledge").WithArgs(userID).
		WillReturnRows(known)
	categories := sqlmock.NewRows([]string{"slug", "category"})
	seen := map[string]bool{}
	for _, d := range catalog {
		if !seen[d.SituationSlug] {
			categories.AddRow(d.SituationSlug, "social")
			seen[d.SituationSlug] = true
		}
	}
	mock.ExpectQuery("SELECT slug,category FROM journey_situations").WillReturnRows(categories)
}

func recommendationContentRows(t *testing.T, edit func(string, Word) (string, Word, bool)) *sqlmock.Rows {
	t.Helper()
	rows := sqlmock.NewRows([]string{"situation_slug", "id", "text", "part_of_speech", "definition", "example", "usage_note"})
	// Deliberately reverse catalog order: SQL row order must not rank lessons
	// or change the teaching order inside the canonical identity check.
	for i := len(catalog) - 1; i >= 0; i-- {
		d := catalog[i]
		for _, word := range seedWords(t, d) {
			slug, include := d.SituationSlug, true
			if edit != nil {
				slug, word, include = edit(slug, word)
			}
			if include {
				rows.AddRow(slug, word.MeaningID, word.WordText, word.PartOfSpeech, word.Definition, word.Example, word.UsageNote)
			}
		}
	}
	return rows
}

func expectRecommendationContent(mock sqlmock.Sqlmock) *sqlmock.ExpectedQuery {
	ids, slugs := []string{}, []string{}
	for _, d := range catalog {
		if !containsString(slugs, d.SituationSlug) {
			slugs = append(slugs, d.SituationSlug)
		}
		for _, ref := range d.Words {
			if !containsString(ids, ref.MeaningID) {
				ids = append(ids, ref.MeaningID)
			}
		}
	}
	return mock.ExpectQuery("SELECT js.slug,wm.id").WithArgs(pq.Array(ids), pq.Array(slugs))
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func TestRecommendationRepositoryContentQueryBound(t *testing.T) {
	contentQueries := 0
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expected, actual string) error {
		if strings.Contains(actual, "FROM word_meanings wm") {
			contentQueries++
		}
		return sqlmock.QueryMatcherRegexp.Match(expected, actual)
	})))
	require.NoError(t, err)
	defer db.Close()
	userID := uuid.New()
	expectRecommendationReads(t, mock, userID, recommendationReadFixture{})
	expectRecommendationContent(mock).WillReturnRows(recommendationContentRows(t, nil))
	mock.ExpectCommit()
	data, err := NewPostgreSQLRepository(db).ReadRecommendation(t.Context(), userID)
	require.NoError(t, err)
	for _, d := range catalog {
		require.True(t, data.Available[d.Key], d.Key)
	}
	require.Equal(t, 1, contentQueries, "one content read for the entire catalog, regardless of lesson count")
	require.NoError(t, mock.ExpectationsWereMet())
	t.Logf("canonical content SELECTs: %d", contentQueries)
}

func TestRecommendationBatchPreservesCanonicalAvailabilityAndStableRanking(t *testing.T) {
	for _, scenario := range []string{"valid", "missing-target", "wrong-situation", "changed-word", "missing-definition", "missing-example"} {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			userID := uuid.New()
			fixture := recommendationReadFixture{focus: "social"}
			// Only the first lesson has useful targets. Missing content must not
			// become a claim that the learner already knows those targets.
			for _, d := range catalog[1:] {
				for _, ref := range d.Words {
					fixture.known = append(fixture.known, ref.MeaningID)
				}
			}
			expectRecommendationReads(t, mock, userID, fixture)
			rows := recommendationContentRows(t, func(slug string, word Word) (string, Word, bool) {
				if word.MeaningID != catalog[0].Words[0].MeaningID {
					return slug, word, true
				}
				switch scenario {
				case "missing-target":
					return slug, word, false
				case "wrong-situation":
					slug = "restaurant"
				case "changed-word":
					word.WordText = "a different canonical word"
				case "missing-definition":
					word.Definition = ""
				case "missing-example":
					word.Example = ""
				}
				return slug, word, true
			})
			expectRecommendationContent(mock).WillReturnRows(rows)
			mock.ExpectCommit()
			got, err := NewRecommendationService(NewPostgreSQLRepository(db)).Get(t.Context(), userID)
			require.NoError(t, err)
			if scenario == "valid" {
				expected := recommendationFixture()
				expected.Focus = fixture.focus
				for _, d := range catalog {
					expected.Categories[d.SituationSlug] = "social"
				}
				for _, id := range fixture.known {
					expected.KnownOrMastered[id] = true
				}
				want, err := NewRecommendationService(&recommendationRepoStub{data: expected}).Get(t.Context(), userID)
				require.NoError(t, err)
				require.Equal(t, want, got)
			} else {
				require.Equal(t, "content_unavailable", got.Status)
				require.Nil(t, got.Recommendation)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRecommendationBatchPreservesMostRecentHistoricalResume(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	userID := uuid.New()
	oldDefinition := catalog[0]
	oldDefinition.Title, oldDefinition.Version = "Original saved lesson", "historical"
	snapshot, err := buildSnapshot(oldDefinition, seedWords(t, oldDefinition), "saved-session")
	require.NoError(t, err)
	completedSnapshot, err := buildSnapshot(catalog[6], seedWords(t, catalog[6]), "completed-session")
	require.NoError(t, err)
	olderSnapshot, err := buildSnapshot(catalog[1], seedWords(t, catalog[1]), "older-session")
	require.NoError(t, err)
	now := time.Now()
	fixture := recommendationReadFixture{focus: "social", states: []State{
		{ID: uuid.New(), Snapshot: snapshot, Index: 2, Revision: 3},
		{ID: uuid.New(), Snapshot: completedSnapshot, Index: len(completedSnapshot.Steps), CompletedAt: &now},
		{ID: uuid.New(), Snapshot: olderSnapshot, Index: 1},
	}}
	for _, ref := range oldDefinition.Words {
		fixture.known = append(fixture.known, ref.MeaningID)
	}
	expectRecommendationReads(t, mock, userID, fixture)
	// Archived current content must not replace or block a resumable snapshot.
	expectRecommendationContent(mock).WillReturnRows(recommendationContentRows(t, func(slug string, word Word) (string, Word, bool) {
		return slug, word, word.MeaningID != oldDefinition.Words[0].MeaningID
	}))
	mock.ExpectCommit()
	got, err := NewRecommendationService(NewPostgreSQLRepository(db)).Get(t.Context(), userID)
	require.NoError(t, err)
	require.Equal(t, "resume", got.Recommendation.Reason)
	require.Equal(t, fixture.states[0].ID.String(), got.Recommendation.Lesson.SessionID)
	require.Equal(t, oldDefinition.Title, got.Recommendation.Lesson.Title)
	require.Equal(t, oldDefinition.Version, got.Recommendation.Lesson.Version)
	require.Equal(t, 2, got.Recommendation.Lesson.CompletedSteps)
	require.Zero(t, got.Recommendation.UsefulTargetCount)
	require.Equal(t, snapshot, fixture.states[0].Snapshot)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecommendationBatchReadFailuresRollBack(t *testing.T) {
	for _, scenario := range []string{"query", "row"} {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			userID := uuid.New()
			expectRecommendationReads(t, mock, userID, recommendationReadFixture{})
			failure := errors.New("synthetic content read failure")
			expected := expectRecommendationContent(mock)
			if scenario == "query" {
				expected.WillReturnError(failure)
			} else {
				expected.WillReturnRows(recommendationContentRows(t, nil).RowError(1, failure))
			}
			mock.ExpectRollback()
			_, err = NewRecommendationService(NewPostgreSQLRepository(db)).Get(t.Context(), userID)
			require.ErrorIs(t, err, failure)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
