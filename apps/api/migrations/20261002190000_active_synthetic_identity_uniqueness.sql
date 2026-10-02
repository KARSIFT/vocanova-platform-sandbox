-- atlas:txmode file
-- Retired synthetic identities retain their provenance and all learning history.
-- Only one non-deleted marked identity may be available for session minting.
DROP INDEX users_single_synthetic_test_account_idx;
CREATE UNIQUE INDEX users_single_synthetic_test_account_idx
  ON users (is_synthetic_test_account)
  WHERE is_synthetic_test_account = true AND deleted_at IS NULL;
