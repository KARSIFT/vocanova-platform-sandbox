package dailylesson

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaningmedia"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// Uses committed authored examples, reviewed image manifest and actual WebP
// bytes. Metadata publication is synthetic; content eligibility is not.
func realMixedFixture(t *testing.T, meaning string) *fixture {
	f := setup(t, 0)
	raw, err := os.ReadFile("../../cmd/seed/voc026-p1.json")
	require.NoError(t, err)
	for _, q := range []string{
		`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at) SELECT (x->>'id')::uuid,x->>'text',x->>'normalized_text','en',x->>'status',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'canonical_words') x`,
		`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) SELECT (x->>'id')::uuid,(x->>'word_id')::uuid,x->>'part_of_speech',x->>'short_definition',(x->>'meaning_order')::integer,x->>'status',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'word_meanings') x`,
		`INSERT INTO word_examples(id,meaning_id,example_text,example_order,status,created_at,updated_at) SELECT (x->>'id')::uuid,(x->>'meaning_id')::uuid,x->>'example_text',(x->>'example_order')::integer,x->>'status',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP FROM jsonb_array_elements($1::jsonb->'word_examples') x`,
	} {
		_, err = f.db.Exec(q, string(raw))
		require.NoError(t, err)
	}
	raw, err = os.ReadFile("../../../web/src/lib/meaning-pictures.json")
	require.NoError(t, err)
	var manifest map[string]struct{ MeaningID, MeaningRevision, Alt string }
	require.NoError(t, json.Unmarshal(raw, &manifest))
	pubs := meaningmedia.Manifest{SchemaVersion: 1}
	for _, id := range []string{"0b593408-6679-5fb8-a21c-b73a525c76e5", "329c4ec9-6562-568e-9e71-9134da2f0d8d", "d990ce6e-4756-52e0-96e8-9926442963f2", "2534b539-a80f-5001-b9a4-27a5e0679539"} {
		data, err := os.ReadFile("../../../web/public/images/meanings/" + id + ".webp")
		require.NoError(t, err)
		h := sha256.Sum256(data)
		m := manifest[id]
		pubs.Images = append(pubs.Images, meaningmedia.Publication{MeaningID: id, MeaningRevision: m.MeaningRevision, Alt: m.Alt, SHA256: hex.EncodeToString(h[:]), Width: 768, Height: 512, Approved: true})
	}
	_, err = meaningmedia.NewService(f.db).Import(t.Context(), pubs)
	require.NoError(t, err)
	uw := uuid.New()
	_, err = f.db.Exec(`INSERT INTO user_words(id,user_id,meaning_id,status,source,added_at,created_at,updated_at)VALUES($1,$2,$3,'new','search',$4,$4,$4)`, uw, f.u, meaning, f.now)
	require.NoError(t, err)
	f.words = append(f.words, uw)
	return f
}

func TestDailyMixedRealCatalogFrozenQuestions(t *testing.T) {
	for _, tc := range []struct{ id, kind string }{{"0b593408-6679-5fb8-a21c-b73a525c76e5", "picture_choice"}, {"6a61d337-80e7-5d50-bc61-8b45b6a054a1", "gap_choice"}} {
		t.Run(tc.kind, func(t *testing.T) {
			f := realMixedFixture(t, tc.id)
			s, err := f.svc.Start(t.Context(), f.u, "mixed")
			require.NoError(t, err)
			require.Equal(t, tc.kind, s.CurrentStep.Kind)
			require.Len(t, s.CurrentStep.Choices, 4)
			if tc.kind == "gap_choice" {
				require.Empty(t, s.CurrentStep.WordText)
				require.Empty(t, s.CurrentStep.WordSlug)
				require.Empty(t, s.CurrentStep.PronunciationMeaningID)
				require.Equal(t, "Our flight leaves from _____ twelve.", s.CurrentStep.Context)
			} else {
				for _, c := range s.CurrentStep.Choices {
					require.NotNil(t, c.Image)
					require.NotEmpty(t, c.Text)
				}
			}
			before := s.CurrentStep
			_, err = f.db.Exec(`UPDATE word_meanings SET short_definition=short_definition||' Changed.' WHERE id=$1`, tc.id)
			require.NoError(t, err)
			reloaded, err := f.svc.Get(t.Context(), f.u, uuid.MustParse(s.ID))
			require.NoError(t, err)
			require.Equal(t, before, reloaded.CurrentStep)
			s = f.answer(t, s, true)
			require.True(t, s.Feedback.Correct)
			require.Equal(t, 1, s.Goal.Reviewed)
		})
	}
}

