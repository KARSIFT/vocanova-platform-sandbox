-- Self-report and private notes are independent of saved words and review credit.
CREATE TABLE user_word_knowledge (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id uuid NOT NULL REFERENCES users(id),
 meaning_id uuid NOT NULL REFERENCES word_meanings(id),
 self_reported_known boolean NOT NULL DEFAULT false,
 note text NOT NULL DEFAULT '' CHECK (char_length(note) <= 2000),
 updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
 UNIQUE (user_id, meaning_id)
);
-- Fingerprints only: retries must not retain deleted private note text.
CREATE TABLE word_knowledge_actions (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id uuid NOT NULL REFERENCES users(id),
 idempotency_key text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 1 AND 128),
 fingerprint text NOT NULL CHECK (fingerprint ~ '^[0-9a-f]{64}$'),
 created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
 UNIQUE (user_id, idempotency_key)
);
