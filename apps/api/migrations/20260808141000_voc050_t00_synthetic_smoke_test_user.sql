-- atlas:txmode file
-- VOC-050-T00: add a durable synthetic-account marker on users.
-- This supports an idempotent deploy-time seed that guarantees the
-- smoke-test account exists and is distinguishable from real users.
--
-- The marker is explicitly additive: no existing auth/session behavior
-- changes for real users because the default is false.
ALTER TABLE users
  ADD COLUMN is_synthetic_test_account boolean NOT NULL DEFAULT false;

-- At most one marked synthetic smoke-test account may exist at once.
-- This keeps the deploy-time seed deterministic and avoids accidental
-- creation of multiple privileged synthetic identities.
CREATE UNIQUE INDEX users_single_synthetic_test_account_idx
  ON users (is_synthetic_test_account)
  WHERE is_synthetic_test_account = true;
