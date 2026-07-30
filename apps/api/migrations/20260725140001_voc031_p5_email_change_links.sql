-- atlas:txmode file
-- VOC-031-T03: email_change_links (DOC-06 §6, VOC-031-D05). Mirrors the
-- magic_links table almost exactly, with three deliberate differences
-- required because this is not a login mechanism:
--
--   1. user_id is NOT NULL: requesting an email change requires an
--      already-authenticated session, so every row must belong to a
--      known learner. magic_links' user_id is nullable because
--      pre-consume links are created before any user exists.
--   2. new_email replaces email: the destination address the
--      requester wants to switch to, not the current sign-in address.
--   3. octet_length(token_hash) is still 32 bytes (SHA-256 of a 32-
--      random-byte input), and 15-minute expiry, single-use, and
--      environment-scoping all match magic_links so a future audit
--      can apply the same secret-handling discipline to both.
--
-- No existing A1-P4 table, column, or constraint is altered. The
-- new_email column is the destination; uniqueness of users.email is
-- still enforced by users_active_email_key (the partial unique index
-- on lower(email) WHERE deleted_at IS NULL), and that index is the
-- authoritative guard against duplicate-email confirmation races
-- (VOC-031-R02). email_change_links itself carries no email
-- uniqueness constraint because the same new_email can be requested
-- multiple times by different learners (only one will be allowed to
-- confirm it; the others receive a stable conflict response).

CREATE TABLE email_change_links (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  new_email text NOT NULL CHECK (new_email <> ''),
  token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
  environment text NOT NULL CHECK (environment <> ''),
  created_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  consumed_at timestamptz,
  revoked_at timestamptz,
  CHECK (expires_at > created_at),
  CHECK (expires_at <= created_at + interval '15 minutes'),
  CHECK (consumed_at IS NULL OR consumed_at >= created_at),
  CHECK (revoked_at IS NULL OR revoked_at >= created_at),
  CHECK (consumed_at IS NULL OR revoked_at IS NULL)
);

CREATE INDEX email_change_links_active_expiry_idx
  ON email_change_links (expires_at)
  WHERE consumed_at IS NULL AND revoked_at IS NULL;

CREATE INDEX email_change_links_user_id_idx
  ON email_change_links (user_id);
