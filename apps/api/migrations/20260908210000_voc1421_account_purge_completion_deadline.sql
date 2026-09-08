-- atlas:txmode file
-- Issue #1421: a staged account purge cannot complete before its per-request
-- legal-review deadline. NOT VALID preserves any legacy early-completion rows
-- while protecting every new or updated deletion request.

ALTER TABLE account_deletion_requests
  ADD CONSTRAINT account_deletion_requests_completed_after_purge_due
  CHECK (status <> 'completed' OR completed_at >= purge_after)
  NOT VALID;
