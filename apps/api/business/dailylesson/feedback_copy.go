package dailylesson

// AssistedReason marks feedback for an "I'm not sure" request. The learner saw
// the meaning with help: it is scheduled like a miss, never graded correct, and
// earns no daily-goal effort, success credit or rating points.
const AssistedReason = "assisted"

// vettedSentence restores the exact authored sentence for a frozen gap question.
// It is returned only while the frozen cloze still equals the pinned example,
// so feedback never pairs a question with different content.
func vettedSentence(w *word) (string, bool) {
	if w.Public.Kind != "gap_choice" {
		return "", false
	}
	gap, ok := gapSets[w.MeaningID]
	if !ok {
		return "", false
	}
	cloze, ok := mixedCloze(gap.Example, w.WordText)
	if !ok || cloze != w.Public.Context {
		return "", false
	}
	return gap.Example, true
}

// feedbackExplanation is plain-English feedback for one graded or assisted
// daily answer. It depends only on the frozen question and its outcome, so a
// reload or replayed receipt always shows the same words. Distractors in every
// daily mode come from different canonical words, which the copy relies on.
func feedbackExplanation(w *word, correct, assisted bool) string {
	sentence, hasSentence := vettedSentence(w)
	full := ""
	if hasSentence {
		full = " Read the full sentence. “" + sentence + "”"
	}
	if assisted {
		if hasSentence {
			return "Here is the word that fits." + full + " This counts as practice with help, not a remembered answer. The word will return on another day so you can try it on your own."
		}
		return "Here is what this word means. This counts as practice with help, not a remembered answer. The word will return on another day so you can try it on your own."
	}
	if correct {
		switch w.Public.Kind {
		case "gap_choice":
			if hasSentence {
				return "Yes, that word fits." + full + " You’ve reviewed this word for today."
			}
			return "Yes, that word fits the sentence. You’ve reviewed this word for today."
		case "picture_choice":
			return "Yes, that choice matches this meaning. You’ve reviewed this word for today."
		default:
			return "Yes, that is what this word means. You’ve reviewed this word for today."
		}
	}
	switch w.Public.Kind {
	case "gap_choice":
		if hasSentence {
			return "That word doesn’t fit here." + full + " Read it once more. The word will return on another day."
		}
		return "That word doesn’t fit this sentence. Read the meaning once more. The word will return on another day."
	case "picture_choice":
		return "That choice shows a different meaning. Read the meaning once more. The word will return on another day."
	default:
		return "That meaning belongs to a different word. Read the right meaning once more. The word will return on another day."
	}
}
