package lessons

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type recommendationRepoStub struct {
	data  RecommendationData
	err   error
	owner uuid.UUID
}

func (r *recommendationRepoStub) ReadRecommendation(_ context.Context, user uuid.UUID) (RecommendationData, error) {
	r.owner = user
	return r.data, r.err
}

func recommendationFixture() RecommendationData {
	d := RecommendationData{Available: map[string]bool{}, Categories: map[string]string{}, KnownOrMastered: map[string]bool{}}
	for _, lesson := range catalog {
		d.Available[lesson.Key] = true
	}
	return d
}

func TestLessonRecommendationFocusCoverageAndStableOrder(t *testing.T) {
	d := recommendationFixture()
	d.Focus = "social"
	d.Categories[catalog[0].SituationSlug] = "social"
	for _, lesson := range catalog[1:] {
		if lesson.SituationSlug == catalog[0].SituationSlug {
			for _, word := range lesson.Words {
				d.KnownOrMastered[word.MeaningID] = true
			}
		}
	}
	d.KnownOrMastered[catalog[0].Words[0].MeaningID] = true
	d.KnownOrMastered[catalog[0].Words[1].MeaningID] = true
	r := &recommendationRepoStub{data: d}
	s := NewRecommendationService(r)
	u := uuid.New()
	got, err := s.Get(t.Context(), u)
	require.NoError(t, err)
	require.Equal(t, u, r.owner)
	require.Equal(t, catalog[0].Key, got.Recommendation.Lesson.Key, "one useful focused target outranks three outside focus")
	require.Equal(t, 1, got.Recommendation.UsefulTargetCount)
	require.Equal(t, 3, got.Recommendation.TotalTargetCount)
	require.Equal(t, "focus_and_useful_words", got.Recommendation.Reason)
	r.data.Focus = ""
	got, err = s.Get(t.Context(), u)
	require.NoError(t, err)
	require.Equal(t, "restaurant", got.Recommendation.Lesson.Key, "more useful targets, then catalog order")
	require.Equal(t, "useful_words", got.Recommendation.Reason)
	r.data = recommendationFixture()
	got, err = s.Get(t.Context(), u)
	require.NoError(t, err)
	require.Equal(t, catalog[0].Key, got.Recommendation.Lesson.Key)
}

func TestLessonRecommendationResumesSnapshotBeforeKnownCoverage(t *testing.T) {
	d := recommendationFixture()
	for _, lesson := range catalog {
		for _, w := range lesson.Words {
			d.KnownOrMastered[w.MeaningID] = true
		}
	}
	old := catalog[0]
	old.Version, old.Title = "historical", "Original lesson title"
	id := uuid.New()
	d.States = []State{{ID: id, Index: 1, Snapshot: Snapshot{Definition: old, Words: []Word{{MeaningID: catalog[0].Words[0].MeaningID}}, Steps: make([]privateStep, 3)}}}
	got, err := NewRecommendationService(&recommendationRepoStub{data: d}).Get(t.Context(), uuid.New())
	require.NoError(t, err)
	require.Equal(t, "resume", got.Recommendation.Reason)
	require.Equal(t, id.String(), got.Recommendation.Lesson.SessionID)
	require.Equal(t, "Original lesson title", got.Recommendation.Lesson.Title)
	require.Equal(t, "historical", got.Recommendation.Lesson.Version)
	require.Equal(t, 1, got.Recommendation.Lesson.CompletedSteps)
	require.Equal(t, 3, got.Recommendation.Lesson.StepCount)
	require.Zero(t, got.Recommendation.UsefulTargetCount)
	require.Equal(t, 1, got.Recommendation.TotalTargetCount)
}

func TestLessonRecommendationEmptyUnavailableAndFailureAreDistinct(t *testing.T) {
	r := &recommendationRepoStub{data: recommendationFixture()}
	s := NewRecommendationService(r)
	u := uuid.New()
	for _, lesson := range catalog {
		for _, w := range lesson.Words {
			r.data.KnownOrMastered[w.MeaningID] = true
		}
	}
	got, err := s.Get(t.Context(), u)
	require.NoError(t, err)
	require.Equal(t, "no_useful_targets", got.Status)
	require.Nil(t, got.Recommendation)
	r.data.Available[catalog[0].Key] = false
	got, err = s.Get(t.Context(), u)
	require.NoError(t, err)
	require.Equal(t, "content_unavailable", got.Status)
	require.Nil(t, got.Recommendation)
	now := time.Now()
	for _, lesson := range catalog {
		r.data.States = append(r.data.States, State{Snapshot: Snapshot{Definition: lesson}, CompletedAt: &now})
	}
	got, err = s.Get(t.Context(), u)
	require.NoError(t, err)
	require.Equal(t, "no_unfinished_lessons", got.Status)
	require.Nil(t, got.Recommendation)
	r.err = errors.New("synthetic database unavailable")
	_, err = s.Get(t.Context(), u)
	require.ErrorIs(t, err, r.err)
	_, err = s.Get(t.Context(), uuid.Nil)
	require.ErrorIs(t, err, ErrNotFound)
}
