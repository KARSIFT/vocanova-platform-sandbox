//go:build integration

package aifeedback_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/gamification"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/learning"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/missions"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type storedOriginalProvider struct{ sentences []string }

func (p *storedOriginalProvider) GenerateFeedback(ctx context.Context, task aifeedback.ProviderTask) (*aifeedback.ProviderFeedback, error) {
	p.sentences = append(p.sentences, task.UserPayload["learner_sentence"].(string))
	if len(p.sentences) == 1 {
		return nil, errors.New("synthetic provider failure")
	}
	return aifeedback.NewMockProvider().GenerateFeedback(ctx, task)
}

func TestServicePostgreSQLRetryPreservesFirstOriginalAndSingleCredit(t *testing.T) {
	db := missionAccountingDB(t)
	ctx := t.Context()
	userID, wordID, meaningID, userWordID, reviewID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	require.NoError(t, seedReviewFeedbackTarget(ctx, db, userID, wordID, meaningID, userWordID, reviewID, time.Now().UTC()))
	repo := aifeedback.NewPostgreSQLRepository(db, nil)
	provider := &storedOriginalProvider{}
	gam := gamification.NewService(gamification.NewRepository(db))
	updater := missions.NewMissionUpdater(missions.NewService(missions.NewRepository(db), gam), gam)
	service := aifeedback.NewService(repo, provider, aifeedback.NewCompositeSafetyClassifier(aifeedback.NewDefaultLocalAbuseChecker(), aifeedback.NewMockProvider()), nil, learning.NewPostgreSQLIdempotencyStore(db), updater, nil, nil, nil, nil, aifeedback.DefaultServiceConfig())
	request := aifeedback.SubmitSentenceFeedbackRequest{UserID: userID, Source: aifeedback.SourceReview, AttemptID: reviewID, SentenceText: "  Ｉ work in London.  ", IdempotencyKey: uuid.NewString()}
	first, err := service.SubmitSentenceFeedback(ctx, request)
	require.NoError(t, err)
	require.Equal(t, aifeedback.ProcessingStatusFailed, first.ProcessingStatus)
	original := request.SentenceText
	request.SentenceText = "i WORK in LONDON."
	replay, err := service.SubmitSentenceFeedback(ctx, request)
	require.NoError(t, err)
	require.Equal(t, first.AttemptID, replay.AttemptID)
	require.Len(t, provider.sentences, 1)
	request.IdempotencyKey = uuid.NewString()
	retried, err := service.SubmitSentenceFeedback(ctx, request)
	require.NoError(t, err)
	require.Equal(t, aifeedback.ProcessingStatusCompleted, retried.ProcessingStatus)
	require.Equal(t, []string{"I work in London.", "I work in London."}, provider.sentences)
	stored, err := repo.GetLearnerSentence(ctx, userID, first.SentenceID)
	require.NoError(t, err)
	require.Equal(t, original, stored.OriginalSentence)
	_, err = repo.GetLearnerSentence(ctx, uuid.New(), first.SentenceID)
	require.ErrorIs(t, err, aifeedback.ErrTargetNotFound)
	request.IdempotencyKey = uuid.NewString()
	final, err := service.SubmitSentenceFeedback(ctx, request)
	require.NoError(t, err)
	require.Equal(t, retried.AttemptID, final.AttemptID)
	require.Len(t, provider.sentences, 2)
	var sentences, attempts, failed, succeeded, credits, received, submitted int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM learner_sentences WHERE user_id=$1`, userID).Scan(&sentences))
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER(WHERE status='failed'), count(*) FILTER(WHERE status='succeeded') FROM ai_feedback_attempts WHERE learner_sentence_id=$1`, first.SentenceID).Scan(&attempts, &failed, &succeeded))
	require.Equal(t, 1, sentences)
	require.Equal(t, 2, attempts)
	require.Equal(t, 1, failed)
	require.Equal(t, 1, succeeded)
	var normalized string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT normalized_sentence_text FROM learner_sentences WHERE id=$1`, first.SentenceID).Scan(&normalized))
	require.Equal(t, "i work in london.", normalized)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM confidence_point_ledger WHERE user_id=$1`, userID).Scan(&credits))
	require.Equal(t, 2, credits, "one sentence award and one successful feedback award")
	require.NoError(t, db.QueryRowContext(ctx, `SELECT sum(ai_feedback_received),sum(sentences_submitted) FROM daily_activity_summaries WHERE user_id=$1`, userID).Scan(&received, &submitted))
	require.Equal(t, 1, received)
	require.Equal(t, 1, submitted)
}
