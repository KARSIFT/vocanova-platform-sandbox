-- atlas:txmode file
-- VOC-025-T00: identity persistence foundation. Applied explicitly by migration tooling;
-- the API process does not execute migrations at startup.

CREATE TABLE users (
  id uuid PRIMARY KEY,
  email text,
  display_name text,
  avatar_url text,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled', 'deleted')),
  onboarding_status text NOT NULL DEFAULT 'not_started' CHECK (onboarding_status IN ('not_started', 'in_progress', 'completed')),
  email_verified_at timestamptz,
  last_login_at timestamptz,
  deleted_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX users_active_email_key ON users (lower(email))
  WHERE email IS NOT NULL AND deleted_at IS NULL;

CREATE TABLE external_identities (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  provider text NOT NULL CHECK (provider IN ('google', 'email')),
  provider_subject text NOT NULL CHECK (provider_subject <> ''),
  provider_email text,
  provider_email_verified boolean NOT NULL DEFAULT false,
  deleted_at timestamptz,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX external_identities_provider_subject_key
  ON external_identities (provider, provider_subject) WHERE deleted_at IS NULL;
CREATE INDEX external_identities_user_id_idx ON external_identities (user_id);

CREATE TABLE sessions (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
  created_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  CHECK (expires_at > created_at),
  CHECK (expires_at <= created_at + interval '30 days'),
  CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

CREATE INDEX sessions_user_expiry_idx ON sessions (user_id, expires_at);
CREATE INDEX sessions_active_expiry_idx ON sessions (expires_at) WHERE revoked_at IS NULL;

CREATE TABLE magic_links (
  id uuid PRIMARY KEY,
  user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
  email text NOT NULL CHECK (email <> ''),
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

CREATE INDEX magic_links_active_expiry_idx ON magic_links (expires_at)
  WHERE consumed_at IS NULL AND revoked_at IS NULL;
