-- atlas:txmode file
-- Password credentials are deliberately separate from users and external
-- identities. Pending registration records never create an account, and every
-- bearer token is persisted only as a SHA-256 hash.

CREATE TABLE password_credentials (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
  password_hash text NOT NULL CHECK (octet_length(password_hash) BETWEEN 32 AND 512),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE password_registration_links (
  id uuid PRIMARY KEY,
  email text NOT NULL CHECK (email <> ''),
  display_name text,
  password_hash text NOT NULL CHECK (octet_length(password_hash) BETWEEN 32 AND 512),
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
CREATE INDEX password_registration_links_active_expiry_idx
  ON password_registration_links (expires_at)
  WHERE consumed_at IS NULL AND revoked_at IS NULL;
CREATE INDEX password_registration_links_email_idx
  ON password_registration_links (lower(email));

CREATE TABLE password_reset_links (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  email_at_issue text NOT NULL CHECK (email_at_issue <> ''),
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
CREATE INDEX password_reset_links_active_expiry_idx
  ON password_reset_links (expires_at)
  WHERE consumed_at IS NULL AND revoked_at IS NULL;
CREATE INDEX password_reset_links_user_id_idx ON password_reset_links (user_id);
