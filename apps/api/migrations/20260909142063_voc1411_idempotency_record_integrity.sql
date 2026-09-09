-- atlas:txmode file
-- Issue #1411: idempotency records are authenticated-user-scoped request
-- claims. Bind each claim to its owning user and require an actual request
-- fingerprint. NOT VALID preserves rollout over possible legacy corruption
-- while both constraints protect every new row. PostgreSQL also enforces the
-- check constraint on every update and the foreign key when user_id changes;
-- an unrelated update to a pre-existing orphan remains possible until a
-- separate legacy cleanup and VALIDATE CONSTRAINT rollout.

ALTER TABLE idempotency_keys
  ADD CONSTRAINT idempotency_keys_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
  NOT VALID;

ALTER TABLE idempotency_keys
  ADD CONSTRAINT idempotency_keys_fingerprint_nonempty
  CHECK (char_length(fingerprint) > 0)
  NOT VALID;
