package aifeedback

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Read the shipped content rather than inventing dictionary entries. Loading
// each target through the repository also verifies that accepted forms reach
// the real validation path, rather than only existing in test-built Targets.
func TestSeededTargetsAcceptExamplesAndApprovedVariants(t *testing.T) {
	var seed struct {
		Words []struct {
			ID         string
			Text       string
			Normalized string `json:"normalized_text"`
			Type       string `json:"word_type"`
		} `json:"canonical_words"`
		Meanings []struct {
			ID         string
			WordID     string `json:"word_id"`
			POS        string `json:"part_of_speech"`
			Definition string `json:"short_definition"`
		} `json:"word_meanings"`
		Examples []struct {
			MeaningID string `json:"meaning_id"`
			Text      string `json:"example_text"`
		} `json:"word_examples"`
	}
	data, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &seed))
	require.NotEmpty(t, seed.Examples)
	variants := map[string]struct{ accepted, rejected string }{
		"syllabus":      {"The syllabi are available online.", "The syllabic structure is complex."},
		"follow-up":     {"I sent two follow-ups.", "I sent two follow-updates."},
		"check-out":     {"The hotel has late check-outs.", "The hotel has late check-outings."},
		"catch up":      {"I caught up with Maya yesterday.", "I caught the bus yesterday."},
		"meet up":       {"We met up after class.", "We met after class yesterday."},
		"keep in touch": {"We kept in touch after school.", "We lost touch after school."},
		"sounds good":   {"Those plans sound good to me.", "That good sound is music."},
		"cancel":        {"I cancelled our dinner booking yesterday.", "The cancellation came too late."},
	}
	checkedExamples, checkedVariants := 0, 0
	for _, meaning := range seed.Meanings {
		for _, word := range seed.Words {
			if word.ID != meaning.WordID {
				continue
			}
			t.Run(word.Text+"/"+meaning.ID, func(t *testing.T) {
				db, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer db.Close()
				userID, userWordID := uuid.New(), uuid.New()
				mock.ExpectQuery("SELECT cw.id, cw.text, cw.normalized_text, cw.word_type, cw.difficulty_level").
					WithArgs(userWordID, userID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "text", "normalized_text", "word_type", "difficulty_level", "meaning_id", "part_of_speech", "short_definition", "user_word_id"}).
						AddRow(word.ID, word.Text, word.Normalized, word.Type, nil, meaning.ID, meaning.POS, meaning.Definition, userWordID))
				target, err := NewPostgreSQLRepository(db, nil).LoadTarget(t.Context(), LoadTargetRequest{UserID: userID, Source: SourceWordDetail, AttemptID: userWordID})
				require.NoError(t, err)
				require.NoError(t, mock.ExpectationsWereMet())
				for _, example := range seed.Examples {
					if example.MeaningID != meaning.ID {
						continue
					}
					checkedExamples++
					result := ValidateSentence(example.Text, target)
					require.True(t, result.Valid, "seed example %q rejected: %s", example.Text, result.Code)
				}
				if variant, ok := variants[word.Normalized]; ok {
					checkedVariants++
					result := ValidateSentence(variant.accepted, target)
					require.True(t, result.Valid, "approved variant %q rejected: %s", variant.accepted, result.Code)
					require.Equal(t, ValidationCodeMissingTarget, ValidateSentence(variant.rejected, target).Code)
				}
			})
		}
	}
	require.Equal(t, len(seed.Examples), checkedExamples, "every shipped example must have a loaded target")
	require.Equal(t, 9, checkedVariants, "four noun meaning pairs and five Daily Conversation targets")
	t.Logf("validated %d shipped examples and %d approved variant/meaning pairs", checkedExamples, checkedVariants)
}

func TestApprovedNounVariantsDoNotExpandOtherPartsOfSpeech(t *testing.T) {
	require.NotContains(t, BuildAcceptedForms("syllabus", "word", "verb"), "syllabi")
	require.NotContains(t, BuildAcceptedForms("follow-up", "phrase", "verb"), "follow-ups")
	require.NotContains(t, BuildAcceptedForms("check-out", "phrasal_verb", "verb"), "check-outs")
	require.Equal(t, []string{"follow-up"}, BuildAcceptedForms("follow-up", "idiom", "noun"))
}
