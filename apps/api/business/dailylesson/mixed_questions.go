package dailylesson

import (
	"context"
	"crypto/rand"
	"database/sql"
	"io"
	"math/big"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/knowledgecatalog"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaningmedia"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type mixedCandidate struct {
	ID, Word, POS, Definition, SourceSense string
	Image                                  *meaningmedia.Image
}

// Picture availability is catalog-wide. Bound the returned publication pool,
// including old revisions, before doing semantic comparisons in Go. No random
// database sort, external request or corpus-wide materialization happens here.
const mixedPicturePoolSQL = `SELECT wm.id,cw.text,wm.part_of_speech,wm.short_definition,COALESCE(k.source_sense_key,''),a.meaning_revision,a.sha256,a.alt,a.width,a.height
 FROM meaning_image_assets a JOIN word_meanings wm ON wm.id=a.meaning_id JOIN canonical_words cw ON cw.id=wm.word_id
 LEFT JOIN knowledge_catalog_entries k ON k.meaning_id=wm.id
 WHERE a.approved=true AND wm.status='active' AND cw.status='active' AND wm.part_of_speech=$1 AND wm.id<>$2
 ORDER BY a.meaning_id,a.meaning_revision LIMIT 64`

func scanMixedPicture(rows *sql.Rows) (mixedCandidate, error) {
	var c mixedCandidate
	var p meaningmedia.Publication
	err := rows.Scan(&c.ID, &c.Word, &c.POS, &c.Definition, &c.SourceSense, &p.MeaningRevision, &p.SHA256, &p.Alt, &p.Width, &p.Height)
	if err != nil {
		return c, err
	}
	p.MeaningID, p.Approved = c.ID, true
	if p.Validate() == nil && meaningidentity.Revision(c.ID, c.Word, c.POS, c.Definition) == p.MeaningRevision {
		c.Image = &meaningmedia.Image{MeaningID: c.ID, MeaningRevision: p.MeaningRevision, SHA256: p.SHA256, ImageURL: meaningmedia.PublicOrigin + "/" + p.ObjectKey(), Alt: p.Alt, Width: p.Width, Height: p.Height}
	}
	return c, nil
}

func mixedPictures(ctx context.Context, tx *sql.Tx, w *word) ([]mixedCandidate, error) {
	// The target must have its own current approved image. A similar picture or
	// another sense of the same spelling cannot stand in for that publication.
	revision := meaningidentity.Revision(w.MeaningID, w.WordText, w.Public.PartOfSpeech, w.Definition)
	rows, err := tx.QueryContext(ctx, `SELECT wm.id,cw.text,wm.part_of_speech,wm.short_definition,COALESCE(k.source_sense_key,''),a.meaning_revision,a.sha256,a.alt,a.width,a.height
 FROM meaning_image_assets a JOIN word_meanings wm ON wm.id=a.meaning_id JOIN canonical_words cw ON cw.id=wm.word_id
 LEFT JOIN knowledge_catalog_entries k ON k.meaning_id=wm.id
 WHERE a.meaning_id=$1 AND a.meaning_revision=$2 AND a.approved=true AND wm.status='active' AND cw.status='active'`, w.MeaningID, revision)
	if err != nil {
		return nil, err
	}
	var target mixedCandidate
	for rows.Next() {
		c, e := scanMixedPicture(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		if c.Image != nil && c.Word == w.WordText && c.POS == w.Public.PartOfSpeech && c.Definition == w.Definition {
			target = c
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if target.Image == nil {
		return nil, nil
	}
	rows, err = tx.QueryContext(ctx, mixedPicturePoolSQL, target.POS, target.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// A mutually safe subset allows private sampling without accidentally pairing
	// two aliases or two copies of the same image as distractors.
	out := []mixedCandidate{target}
	for rows.Next() {
		c, e := scanMixedPicture(rows)
		if e != nil {
			return nil, e
		}
		if c.Image == nil {
			continue
		}
		safe := true
		for _, other := range out {
			if !safeMixedPair(c, other) {
				safe = false
				break
			}
		}
		if safe && len(out) < 16 {
			out = append(out, c)
		}
	}
	return out, rows.Err()
}

func safeMixedPair(a, b mixedCandidate) bool {
	if a.ID == b.ID || a.POS != b.POS || strings.EqualFold(strings.TrimSpace(a.Word), strings.TrimSpace(b.Word)) || strings.EqualFold(strings.TrimSpace(a.Definition), strings.TrimSpace(b.Definition)) || a.SourceSense != "" && a.SourceSense == b.SourceSense {
		return false
	}
	if a.Image != nil && b.Image != nil && a.Image.SHA256 == b.Image.SHA256 {
		return false
	}
	alias, err := knowledgecatalog.SemanticAliases(a.Word, b.Word)
	return err == nil && !alias
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

func safeMixedSet(candidates map[string]mixedCandidate, ids []string) bool {
	if len(ids) != 4 || len(candidates) != 4 {
		return false
	}
	for i, id := range ids {
		c, ok := candidates[id]
		if !ok {
			return false
		}
		for _, other := range ids[:i] {
			if !safeMixedPair(c, candidates[other]) {
				return false
			}
		}
	}
	return true
}

func mixedGap(ctx context.Context, tx *sql.Tx, w *word) ([]mixedCandidate, string, error) {
	gap, ok := gapSets[w.MeaningID]
	if !ok {
		return nil, "", nil
	}
	candidates, err := mixedCandidates(ctx, tx, gap.IDs)
	if err != nil {
		return nil, "", err
	}
	if !safeMixedSet(candidates, gap.IDs) {
		return nil, "", nil
	}
	target := candidates[w.MeaningID]
	if target.Word != w.WordText || target.POS != w.Public.PartOfSpeech || target.Definition != w.Definition {
		return nil, "", nil
	}
	context, ok := mixedCloze(gap.Example, w.WordText)
	if !ok {
		return nil, "", nil
	}
	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM word_examples WHERE meaning_id=$1 AND example_text=$2 AND status='active')`, w.MeaningID, gap.Example).Scan(&exists)
	if err != nil {
		return nil, "", err
	}
	if !exists {
		return nil, "", nil
	}
	out := make([]mixedCandidate, 0, 4)
	for _, id := range gap.IDs {
		out = append(out, candidates[id])
	}
	return out, context, nil
}

// A cloze cannot split a word, possessive, or hyphenated form, and cannot leave
// another complete occurrence of its answer visible. Exact sentences are also
// editorially pinned; this lexical guard alone does not prove contextual safety.
func mixedCloze(sentence, target string) (string, bool) {
	if target == "" {
		return "", false
	}
	matches := regexp.MustCompile(`(?i)`+regexp.QuoteMeta(target)).FindAllStringIndex(sentence, -1)
	var selected []int
	boundary := func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' || r == '’' || r == '-' || r == '_'
	}
	for _, m := range matches {
		if m[0] > 0 {
			r, _ := utf8.DecodeLastRuneInString(sentence[:m[0]])
			if boundary(r) {
				continue
			}
		}
		if m[1] < len(sentence) {
			r, _ := utf8.DecodeRuneInString(sentence[m[1]:])
			if boundary(r) {
				continue
			}
		}
		if selected != nil {
			return "", false
		}
		selected = m
	}
	if selected == nil {
		return "", false
	}
	return sentence[:selected[0]] + "_____" + sentence[selected[1]:], true
}

// Weights are renormalized over eligible modes. A catalog without a safe cloze
// therefore still gets pictures frequently (60/70), with occasional text.
func chooseMixedKind(entropy io.Reader, picture, gap bool) (string, error) {
	if !picture && !gap {
		return "meaning_choice", nil
	}
	if entropy == nil {
		entropy = rand.Reader
	}
	total := 10
	if picture {
		total += 60
	}
	if gap {
		total += 30
	}
	draw, err := rand.Int(entropy, big.NewInt(int64(total)))
	if err != nil {
		return "", err
	}
	n := int(draw.Int64())
	if picture {
		if n < 60 {
			return "picture_choice", nil
		}
		n -= 60
	}
	if gap && n < 30 {
		return "gap_choice", nil
	}
	return "meaning_choice", nil
}

func (s *Service) mixQuestion(ctx context.Context, tx *sql.Tx, w *word, _ int) error {
	pictures, err := mixedPictures(ctx, tx, w)
	if err != nil {
		return err
	}
	gaps, cloze, err := mixedGap(ctx, tx, w)
	if err != nil {
		return err
	}
	entropy := s.questionEntropy
	if entropy == nil {
		entropy = rand.Reader
	}
	kind, err := chooseMixedKind(entropy, len(pictures) >= 4, len(gaps) == 4)
	if err != nil {
		return err
	}
	if kind == "meaning_choice" {
		return nil
	}
	candidates := gaps
	if kind == "picture_choice" {
		// Keep the target and privately sample three alternatives without replacement.
		for i := len(pictures) - 1; i > 1; i-- {
			n, e := rand.Int(entropy, big.NewInt(int64(i)))
			if e != nil {
				return e
			}
			j := 1 + int(n.Int64())
			pictures[i], pictures[j] = pictures[j], pictures[i]
		}
		candidates = pictures[:4]
	}
	choices := make([]Choice, 0, 4)
	for _, c := range candidates {
		choice := Choice{ID: uuid.NewString(), Text: c.Definition, Image: c.Image}
		if kind == "gap_choice" {
			choice.Text = c.Word
			choice.Image = nil
		}
		if c.ID == w.MeaningID {
			w.AnswerID = choice.ID
		}
		choices = append(choices, choice)
	}
	w.Public.Kind, w.Public.Choices = kind, choices
	if kind == "gap_choice" {
		w.Public.Context = cloze
		w.Public.WordText = ""
		w.Public.WordSlug = ""
		w.Public.PronunciationMeaningID = ""
	}
	return nil
}