func TestDailyMixedContentEligibilityFallbacks(t *testing.T) {
	for _, tc := range []struct{ name, id, mutation string }{
		{"unapproved picture", "0b593408-6679-5fb8-a21c-b73a525c76e5", `UPDATE meaning_image_assets SET approved=false WHERE meaning_id='329c4ec9-6562-568e-9e71-9134da2f0d8d'`},
		{"missing picture", "0b593408-6679-5fb8-a21c-b73a525c76e5", `DELETE FROM meaning_image_assets WHERE meaning_id='329c4ec9-6562-568e-9e71-9134da2f0d8d'`},
		{"stale alternative", "0b593408-6679-5fb8-a21c-b73a525c76e5", `UPDATE word_meanings SET short_definition='A changed menu definition.' WHERE id='329c4ec9-6562-568e-9e71-9134da2f0d8d'`},
		{"unreviewed example", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", `UPDATE word_examples SET example_text='We went to the gate.' WHERE meaning_id='6a61d337-80e7-5d50-bc61-8b45b6a054a1'`},
		{"absent example", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", `DELETE FROM word_examples WHERE meaning_id='6a61d337-80e7-5d50-bc61-8b45b6a054a1'`},
		{"ambiguous changed candidate", "6a61d337-80e7-5d50-bc61-8b45b6a054a1", `UPDATE canonical_words SET text='terminal',normalized_text='terminal' WHERE id=(SELECT word_id FROM word_meanings WHERE id='58529fdd-9425-5aef-a8c7-bc1086be7080')`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := realMixedFixture(t, tc.id)
			_, err := f.db.Exec(tc.mutation)
			require.NoError(t, err)
			s, err := f.svc.Start(t.Context(), f.u, "fallback")
			require.NoError(t, err)
			require.Equal(t, "meaning_choice", s.CurrentStep.Kind)
			require.Len(t, s.CurrentStep.Choices, 4)
			for _, c := range s.CurrentStep.Choices {
				require.Nil(t, c.Image)
			}
			s = f.answer(t, s, false)
			require.False(t, s.Feedback.Correct)
			require.Equal(t, 1, s.Goal.Reviewed)
		})
	}
}

func TestDailyMixedWrongAnswersPreserveTeachingAndEffort(t *testing.T) {
	for _, id := range []string{"0b593408-6679-5fb8-a21c-b73a525c76e5", "6a61d337-80e7-5d50-bc61-8b45b6a054a1"} {
		t.Run(id, func(t *testing.T) {
			f := realMixedFixture(t, id)
			s, err := f.svc.Start(t.Context(), f.u, "wrong-mixed")
			require.NoError(t, err)
			frozen := f.state(t, s).Words[0]
			s = f.answer(t, s, false)
			require.False(t, s.Feedback.Correct)
			require.False(t, s.Feedback.Credited)
			require.Equal(t, 1, s.Goal.Reviewed)
			require.Equal(t, frozen.Definition, s.Feedback.Answer)
			require.Equal(t, frozen.WordText, s.Feedback.WordText)
			require.Equal(t, frozen.WordSlug, s.Feedback.WordSlug)
			require.Equal(t, mixedRevisions[id], s.Feedback.MeaningRevision)
			var points int
			require.NoError(t, f.db.QueryRow("SELECT count(*) FROM confidence_point_ledger WHERE user_id=$1", f.u).Scan(&points))
			require.Zero(t, points)
		})
	}
}
