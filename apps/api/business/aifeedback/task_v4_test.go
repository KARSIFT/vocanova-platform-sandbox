package aifeedback

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFeedbackV4SharedFrozenRubricAndMetadata(t *testing.T) {
	// Freeze the approved case-free rubric, including its final newline. A new
	// editorial revision needs a deliberate version and evidence decision.
	require.Equal(t, "85c24c9e0971ad80b9301d4ba87baa34b185fb2bf4636f849d60d9ab26625e06", fmt.Sprintf("%x", sha256.Sum256([]byte(feedbackRubric()))))
	initial := NewDefaultTaskBuilder().Build(&Target{NormalizedWord: "invite", PartOfSpeech: "verb", ShortDefinition: "ask someone to an event", LearnerLevel: "A2"}, "private-learner-marker")
	repair := NewDefaultTaskBuilder().BuildRepair(initial, "private-validation-marker", map[string]any{"headline": "private-output-marker"})
	for _, task := range []ProviderTask{initial, repair} {
		require.Equal(t, "sentence-feedback-v4", task.PromptVersion)
		require.Equal(t, "feedback-schema-v3", task.SchemaVersion)
		require.Equal(t, 300, task.MaxOutputTokens)
		require.Equal(t, systemPrompt(), task.SystemPrompt)
		require.Equal(t, outputSchema(), task.OutputSchema)
		require.True(t, strings.HasSuffix(task.DeveloperPrompt, feedbackRubric()))
		require.NotContains(t, task.DeveloperPrompt, "private-")
		require.Equal(t, "private-learner-marker", task.UserPayload["learner_sentence"])
	}
	require.Equal(t, "Evaluate the learner sentence using this rubric. "+feedbackRubric(), initial.DeveloperPrompt)
	require.NotContains(t, initial.UserPayload, "repair_attempt")
	require.Equal(t, true, repair.UserPayload["repair_attempt"])
	require.Equal(t, "private-validation-marker", repair.UserPayload["validation_error"])
}
