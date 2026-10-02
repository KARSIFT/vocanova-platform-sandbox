-- Synthetic restore rehearsal fixture v1. Never load into an application database.
-- Apply all forward migrations and the canonical seed first. No provider is called.
-- Explicit inserts intentionally reject repeat loading or partially populated fixtures.
\set ON_ERROR_STOP on
BEGIN;
SET LOCAL search_path = public, pg_catalog;
SET LOCAL TIME ZONE 'UTC';
DO $guard$
BEGIN
  IF current_database() <> 'vocanova_rehearsal' THEN
    RAISE EXCEPTION 'rehearsal_fixture_wrong_database';
  END IF;
  IF EXISTS (SELECT 1 FROM public.users) THEN
    RAISE EXCEPTION 'rehearsal_fixture_requires_empty_users';
  END IF;
END;
$guard$;

-- users: 1 deterministic synthetic record(s).
INSERT INTO public.users (
  id, email, display_name, avatar_url, status, onboarding_status, email_verified_at, last_login_at, deleted_at, created_at, updated_at, is_synthetic_test_account
) VALUES (
  'd0c0a001-0000-4000-8000-000000000001', 'restore-rehearsal@vocanova.invalid', 'Synthetic restore learner', NULL, 'active', 'completed', NULL, NULL, NULL, '2026-10-02 10:00:00+00', '2026-10-02 10:00:00+00', true
);

-- user_settings: 1 deterministic synthetic record(s).
INSERT INTO public.user_settings (
  id, user_id, timezone, daily_review_target, review_interval_preset, notifications_enabled, marketing_emails_enabled, app_language, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000002', 'd0c0a001-0000-4000-8000-000000000001', 'UTC', 5, 'vocanova_default', false, false, 'en', '2026-10-02 10:00:00+00', '2026-10-02 10:00:00+00'
);

-- user_onboarding_profiles: 1 deterministic synthetic record(s).
INSERT INTO public.user_onboarding_profiles (
  id, user_id, english_level, native_language, learning_goal, main_use_case, daily_review_target, completed_at, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000014', 'd0c0a001-0000-4000-8000-000000000001', 'a2', 'es', 'conversation', 'social', 5, '2026-10-02 10:00:00+00', '2026-10-02 10:00:00+00', '2026-10-02 10:00:00+00'
);

-- user_words: 1 deterministic synthetic record(s).
INSERT INTO public.user_words (
  id, user_id, meaning_id, status, source, review_step, next_review_at, last_reviewed_at, last_result, last_rating, consecutive_correct_count, consecutive_incorrect_count, total_review_count, correct_review_count, added_at, mastered_at, ignored_at, deleted_at, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000003', 'd0c0a001-0000-4000-8000-000000000001', '3d64c3c9-ede0-5ffd-b1ef-278f6b70e486', 'learning', 'journey', 1, '2026-10-02 11:05:00+00', '2026-10-02 10:05:00+00', 'correct', 'good', 1, 0, 1, 1, '2026-10-02 10:00:00+00', NULL, NULL, NULL, '2026-10-02 10:00:00+00', '2026-10-02 10:05:00+00'
);

-- review_attempts: 1 deterministic synthetic record(s).
INSERT INTO public.review_attempts (
  id, user_id, user_word_id, meaning_id, attempt_type, prompt_type, result, rating, review_step_before, review_step_after, answered_at, response_time_ms, selected_option_meaning_id, typed_answer, was_hint_used, source, client_attempt_id, metadata, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000004', 'd0c0a001-0000-4000-8000-000000000001', 'd0c0a001-0000-4000-8000-000000000003', '3d64c3c9-ede0-5ffd-b1ef-278f6b70e486', 'review', 'multiple_choice', 'correct', 'good', 0, 1, '2026-10-02 10:05:00+00', 1200, '3d64c3c9-ede0-5ffd-b1ef-278f6b70e486', NULL, false, 'review_session', 'restore-rehearsal-v1-review', '{"fixture":"restore-rehearsal-v1"}'::jsonb, '2026-10-02 10:05:00+00', '2026-10-02 10:05:00+00'
);

-- learner_sentences: 1 deterministic synthetic record(s).
INSERT INTO public.learner_sentences (
  id, user_id, meaning_id, user_word_id, sentence_text, normalized_sentence_text, source, status, submitted_at, deleted_at, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000005', 'd0c0a001-0000-4000-8000-000000000001', '3d64c3c9-ede0-5ffd-b1ef-278f6b70e486', 'd0c0a001-0000-4000-8000-000000000003', 'I caught up with Maya and heard about her new job.', 'i caught up with maya and heard about her new job.', 'word_detail', 'feedback_ready', '2026-10-02 10:06:00+00', NULL, '2026-10-02 10:06:00+00', '2026-10-02 10:06:01+00'
);

