package dailylesson

import (
	"context"
	"database/sql"
	"strings"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaningmedia"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// These small editorial sets are pinned to authored catalog revisions, not
// inferred from arbitrary examples/images. See mixed_questions.md for rationale.
// Any changed/missing definition, example, image or synonym causes text fallback.
var mixedRevisions = map[string]string{
	"0b593408-6679-5fb8-a21c-b73a525c76e5": "58371613d889e51873a9f32f55ab342ad6c6850bfb679bdc2bb2cd92a2de42fb", // parcel
	"329c4ec9-6562-568e-9e71-9134da2f0d8d": "ae7b8d69c5c0f06a7955f25b741fb7ff6be8ceb11a0ba69d64d8d3c79ecd7c01", // menu
	"d990ce6e-4756-52e0-96e8-9926442963f2": "068a52d16045d06ce7f8e5fe0575add338bd96df5e76aee46359ecf350d6fa7d", // key card
	"2534b539-a80f-5001-b9a4-27a5e0679539": "445da3734da17ce98a160e46176790438f40b9b0d472f7a8e41efae2eaff4d84", // appetizer
	"6a61d337-80e7-5d50-bc61-8b45b6a054a1": "112a5a995ccdcaa2b3b3b231a0dcc165926b98b4732daf032eae37e0d0170a0b", // gate
	"58529fdd-9425-5aef-a8c7-bc1086be7080": "d20c054bc0ae66bbb38f6ea80c4904c672628e68fb0dabb79881931b58e0be40", // syllabus
	"bee2408e-5991-5805-923c-6fba33ab7e7a": "5c833aeac56871b2a95ce70f21333818195df88b644d6e808335709545770e5b", // luggage
}
var pictureSet = []string{"0b593408-6679-5fb8-a21c-b73a525c76e5", "329c4ec9-6562-568e-9e71-9134da2f0d8d", "d990ce6e-4756-52e0-96e8-9926442963f2", "2534b539-a80f-5001-b9a4-27a5e0679539"}

type gapSet struct {
	Example string
	IDs     []string
}

var gapSets = map[string]gapSet{
	"6a61d337-80e7-5d50-bc61-8b45b6a054a1": {"Our flight leaves from gate twelve.", []string{"6a61d337-80e7-5d50-bc61-8b45b6a054a1", "58529fdd-9425-5aef-a8c7-bc1086be7080", "bee2408e-5991-5805-923c-6fba33ab7e7a", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
	"d990ce6e-4756-52e0-96e8-9926442963f2": {"My key card stopped working.", []string{"d990ce6e-4756-52e0-96e8-9926442963f2", "58529fdd-9425-5aef-a8c7-bc1086be7080", "bee2408e-5991-5805-923c-6fba33ab7e7a", "2534b539-a80f-5001-b9a4-27a5e0679539"}},
}

type mixedCandidate struct {
	ID, Word, POS, Definition string
	Image                     *meaningmedia.Image
}

func mixedCandidates(ctx context.Context, tx *sql.Tx, ids []string) (map[string]mixedCandidate, error) {
	rows, err := tx.QueryContext(ctx, `SELECT wm.id,cw.text,wm.part_of_speech,wm.short_definition FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id WHERE wm.id=ANY($1::uuid[]) AND wm.status='active' AND cw.status='active'`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]mixedCandidate{}
	for rows.Next() {
		var c mixedCandidate
		if err = rows.Scan(&c.ID, &c.Word, &c.POS, &c.Definition); err != nil {
			return nil, err
		}
		if meaningidentity.Revision(c.ID, c.Word, c.POS, c.Definition) == mixedRevisions[c.ID] {
			out[c.ID] = c
		}
	}
	return out, rows.Err()
}
func approvedPictures(ctx context.Context, tx *sql.Tx, candidates map[string]mixedCandidate, ids []string) error {
	rows, err := tx.QueryContext(ctx, `SELECT meaning_id,meaning_revision,sha256,alt,width,height FROM meaning_image_assets WHERE meaning_id=ANY($1::uuid[]) AND approved=true`, pq.Array(ids))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var p meaningmedia.Publication
		if err = rows.Scan(&p.MeaningID, &p.MeaningRevision, &p.SHA256, &p.Alt, &p.Width, &p.Height); err != nil {
			return err
		}
		p.Approved = true
		c, ok := candidates[p.MeaningID]
		if !ok || p.Validate() != nil || mixedRevisions[p.MeaningID] != p.MeaningRevision {
			continue
		}
		c.Image = &meaningmedia.Image{MeaningID: p.MeaningID, MeaningRevision: p.MeaningRevision, SHA256: p.SHA256, ImageURL: meaningmedia.PublicOrigin + "/" + p.ObjectKey(), Alt: p.Alt, Width: p.Width, Height: p.Height}
		candidates[p.MeaningID] = c
	}
	return rows.Err()
}
func safeMixedSet(candidates map[string]mixedCandidate, ids []string) bool {
	if len(ids) != 4 || len(candidates) != 4 {
		return false
	}
	seen := map[string]bool{}
	for i, id := range ids {
		c, ok := candidates[id]
		if !ok {
			return false
		}
		key := strings.ToLower(strings.TrimSpace(c.Word))
		if seen[key] {
			return false
		}
		seen[key] = true
		for _, other := range ids[:i] {
			alias, err := knowledgecatalog.SemanticAliases(c.Word, candidates[other].Word)
			if err != nil || alias {
				return false
			}
		}
	}
	return true
}
func (s *Service) mixQuestion(ctx context.Context, tx *sql.Tx, w *word, index int) error {
	_, hasGap := gapSets[w.MeaningID]
	hasPicture := false
	for _, id := range pictureSet {
		if id == w.MeaningID {
			hasPicture = true
		}
	}
	if !hasGap && !hasPicture {
		return nil
	}
	kinds := []string{"picture_choice", "gap_choice"}
	if index%2 == 1 {
		kinds = []string{"gap_choice", "picture_choice"}
	}
	for _, kind := range kinds {
		ids := pictureSet
		gap := gapSets[w.MeaningID]
		if kind == "gap_choice" {
			if !hasGap {
				continue
			}
			ids = gap.IDs
		} else if !hasPicture {
			continue
		}
		candidates, err := mixedCandidates(ctx, tx, ids)
		if err != nil {
			return err
		}
		if !safeMixedSet(candidates, ids) {
			continue
		}
		if kind == "gap_choice" {
			var exists bool
			err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM word_examples WHERE meaning_id=$1 AND example_text=$2 AND status='active')`, w.MeaningID, gap.Example).Scan(&exists)
			if err != nil {
				return err
			}
			if !exists || strings.Count(gap.Example, w.WordText) != 1 {
				continue
			}
		} else {
			if err = approvedPictures(ctx, tx, candidates, ids); err != nil {
				return err
			}
			all := true
			for _, c := range candidates {
				if c.Image == nil {
					all = false
				}
			}
			if !all {
				continue
			}
		}
		choices := make([]Choice, 0, 4)
		for _, id := range ids {
			c := candidates[id]
			choice := Choice{ID: uuid.NewString(), Text: c.Definition, Image: c.Image}
			if kind == "gap_choice" {
				choice.Text = c.Word
			}
			if id == w.MeaningID {
				w.AnswerID = choice.ID
			}
			choices = append(choices, choice)
		}
		w.Public.Kind = kind
		w.Public.Choices = choices
		if kind == "gap_choice" {
			w.Public.Context = strings.Replace(gap.Example, w.WordText, "_____", 1)
			w.Public.WordText = ""
			w.Public.WordSlug = ""
		}
		return nil
	}
	return nil
}
