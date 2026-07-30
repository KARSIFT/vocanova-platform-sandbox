-- atlas:txmode file
-- VOC-028-T00: learner_sentences table for P3 original-sentence practice.
-- Applied explicitly by migration tooling; the API process does not execute
-- migrations at startup.

CREATE TABLE learner_sentences (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  meaning_id uuid REFERENCES word_meanings(id) ON DELETE RESTRICT,
  user_word_id uuid REFERENCES user_words(id) ON DELETE RESTRICT,
  sentence_text text NOT NULL CHECK (sentence_text <> ''),
  normalized_sentence_text text NOT NULL CHECK (normalized_sentence_text <> ''),
  source text NOT NULL CHECK (source IN ('word_detail', 'review', 'daily_mission', 'free_practice')),
  status text NOT NULL DEFAULT 'submitted' CHECK (status IN ('submitted', 'feedback_ready', 'feedback_failed', 'archived')),
  submitted_at timestamptz NOT NULL,
  deleted_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT learner_sentences_sentence_text_length CHECK (char_length(sentence_text) <= 1000)
);

CREATE INDEX learner_sentences_user_id_submitted_at_idx
  ON learner_sentences (user_id, submitted_at);
CREATE INDEX learner_sentences_user_id_status_idx
  ON learner_sentences (user_id, status);
CREATE INDEX learner_sentences_meaning_id_submitted_at_idx
  ON learner_sentences (meaning_id, submitted_at);
CREATE INDEX learner_sentences_user_word_id_submitted_at_idx
  ON learner_sentences (user_word_id, submitted_at);