-- ai_feedback_attempts: 1 deterministic synthetic record(s).
INSERT INTO public.ai_feedback_attempts (
  id, learner_sentence_id, status, provider, model, prompt_version, request_hash, feedback_json, feedback_text, error_code, error_message, started_at, completed_at, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000006', 'd0c0a001-0000-4000-8000-000000000005', 'succeeded', 'synthetic_rehearsal', 'fixed-fixture-no-provider-call', 'restore-rehearsal-v1', 'restore-rehearsal-v1-sentence-1', '{"status":"correct","target_word_used_correctly":true,"grammar_acceptable":true,"meaning_clear":true,"naturalness":"natural","headline":"Good use of catch up.","explanation":"Caught up clearly describes sharing news after time apart."}'::jsonb, 'Caught up clearly describes sharing news after time apart.', NULL, NULL, '2026-10-02 10:06:00+00', '2026-10-02 10:06:01+00', '2026-10-02 10:06:00+00', '2026-10-02 10:06:01+00'
);

-- daily_mission_snapshots: 1 deterministic synthetic record(s).
INSERT INTO public.daily_mission_snapshots (
  id, user_id, local_date, timezone, review_target, reviews_completed, new_word_target, new_words_completed, sentence_practice_target, sentence_practices_completed, policy_version, status, completed_at, grace_applied, grace_day_id, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000007', 'd0c0a001-0000-4000-8000-000000000001', '2026-10-02', 'UTC', 5, 1, 1, 1, 1, 1, 'p4-mission-policy-v1', 'open', NULL, false, NULL, '2026-10-02 10:00:00+00', '2026-10-02 10:06:01+00'
);

-- daily_activity_summaries: 1 deterministic synthetic record(s).
INSERT INTO public.daily_activity_summaries (
  id, user_id, local_date, timezone, reviews_attempted, reviews_correct, reviews_skipped, words_discovered, words_added, sentences_submitted, ai_feedback_received, confidence_points_earned, confidence_points_spent, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000008', 'd0c0a001-0000-4000-8000-000000000001', '2026-10-02', 'UTC', 1, 1, 0, 1, 1, 1, 1, 12, 0, '2026-10-02 10:00:00+00', '2026-10-02 10:06:01+00'
);

-- streak_states: 1 deterministic synthetic record(s).
INSERT INTO public.streak_states (
  id, user_id, current_streak_count, longest_streak_count, last_completed_local_date, last_activity_local_date, timezone, status, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000009', 'd0c0a001-0000-4000-8000-000000000001', 0, 0, NULL, '2026-10-02', 'UTC', 'active', '2026-10-02 10:00:00+00', '2026-10-02 10:06:01+00'
);

-- confidence_point_ledger: 4 deterministic synthetic record(s).
INSERT INTO public.confidence_point_ledger (
  id, user_id, amount, balance_after, reason, source_type, source_id, idempotency_key, metadata, occurred_at, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000010', 'd0c0a001-0000-4000-8000-000000000001', 2, 2, 'word_added', 'user_word', 'd0c0a001-0000-4000-8000-000000000003', 'restore-rehearsal-v1-word_added', '{"fixture":"restore-rehearsal-v1"}'::jsonb, '2026-10-02 10:00:00+00', '2026-10-02 10:00:00+00', '2026-10-02 10:00:00+00'
);
INSERT INTO public.confidence_point_ledger (
  id, user_id, amount, balance_after, reason, source_type, source_id, idempotency_key, metadata, occurred_at, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000011', 'd0c0a001-0000-4000-8000-000000000001', 5, 7, 'review_correct', 'review_attempt', 'd0c0a001-0000-4000-8000-000000000004', 'restore-rehearsal-v1-review_correct', '{"fixture":"restore-rehearsal-v1"}'::jsonb, '2026-10-02 10:05:00+00', '2026-10-02 10:05:00+00', '2026-10-02 10:05:00+00'
);
INSERT INTO public.confidence_point_ledger (
  id, user_id, amount, balance_after, reason, source_type, source_id, idempotency_key, metadata, occurred_at, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000012', 'd0c0a001-0000-4000-8000-000000000001', 3, 10, 'sentence_submitted', 'learner_sentence', 'd0c0a001-0000-4000-8000-000000000005', 'restore-rehearsal-v1-sentence_submitted', '{"fixture":"restore-rehearsal-v1"}'::jsonb, '2026-10-02 10:06:00+00', '2026-10-02 10:06:00+00', '2026-10-02 10:06:00+00'
);
INSERT INTO public.confidence_point_ledger (
  id, user_id, amount, balance_after, reason, source_type, source_id, idempotency_key, metadata, occurred_at, created_at, updated_at
) VALUES (
  'd0c0a001-0000-4000-8000-000000000013', 'd0c0a001-0000-4000-8000-000000000001', 2, 12, 'ai_feedback_received', 'ai_feedback_attempt', 'd0c0a001-0000-4000-8000-000000000006', 'restore-rehearsal-v1-ai_feedback_received', '{"fixture":"restore-rehearsal-v1"}'::jsonb, '2026-10-02 10:06:01+00', '2026-10-02 10:06:01+00', '2026-10-02 10:06:01+00'
);

-- One saved-word request claim: the real save fingerprint is meaning UUID + source.
INSERT INTO public.idempotency_keys (id, user_id, operation, key, fingerprint, created_at)
VALUES ('d0c0a001-0000-4000-8000-000000000015', 'd0c0a001-0000-4000-8000-000000000001',
        'user_words:save', 'restore-rehearsal-v1-save',
        '3d64c3c9-ede0-5ffd-b1ef-278f6b70e486|journey', '2026-10-02 10:00:00+00');

COMMIT;
