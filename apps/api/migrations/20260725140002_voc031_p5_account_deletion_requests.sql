-- atlas:txmode transaction
-- VOC-031-T04: account_deletion_requests (DOC-05 §16, DOC-06 §14,
-- VOC-031-D07). One row per account that has been deactivated
-- pending purge. The status field is the only state the sweep
-- function reads: 'deactivated' rows past their purge_after are
-- processed; 'completed' rows are left alone.
--
-- The user_id UNIQUE constraint enforces "at most one open
-- deletion request per user" at the DB layer — a learner cannot
-- hold two parallel account-deletion rows in flight even if a
-- replayed request reaches the transaction before the first
-- idempotency key is recorded.
--
-- The 30-day default purge_after is the DOC-05 §16 baseline. The
-- sweep never uses a fixed wall-clock value: it reads each row's
-- own purge_after so the legal-review process can override the
-- per-account delay on a case-by-case basis without a schema
-- change.
--
-- The created_at / updated_at pair is the standard lifecycle
-- pair. completed_at is set on the sweep transition so the
-- audit log can answer "when was this user fully purged" without
-- joining another table.
--
-- No existing A1–P4 table, column, or constraint is altered.
-- No ON DELETE CASCADE is introduced: the per-table disposition
-- runs in code, not as a database cascade (DOC-05 §16).

CREATE TABLE account_deletion_requests (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
  status text NOT NULL DEFAULT 'deactivated'
    CHECK (status IN ('deactivated', 'anonymizing', 'completed')),
  requested_at timestamptz NOT NULL,
  purge_after timestamptz NOT NULL,
  completed_at timestamptz,
  idempotency_key text NOT NULL CHECK (idempotency_key <> ''),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CHECK (purge_after > requested_at),
  CHECK (purge_after <= requested_at + interval '365 days'),
  CHECK (status <> 'completed' OR completed_at IS NOT NULL),
  CHECK (status = 'completed' OR completed_at IS NULL)
);

CREATE INDEX account_deletion_requests_status_purge_after_idx
  ON account_deletion_requests (status, purge_after)
  WHERE status = 'deactivated';

CREATE INDEX account_deletion_requests_user_id_idx
  ON account_deletion_requests (user_id);
