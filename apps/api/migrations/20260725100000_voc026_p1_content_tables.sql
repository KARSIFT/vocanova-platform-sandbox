-- atlas:txmode transaction
-- VOC-026-T00: P1 canonical content and user_words persistence. Applied explicitly
-- by migration tooling; the API process does not execute migrations at startup.

CREATE TABLE canonical_words (
  id uuid PRIMARY KEY,
  text text NOT NULL CHECK (text <> ''),
  normalized_text text NOT NULL CHECK (normalized_text <> ''),
  word_type text NOT NULL DEFAULT 'word' CHECK (word_type IN ('word', 'phrase', 'phrasal_verb', 'idiom', 'collocation')),
  language_code text NOT NULL DEFAULT 'en',
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
  difficulty_level text CHECK (difficulty_level IN ('a1', 'a2', 'b1', 'b2', 'c1', 'unknown')),
  frequency_rank integer,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX canonical_words_language_normalized_text_key
  ON canonical_words (language_code, normalized_text);

CREATE TABLE word_meanings (
  id uuid PRIMARY KEY,
  word_id uuid NOT NULL REFERENCES canonical_words(id) ON DELETE RESTRICT,
  part_of_speech text NOT NULL CHECK (part_of_speech IN ('noun', 'verb', 'adjective', 'adverb', 'preposition', 'conjunction', 'interjection', 'pronoun', 'determiner', 'phrase', 'idiom', 'phrasal_verb', 'collocation', 'other')),
  short_definition text NOT NULL CHECK (short_definition <> ''),
  learner_definition text,
  meaning_order integer NOT NULL,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
  difficulty_level text CHECK (difficulty_level IN ('a1', 'a2', 'b1', 'b2', 'c1', 'unknown')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX word_meanings_word_id_meaning_order_key
  ON word_meanings (word_id, meaning_order);
CREATE INDEX word_meanings_word_id_idx ON word_meanings (word_id);

CREATE TABLE word_examples (
  id uuid PRIMARY KEY,
  meaning_id uuid NOT NULL REFERENCES word_meanings(id) ON DELETE RESTRICT,
  example_text text NOT NULL CHECK (example_text <> ''),
  example_order integer NOT NULL,
  difficulty_level text CHECK (difficulty_level IN ('a1', 'a2', 'b1', 'b2', 'c1', 'unknown')),
  situation_label text,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX word_examples_meaning_id_example_order_key
  ON word_examples (meaning_id, example_order);
CREATE INDEX word_examples_meaning_id_idx ON word_examples (meaning_id);

CREATE TABLE usage_notes (
  id uuid PRIMARY KEY,
  meaning_id uuid NOT NULL REFERENCES word_meanings(id) ON DELETE RESTRICT,
  note_type text NOT NULL CHECK (note_type IN ('collocation', 'register', 'common_mistake', 'grammar', 'pronunciation', 'other')),
  note_text text NOT NULL CHECK (note_text <> ''),
  note_order integer NOT NULL,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX usage_notes_meaning_id_note_order_key
  ON usage_notes (meaning_id, note_order);
CREATE INDEX usage_notes_meaning_id_idx ON usage_notes (meaning_id);

CREATE TABLE journey_situations (
  id uuid PRIMARY KEY,
  slug text NOT NULL CHECK (slug <> ''),
  title text NOT NULL CHECK (title <> ''),
  short_description text NOT NULL CHECK (short_description <> ''),
  level_band text CHECK (level_band IN ('a1_a2', 'a2_b1', 'b1_b2', 'mixed')),
  category text NOT NULL CHECK (category IN ('daily_life', 'travel', 'work', 'study', 'social')),
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
  display_order integer NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX journey_situations_slug_key ON journey_situations (slug);
CREATE INDEX journey_situations_status_display_order_idx
  ON journey_situations (status, display_order);

CREATE TABLE journey_words (
  id uuid PRIMARY KEY,
  journey_situation_id uuid NOT NULL REFERENCES journey_situations(id) ON DELETE RESTRICT,
  meaning_id uuid NOT NULL REFERENCES word_meanings(id) ON DELETE RESTRICT,
  relevance_score integer NOT NULL DEFAULT 50 CHECK (relevance_score >= 1 AND relevance_score <= 100),
  display_order integer,
  is_core boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX journey_words_situation_meaning_key
  ON journey_words (journey_situation_id, meaning_id);
CREATE INDEX journey_words_situation_id_idx ON journey_words (journey_situation_id);
CREATE INDEX journey_words_meaning_id_idx ON journey_words (meaning_id);

CREATE TABLE user_words (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  meaning_id uuid NOT NULL REFERENCES word_meanings(id) ON DELETE RESTRICT,
  status text NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'learning', 'reviewing', 'mastered', 'ignored', 'archived')),
  source text NOT NULL CHECK (source IN ('journey', 'search', 'ai_suggestion', 'manual', 'seed')),
  review_step integer NOT NULL DEFAULT 0 CHECK (review_step >= 0 AND review_step <= 7),
  next_review_at timestamptz,
  last_reviewed_at timestamptz,
  last_result text CHECK (last_result IN ('correct', 'incorrect', 'skipped')),
  last_rating text CHECK (last_rating IN ('again', 'hard', 'good', 'easy')),
  consecutive_correct_count integer NOT NULL DEFAULT 0,
  consecutive_incorrect_count integer NOT NULL DEFAULT 0,
  total_review_count integer NOT NULL DEFAULT 0,
  correct_review_count integer NOT NULL DEFAULT 0 CHECK (correct_review_count <= total_review_count),
  added_at timestamptz NOT NULL,
  mastered_at timestamptz,
  ignored_at timestamptz,
  deleted_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX user_words_user_id_meaning_id_active_key
  ON user_words (user_id, meaning_id) WHERE deleted_at IS NULL;
CREATE INDEX user_words_user_id_status_idx ON user_words (user_id, status);
CREATE INDEX user_words_user_id_next_review_at_idx ON user_words (user_id, next_review_at) WHERE deleted_at IS NULL;
