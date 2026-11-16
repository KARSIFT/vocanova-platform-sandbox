package dailylesson

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaningmedia"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

// Uses committed authored examples, reviewed image manifest and actual WebP
// bytes. Metadata publication is synthetic; content eligibility is not.
func realMixedFixture(t *testing.T, meaning string) *fixture {
	f := setup(t, 0)
	f.svc.questionEntropy = mixedZeroReader{}
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
			f.svc.questionEntropy = bytes.NewReader(nil)
			replay, err := f.svc.Start(t.Context(), f.u, "mixed")
			require.NoError(t, err)
			require.Equal(t, before, replay.CurrentStep)
			s = f.answer(t, s, true)
			require.True(t, s.Feedback.Correct)
			require.Equal(t, 1, s.Goal.Reviewed)
		})
	}
}

func TestDailyMixedContentEligibilityFallbacks(t *testing.T) {
	for _, tc := range []struct{ name, id, mutation string }{
		{"unapproved target", "0b593408-6679-5fb8-a21c-b73a525c76e5", `UPDATE meaning_image_assets SET approved=false WHERE meaning_id='0b593408-6679-5fb8-a21c-b73a525c76e5'`},
		{"stale target", "0b593408-6679-5fb8-a21c-b73a525c76e5", `UPDATE word_meanings SET short_definition='A newly edited parcel definition.' WHERE id='0b593408-6679-5fb8-a21c-b73a525c76e5'`},
		{"same image bytes", "0b593408-6679-5fb8-a21c-b73a525c76e5", `UPDATE meaning_image_assets SET sha256=(SELECT sha256 FROM meaning_image_assets WHERE meaning_id='0b593408-6679-5fb8-a21c-b73a525c76e5') WHERE meaning_id='329c4ec9-6562-568e-9e71-9134da2f0d8d'`},
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

// An injected source avoids statistical/flaky expectations while production
// always uses crypto/rand. Choice position remains independently randomized.
type mixedZeroReader struct{}

func (mixedZeroReader) Read(p []byte) (int, error) { clear(p); return len(p), nil }

func TestDailyMixedAllEligibleKindsUsePrivateDraw(t *testing.T) {
	for _, tc := range []struct {
		draw byte
		kind string
	}{{0, "picture_choice"}, {60, "gap_choice"}, {90, "meaning_choice"}} {
		t.Run(tc.kind, func(t *testing.T) {
			f := realMixedFixture(t, "d990ce6e-4756-52e0-96e8-9926442963f2")
			f.svc.questionEntropy = io.MultiReader(bytes.NewReader([]byte{tc.draw}), mixedZeroReader{})
			s, err := f.svc.Start(t.Context(), f.u, "weighted")
			require.NoError(t, err)
			require.Equal(t, tc.kind, s.CurrentStep.Kind)
			before := s.CurrentStep
			f.svc.questionEntropy = bytes.NewReader(nil)
			s, err = f.svc.Start(t.Context(), f.u, "weighted")
			require.NoError(t, err)
			require.Equal(t, before, s.CurrentStep)
		})
	}
}

func TestDailyMixedPicturesOutsideFormerAllowlist(t *testing.T) {
	for _, target := range []string{"debris", "wit"} {
		t.Run(target, func(t *testing.T) {
			f := realMixedFixture(t, "0b593408-6679-5fb8-a21c-b73a525c76e5")
			file, err := os.Open("../knowledgecatalog/catalog.json.gz")
			require.NoError(t, err)
			defer file.Close()
			reader, err := gzip.NewReader(file)
			require.NoError(t, err)
			defer reader.Close()
			var data struct {
				Items []knowledgecatalog.CatalogEntry `json:"items"`
			}
			require.NoError(t, json.NewDecoder(reader).Decode(&data))
			var entry knowledgecatalog.CatalogEntry
			for _, e := range data.Items {
				if e.Word == target {
					entry = e
					break
				}
			}
			require.NotEqual(t, uuid.Nil, entry.MeaningID)
			require.NotContains(t, mixedRevisions, entry.MeaningID.String())
			_, err = f.db.Exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at) VALUES($1,$2,$2,'en','active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, entry.WordID, entry.Word)
			require.NoError(t, err)
			_, err = f.db.Exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES($1,$2,$3,$4,1,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, entry.MeaningID, entry.WordID, entry.PartOfSpeech, entry.Definition)
			require.NoError(t, err)
			// This publication is synthetic eligibility evidence, not a live asset
			// assertion. The original four candidates use committed actual WebP hashes.
			digest := sha256.Sum256([]byte(target))
			_, err = meaningmedia.NewService(f.db).Import(t.Context(), meaningmedia.Manifest{SchemaVersion: 1, Images: []meaningmedia.Publication{{MeaningID: entry.MeaningID.String(), MeaningRevision: meaningidentity.Revision(entry.MeaningID.String(), entry.Word, entry.PartOfSpeech, entry.Definition), SHA256: hex.EncodeToString(digest[:]), Alt: "Synthetic test illustration", Width: 768, Height: 512, Approved: true}}})
			require.NoError(t, err)
			_, err = f.db.Exec(`UPDATE user_words SET meaning_id=$1 WHERE user_id=$2`, entry.MeaningID, f.u)
			require.NoError(t, err)
			s, err := f.svc.Start(t.Context(), f.u, "outside")
			require.NoError(t, err)
			require.Equal(t, "picture_choice", s.CurrentStep.Kind)
			require.Len(t, s.CurrentStep.Choices, 4)
			targetSeen := false
			ids := map[string]bool{}
			hashes := map[string]bool{}
			for _, c := range s.CurrentStep.Choices {
				require.NotNil(t, c.Image)
				require.False(t, ids[c.Image.MeaningID])
				require.False(t, hashes[c.Image.SHA256])
				ids[c.Image.MeaningID] = true
				hashes[c.Image.SHA256] = true
				if c.Image.MeaningID == entry.MeaningID.String() {
					targetSeen = true
					require.Equal(t, entry.Definition, c.Text)
				}
			}
			require.True(t, targetSeen)
			s = f.answer(t, s, false)
			require.False(t, s.Feedback.Correct)
			require.False(t, s.Feedback.Credited)
			require.Equal(t, 1, s.Goal.Reviewed)
		})
	}
}

func TestDailyMixedEveryReviewedGapIsAvailableAndHidesTarget(t *testing.T) {
	require.Len(t, gapSets, 12)
	for id, gap := range gapSets {
		t.Run(id, func(t *testing.T) {
			f := realMixedFixture(t, id)
			// Disable all pictures so the draw selects the reviewed sentence mode.
			_, err := f.db.Exec(`UPDATE meaning_image_assets SET approved=false`)
			require.NoError(t, err)
			s, err := f.svc.Start(t.Context(), f.u, "all-gaps")
			require.NoError(t, err)
			require.Equal(t, "gap_choice", s.CurrentStep.Kind)
			require.Len(t, s.CurrentStep.Choices, 4)
			frozen := f.state(t, s).Words[0]
			want, ok := mixedCloze(gap.Example, frozen.WordText)
			require.True(t, ok)
			require.Equal(t, want, s.CurrentStep.Context)
			require.Empty(t, s.CurrentStep.WordText)
			require.Empty(t, s.CurrentStep.WordSlug)
			require.Empty(t, s.CurrentStep.PronunciationMeaningID)
			for _, c := range s.CurrentStep.Choices {
				require.Nil(t, c.Image)
			}
			s = f.answer(t, s, true)
			require.True(t, s.Feedback.Correct)
			require.Equal(t, frozen.WordText, s.Feedback.WordText)
		})
	}
}

func TestDailyMixedUnavailablePictureUsesEligibleGap(t *testing.T) {
	f := realMixedFixture(t, "d990ce6e-4756-52e0-96e8-9926442963f2")
	_, err := f.db.Exec(`UPDATE meaning_image_assets SET approved=false`)
	require.NoError(t, err)
	s, err := f.svc.Start(t.Context(), f.u, "gap-fallback")
	require.NoError(t, err)
	require.Equal(t, "gap_choice", s.CurrentStep.Kind)
	require.Equal(t, "My _____ stopped working.", s.CurrentStep.Context)
}

func TestDailyMixedCurrentTargetImageSurvivesOlderPublications(t *testing.T) {
	const target = "0b593408-6679-5fb8-a21c-b73a525c76e5"
	f := realMixedFixture(t, target)
	// Seventeen obsolete revisions sort before the current revision. They must
	// never hide the exact current publication, regardless of historical count.
	_, err := f.db.Exec(`INSERT INTO meaning_image_assets(id,meaning_id,meaning_revision,sha256,alt,width,height,approved,created_at)
 SELECT gen_random_uuid(),meaning_id,lpad(to_hex(n),64,'0'),sha256,alt,width,height,true,CURRENT_TIMESTAMP
 FROM meaning_image_assets CROSS JOIN generate_series(1,17) n WHERE meaning_id=$1`, target)
	require.NoError(t, err)
	s, err := f.svc.Start(t.Context(), f.u, "current-target-image")
	require.NoError(t, err)
	require.Equal(t, "picture_choice", s.CurrentStep.Kind)
	for _, choice := range s.CurrentStep.Choices {
		require.NotNil(t, choice.Image)
		if choice.Image.MeaningID == target {
			require.Equal(t, mixedRevisions[target], choice.Image.MeaningRevision)
		}
	}
}

func TestDailyMixedSyllabusGapExcludesPlausibleContainer(t *testing.T) {
	f := realMixedFixture(t, "58529fdd-9425-5aef-a8c7-bc1086be7080")
	s, err := f.svc.Start(t.Context(), f.u, "syllabus-gap")
	require.NoError(t, err)
	require.Equal(t, "gap_choice", s.CurrentStep.Kind)
	texts := make([]string, 0, 4)
	for _, choice := range s.CurrentStep.Choices {
		texts = append(texts, choice.Text)
	}
	// A reading list can ordinarily be in luggage, so that answer is ambiguous.
	require.NotContains(t, texts, "luggage")
	require.ElementsMatch(t, []string{"syllabus", "gate", "refund", "appetizer"}, texts)
}
