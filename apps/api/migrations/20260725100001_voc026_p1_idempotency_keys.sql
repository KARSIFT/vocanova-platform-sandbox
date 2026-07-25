-- atlas:txmode transaction
-- VOC-026-T02: Idempotency key persistence for user-words save. Applied
-- explicitly by migration tooling; the API process does not execute migrations
-- at startup.

CREATE TABLE idempotency_keys (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL,
  operation text NOT NULL CHECK (operation <> ''),
  key text NOT NULL CHECK (key <> ''),
  fingerprint text NOT NULL,
  created_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX idempotency_keys_user_id_operation_key_key
  ON idempotency_keys (user_id, operation, key);
CREATE INDEX idempotency_keys_created_at_idx
  ON idempotency_keys (created_at);
