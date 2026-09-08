-- atlas:txmode file
-- Issue #1415: terminal credential timestamps must describe a possible
-- lifecycle. NOT VALID preserves any legacy contradictions while protecting
-- every new and updated session, magic link, OAuth state, and email-change
-- link. Revocation may occur after expiry, but consumption must occur while
-- the credential is valid (strictly before expires_at).

ALTER TABLE sessions
  ADD CONSTRAINT sessions_revoked_at_not_before_created
  CHECK (revoked_at IS NULL OR revoked_at >= created_at)
  NOT VALID;

ALTER TABLE magic_links
  ADD CONSTRAINT magic_links_consumed_at_within_lifetime
  CHECK (consumed_at IS NULL OR (consumed_at >= created_at AND consumed_at < expires_at))
  NOT VALID;

ALTER TABLE magic_links
  ADD CONSTRAINT magic_links_revoked_at_not_before_created
  CHECK (revoked_at IS NULL OR revoked_at >= created_at)
  NOT VALID;

ALTER TABLE oauth_states
  ADD CONSTRAINT oauth_states_consumed_at_within_lifetime
  CHECK (consumed_at IS NULL OR (consumed_at >= created_at AND consumed_at < expires_at))
  NOT VALID;

ALTER TABLE email_change_links
  ADD CONSTRAINT email_change_links_consumed_at_within_lifetime
  CHECK (consumed_at IS NULL OR (consumed_at >= created_at AND consumed_at < expires_at))
  NOT VALID;

ALTER TABLE email_change_links
  ADD CONSTRAINT email_change_links_revoked_at_not_before_created
  CHECK (revoked_at IS NULL OR revoked_at >= created_at)
  NOT VALID;
