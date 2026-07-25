-- atlas:txmode transaction
-- VOC-025-T02: OAuth state persistence. Applied explicitly by migration tooling;
-- the API process does not execute migrations at startup.

CREATE TABLE oauth_states (
  id uuid PRIMARY KEY,
  token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
  environment text NOT NULL CHECK (environment <> ''),
  provider text NOT NULL CHECK (provider IN ('google', 'email')),
  app_return_url text NOT NULL CHECK (app_return_url <> ''),
  created_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  consumed_at timestamptz,
  CHECK (expires_at > created_at),
  CHECK (expires_at <= created_at + interval '10 minutes')
);

CREATE INDEX oauth_states_active_expiry_idx ON oauth_states (expires_at)
  WHERE consumed_at IS NULL;
