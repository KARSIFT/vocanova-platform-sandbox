-- Private named collections. Tombstones prevent accidental resurrection of deleted IDs.
CREATE TABLE user_word_lists (
 id uuid PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
 name text NOT NULL CHECK (length(btrim(name)) BETWEEN 1 AND 80),
 revision bigint NOT NULL CHECK (revision > 0),
 created_at timestamptz NOT NULL,
 updated_at timestamptz NOT NULL,
 deleted_at timestamptz
);
CREATE INDEX user_word_lists_owner ON user_word_lists(user_id);
CREATE TABLE user_word_list_members (
 list_id uuid NOT NULL REFERENCES user_word_lists(id) ON DELETE CASCADE,
 meaning_id uuid NOT NULL REFERENCES word_meanings(id) ON DELETE RESTRICT,
 created_at timestamptz NOT NULL,
 PRIMARY KEY(list_id,meaning_id)
);
CREATE TABLE word_list_actions (
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
 idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 1 AND 128),
 fingerprint text NOT NULL,
 list_id uuid NOT NULL REFERENCES user_word_lists(id) ON DELETE CASCADE,
 operation text NOT NULL CHECK (operation IN ('put','delete','add','remove')),
 created_at timestamptz NOT NULL,
 PRIMARY KEY(user_id,idempotency_key)
);
