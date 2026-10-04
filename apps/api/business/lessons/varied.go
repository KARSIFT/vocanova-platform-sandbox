package lessons

const ExerciseVersion = "mixed-recall-v1"

// buildVariedSnapshot is used for newly created sessions only. buildSnapshot
// remains the original builder so historical lesson definitions and snapshots
// retain their previous prompts, choice IDs and action fingerprints.
func buildVariedSnapshot(d Definition, words []Word, seed string) (Snapshot, error) {
	snap, err := buildSnapshot(d, words, seed)
	if err != nil {
		return Snapshot{}, err
	}
	snap.ExerciseVersion = ExerciseVersion
	typed := &snap.Steps[4]
	typed.Step.ID = "typed_recall-2"
	typed.Step.Kind = "typed_recall"
	typed.Step.Prompt = d.Words[1].Prompt
	typed.Step.Context = d.Words[1].Context
	typed.Step.Choices = []Choice{}
	typed.CorrectChoiceID = ""
	typed.Explanations = map[string]string{}
	typed.Accepted = []string{words[1].WordText}
	listening := &snap.Steps[5]
	listening.Step.ID = "listening_choice-3"
	listening.Step.Kind = "listening_choice"
	listening.Step.Prompt = "Listen with device pronunciation. Which meaning matches?"
	listening.Step.SpeechText = words[2].WordText
	listening.Step.SpeechLanguage = "en-US"
	return snap, nil
}
