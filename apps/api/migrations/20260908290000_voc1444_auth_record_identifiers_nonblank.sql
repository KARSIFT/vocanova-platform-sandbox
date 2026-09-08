-- atlas:txmode file
-- Issue #1444: persisted identity and token-routing identifiers must contain
-- at least one non-whitespace character. The explicit Unicode White_Space set
-- avoids locale-dependent POSIX character classes. NOT VALID retains legacy
-- records while protecting all new inserts and updates.

ALTER TABLE external_identities
  ADD CONSTRAINT external_identities_provider_subject_nonblank
    CHECK (provider_subject ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]')
    NOT VALID;

ALTER TABLE magic_links
  ADD CONSTRAINT magic_links_email_nonblank
    CHECK (email ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]')
    NOT VALID,
  ADD CONSTRAINT magic_links_environment_nonblank
    CHECK (environment ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]')
    NOT VALID;

ALTER TABLE email_change_links
  ADD CONSTRAINT email_change_links_new_email_nonblank
    CHECK (new_email ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]')
    NOT VALID,
  ADD CONSTRAINT email_change_links_environment_nonblank
    CHECK (environment ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]')
    NOT VALID;
