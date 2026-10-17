-- atlas:txmode file
-- Preserve in-flight OAuth flows and older clients while binding new session
-- duration choices to server-side, single-use OAuth state. No bearer material.
ALTER TABLE oauth_states
  ADD COLUMN session_persistent boolean NOT NULL DEFAULT true;
