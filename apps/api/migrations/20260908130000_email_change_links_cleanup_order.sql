-- atlas:txmode file
-- The bounded email-change credential cleanup claims rows in created_at order.
-- This index lets PostgreSQL stop after one cleanup batch without sorting the
-- full inactive set first. It also covers consumed and revoked links, which
-- the active-expiry partial index intentionally excludes.

CREATE INDEX email_change_links_cleanup_created_at_id_idx
  ON email_change_links (created_at, id);
