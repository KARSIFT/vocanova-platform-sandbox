package aifeedback

import (
	"encoding/json"
	"regexp"
	"strings"
)

// These are narrow output heuristics, not proof of prompt leakage or injection
// resistance. In particular, an echoed grading command is inappropriate tutor
// feedback even when the command originated in the learner's own sentence.
// Ordinary vocabulary such as "instructions" and "you are a" is not a signal.
var feedbackInstructionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(?:the|my|our)\s+(?:system|developer)\s+(?:prompt|instructions?)\s+(?:says?|said|tells?|told|requires?|instructs?|asks?|directs?)\b`),
	regexp.MustCompile(`(?i)\b(?:system|developer)\s+prompt\s*:`),
	regexp.MustCompile(`(?i)\b(?:reveal|expose|print|show|repeat)\s+(?:(?:the|my|your)\s+)?(?:hidden\s+)?(?:system|developer)\s+(?:prompt|instructions?)\b`),
	regexp.MustCompile(`(?i)\b(?:ignore|disregard|override)\s+(?:(?:all|the|any)\s+)?(?:previous|prior|above)\s+(?:instructions?|prompts?|rules?)\b`),
	regexp.MustCompile(`(?i)(?:^|[.!?;]\s*)(?:please\s+)?(?:mark|grade|classify|label)\s+(?:this|it|this sentence|my sentence|the sentence)\s+(?:as\s+)?(?:correct|incorrect|needs_improvement)\b`),
	// Distinctive excerpts of the trusted tutor instructions, not generic roles.
	regexp.MustCompile(`(?i)\byou are a concise,\s*supportive english-learning tutor\b`),
	regexp.MustCompile(`(?i)\byour only job is to evaluate whether the learner['’]s sentence\b`),
	regexp.MustCompile(`(?i)\balways return a single valid json object matching the provided schema\b`),
}

func containsLeakedInstructions(feedback *ProviderFeedback) bool {
	fields := []string{feedback.Headline, feedback.Explanation}
	if feedback.CorrectedSentence != nil {
		fields = append(fields, *feedback.CorrectedSentence)
	}
	if feedback.ImprovementTip != nil {
		fields = append(fields, *feedback.ImprovementTip)
	}
	for _, field := range fields {
		// Whitespace normalization is local to each field: joining fields can
		// manufacture a phrase that the provider never actually returned.
		text := strings.Join(strings.Fields(field), " ")
		for _, pattern := range feedbackInstructionPatterns {
			if pattern.MatchString(text) {
				return true
			}
		}
	}

	// Preserve the existing raw-output validity guard independently of the
	// text patterns. This function's legacy boolean does not classify a cause.
	if feedback.RawJSON == nil {
		return true
	}
	if _, err := json.Marshal(feedback.RawJSON); err != nil {
		return true
	}
	return false
}
