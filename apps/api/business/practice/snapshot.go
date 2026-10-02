package practice

import (
	"crypto/sha256"
	"github.com/google/uuid"
	"sort"
	"strings"
)

func order(seed, s string) string { h := sha256.Sum256([]byte(seed + ":" + s)); return string(h[:]) }
func build(req StartRequest, words []Word, sources map[string]Source, seed string) (Snapshot, error) {
	snap := Snapshot{Mode: req.Mode, LessonKey: req.LessonKey, ContentVersion: ContentVersion, GradingVersion: GradingVersion, Steps: []privateStep{}}
	if len(words) < 3 {
		return snap, ErrContentUnavailable
	}
	selected := []Word{}
	for _, w := range words {
		if req.LessonKey != "" && w.LessonKey != req.LessonKey {
			continue
		}
		if req.Mode == "mistakes" {
			if _, ok := sources[w.MeaningID]; !ok {
				continue
			}
		}
		selected = append(selected, w)
	}
	sort.Slice(selected, func(i, j int) bool { return order(seed, selected[i].MeaningID) < order(seed, selected[j].MeaningID) })
	if len(selected) > 6 {
		selected = selected[:6]
	}
	if len(selected) == 0 {
		if req.Mode == "mistakes" {
			return snap, ErrNoMistakes
		}
		return snap, ErrContentUnavailable
	}
	for _, w := range selected {
		id := uuid.NewString()
		step := privateStep{Word: w, Accepted: acceptedAnswers(w.WordText), Public: Step{ID: id, Kind: "typed_recall", Prompt: "Type the word or phrase you learned: " + recallDefinition(w), Choices: []Choice{}}}
		if src, ok := sources[w.MeaningID]; ok && req.Mode == "mistakes" {
			copy := src
			step.Source = &copy
		}
		if req.Mode == "listening_choice" {
			step.Public.Kind = "listening_choice"
			step.Public.Prompt = "Listen with device pronunciation. Which meaning matches?"
			step.Public.SpeechText = w.WordText
			if w.WordText == "resume" {
				step.Public.SpeechText = "résumé"
			}
			step.Public.SpeechLanguage = "en-US"
			options := []Word{}
			for _, o := range words {
				if o.LessonKey == w.LessonKey {
					options = append(options, o)
				}
			}
			if len(options) != 3 {
				return snap, ErrContentUnavailable
			}
			for _, o := range options {
				cid := uuid.NewString()
				step.Public.Choices = append(step.Public.Choices, Choice{ID: cid, Text: o.Definition})
				if o.MeaningID == w.MeaningID {
					step.CorrectChoice = cid
				}
			}
			sort.Slice(step.Public.Choices, func(i, j int) bool {
				return order(seed+id, step.Public.Choices[i].ID) < order(seed+id, step.Public.Choices[j].ID)
			})
		}
		// A typed definition containing its literal answer must not become a recall prompt.
		if step.Public.Kind == "typed_recall" && strings.Contains(" "+normalize(recallDefinition(w))+" ", " "+normalize(w.WordText)+" ") {
			return snap, ErrContentUnavailable
		}
		snap.Steps = append(snap.Steps, step)
	}
	return snap, nil
}
