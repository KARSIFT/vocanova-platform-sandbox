package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestApplySeedExecutesUpsertStatementsInOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	displayOrder := 1
	seed := seedData{
		JourneySituations: []journeySituation{
			{
				ID:               "00000000-0000-0000-0000-000000000001",
				Slug:             "test-situation",
				Title:            "Test Situation",
				ShortDescription: "A test situation.",
				Category:         "daily_life",
				Status:           "active",
				DisplayOrder:     1,
			},
		},
		CanonicalWords: []canonicalWord{
			{
				ID:             "00000000-0000-0000-0000-000000000002",
				Text:           "test word",
				NormalizedText: "test word",
				WordType:       "word",
				LanguageCode:   "en",
				Status:         "active",
			},
		},
		WordMeanings: []wordMeaning{
			{
				ID:              "00000000-0000-0000-0000-000000000003",
				WordID:          "00000000-0000-0000-0000-000000000002",
				PartOfSpeech:    "noun",
				ShortDefinition: "A test word.",
				MeaningOrder:    1,
				Status:          "active",
			},
		},
		WordExamples: []wordExample{
			{
				ID:           "00000000-0000-0000-0000-000000000004",
				MeaningID:    "00000000-0000-0000-0000-000000000003",
				ExampleText:  "This is a test word.",
				ExampleOrder: 1,
				Status:       "active",
			},
		},
		UsageNotes: []usageNote{
			{
				ID:        "00000000-0000-0000-0000-000000000005",
				MeaningID: "00000000-0000-0000-0000-000000000003",
				NoteType:  "collocation",
				NoteText:  "a test word",
				NoteOrder: 1,
				Status:    "active",
			},
		},
		JourneyWords: []journeyWord{
			{
				ID:                 "00000000-0000-0000-0000-000000000006",
				JourneySituationID: "00000000-0000-0000-0000-000000000001",
				MeaningID:          "00000000-0000-0000-0000-000000000003",
				RelevanceScore:     80,
				DisplayOrder:       &displayOrder,
				IsCore:             true,
			},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO journey_situations").
		WithArgs(seed.JourneySituations[0].ID, seed.JourneySituations[0].Slug, seed.JourneySituations[0].Title, seed.JourneySituations[0].ShortDescription, nil, seed.JourneySituations[0].Category, seed.JourneySituations[0].Status, seed.JourneySituations[0].DisplayOrder).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO canonical_words").
		WithArgs(seed.CanonicalWords[0].ID, seed.CanonicalWords[0].Text, seed.CanonicalWords[0].NormalizedText, seed.CanonicalWords[0].WordType, seed.CanonicalWords[0].LanguageCode, seed.CanonicalWords[0].Status, nil, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO word_meanings").
		WithArgs(seed.WordMeanings[0].ID, seed.WordMeanings[0].WordID, seed.WordMeanings[0].PartOfSpeech, seed.WordMeanings[0].ShortDefinition, nil, seed.WordMeanings[0].MeaningOrder, seed.WordMeanings[0].Status, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO word_examples").
		WithArgs(seed.WordExamples[0].ID, seed.WordExamples[0].MeaningID, seed.WordExamples[0].ExampleText, seed.WordExamples[0].ExampleOrder, nil, nil, seed.WordExamples[0].Status).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO usage_notes").
		WithArgs(seed.UsageNotes[0].ID, seed.UsageNotes[0].MeaningID, seed.UsageNotes[0].NoteType, seed.UsageNotes[0].NoteText, seed.UsageNotes[0].NoteOrder, seed.UsageNotes[0].Status).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO journey_words").
		WithArgs(seed.JourneyWords[0].ID, seed.JourneyWords[0].JourneySituationID, seed.JourneyWords[0].MeaningID, seed.JourneyWords[0].RelevanceScore, seed.JourneyWords[0].DisplayOrder, seed.JourneyWords[0].IsCore).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tx, err := db.Begin()
	require.NoError(t, err)

	require.NoError(t, applySeed(tx, seed))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadSeedParsesEmbeddedJSON(t *testing.T) {
	seed, err := loadSeed()
	require.NoError(t, err)
	require.Len(t, seed.JourneySituations, 7)
	require.Len(t, seed.CanonicalWords, 39)
	require.Len(t, seed.WordMeanings, 42)
	require.Len(t, seed.WordExamples, 42)
	require.Len(t, seed.UsageNotes, 126)
	require.Len(t, seed.JourneyWords, 42)
}
