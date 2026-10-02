-- Independent fixed expectations: success never depends only on source/target equality.
-- No data rows or learner/provider contents are returned; only named boolean evidence.
\set ON_ERROR_STOP on
-- All functional probes below are rolled back, including a successful settings write.
BEGIN;
SET LOCAL search_path = public, pg_catalog;
SET LOCAL TIME ZONE 'UTC';
DO $check$
DECLARE
  changed integer;
  violated_constraint text;
BEGIN
  IF current_database() <> 'vocanova_rehearsal' THEN
    RAISE EXCEPTION 'rehearsal_check_wrong_database';
  END IF;

  -- Required current seed inventory and stable selected sense.
  IF (SELECT count(*) FROM public.journey_situations) <> 7 THEN
    RAISE EXCEPTION 'rehearsal_canonical_count_journey_situations';
  END IF;
  IF (SELECT count(*) FROM public.canonical_words) <> 51 THEN
    RAISE EXCEPTION 'rehearsal_canonical_count_canonical_words';
  END IF;
  IF (SELECT count(*) FROM public.word_meanings) <> 54 THEN
    RAISE EXCEPTION 'rehearsal_canonical_count_word_meanings';
  END IF;
  IF (SELECT count(*) FROM public.word_examples) <> 72 THEN
    RAISE EXCEPTION 'rehearsal_canonical_count_word_examples';
  END IF;
  IF (SELECT count(*) FROM public.usage_notes) <> 162 THEN
    RAISE EXCEPTION 'rehearsal_canonical_count_usage_notes';
  END IF;
  IF (SELECT count(*) FROM public.journey_words) <> 54 THEN
    RAISE EXCEPTION 'rehearsal_canonical_count_journey_words';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.word_meanings m JOIN public.canonical_words w ON w.id=m.word_id WHERE m.id='3d64c3c9-ede0-5ffd-b1ef-278f6b70e486' AND w.id='812d1cd9-52c8-5726-a7b8-08c094e3fdd2' AND w.normalized_text='catch up' AND w.word_type='phrase' AND m.part_of_speech='verb' AND m.short_definition='To talk after time apart.') THEN
    RAISE EXCEPTION 'rehearsal_canonical_meaning_identity';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.users) <> 1 THEN
    RAISE EXCEPTION 'rehearsal_count_users';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.users WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND email IS NOT DISTINCT FROM 'restore-rehearsal@vocanova.invalid'
      AND display_name IS NOT DISTINCT FROM 'Synthetic restore learner'
      AND avatar_url IS NOT DISTINCT FROM NULL
      AND status IS NOT DISTINCT FROM 'active'
      AND onboarding_status IS NOT DISTINCT FROM 'completed'
      AND email_verified_at IS NOT DISTINCT FROM NULL
      AND last_login_at IS NOT DISTINCT FROM NULL
      AND deleted_at IS NOT DISTINCT FROM NULL
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND is_synthetic_test_account IS NOT DISTINCT FROM true
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_users_1';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.user_settings) <> 1 THEN
    RAISE EXCEPTION 'rehearsal_count_user_settings';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.user_settings WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000002'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND timezone IS NOT DISTINCT FROM 'UTC'
      AND daily_review_target IS NOT DISTINCT FROM 5
      AND review_interval_preset IS NOT DISTINCT FROM 'vocanova_default'
      AND notifications_enabled IS NOT DISTINCT FROM false
      AND marketing_emails_enabled IS NOT DISTINCT FROM false
      AND app_language IS NOT DISTINCT FROM 'en'
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_user_settings_1';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.user_onboarding_profiles) <> 1 THEN
    RAISE EXCEPTION 'rehearsal_count_user_onboarding_profiles';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.user_onboarding_profiles WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000014'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND english_level IS NOT DISTINCT FROM 'a2'
      AND native_language IS NOT DISTINCT FROM 'es'
      AND learning_goal IS NOT DISTINCT FROM 'conversation'
      AND main_use_case IS NOT DISTINCT FROM 'social'
      AND daily_review_target IS NOT DISTINCT FROM 5
      AND completed_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_user_onboarding_profiles_1';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.user_words) <> 1 THEN
    RAISE EXCEPTION 'rehearsal_count_user_words';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.user_words WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000003'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND meaning_id IS NOT DISTINCT FROM '3d64c3c9-ede0-5ffd-b1ef-278f6b70e486'
      AND status IS NOT DISTINCT FROM 'learning'
      AND source IS NOT DISTINCT FROM 'journey'
      AND review_step IS NOT DISTINCT FROM 1
      AND next_review_at IS NOT DISTINCT FROM '2026-10-02 11:05:00+00'
      AND last_reviewed_at IS NOT DISTINCT FROM '2026-10-02 10:05:00+00'
      AND last_result IS NOT DISTINCT FROM 'correct'
      AND last_rating IS NOT DISTINCT FROM 'good'
      AND consecutive_correct_count IS NOT DISTINCT FROM 1
      AND consecutive_incorrect_count IS NOT DISTINCT FROM 0
      AND total_review_count IS NOT DISTINCT FROM 1
      AND correct_review_count IS NOT DISTINCT FROM 1
      AND added_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND mastered_at IS NOT DISTINCT FROM NULL
      AND ignored_at IS NOT DISTINCT FROM NULL
      AND deleted_at IS NOT DISTINCT FROM NULL
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:05:00+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_user_words_1';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.review_attempts) <> 1 THEN
    RAISE EXCEPTION 'rehearsal_count_review_attempts';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.review_attempts WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000004'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND user_word_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000003'
      AND meaning_id IS NOT DISTINCT FROM '3d64c3c9-ede0-5ffd-b1ef-278f6b70e486'
      AND attempt_type IS NOT DISTINCT FROM 'review'
      AND prompt_type IS NOT DISTINCT FROM 'multiple_choice'
      AND result IS NOT DISTINCT FROM 'correct'
      AND rating IS NOT DISTINCT FROM 'good'
      AND review_step_before IS NOT DISTINCT FROM 0
      AND review_step_after IS NOT DISTINCT FROM 1
      AND answered_at IS NOT DISTINCT FROM '2026-10-02 10:05:00+00'
      AND response_time_ms IS NOT DISTINCT FROM 1200
      AND selected_option_meaning_id IS NOT DISTINCT FROM '3d64c3c9-ede0-5ffd-b1ef-278f6b70e486'
      AND typed_answer IS NOT DISTINCT FROM NULL
      AND was_hint_used IS NOT DISTINCT FROM false
      AND source IS NOT DISTINCT FROM 'review_session'
      AND client_attempt_id IS NOT DISTINCT FROM 'restore-rehearsal-v1-review'
      AND metadata IS NOT DISTINCT FROM '{"fixture":"restore-rehearsal-v1"}'::jsonb
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:05:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:05:00+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_review_attempts_1';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.learner_sentences) <> 1 THEN
    RAISE EXCEPTION 'rehearsal_count_learner_sentences';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.learner_sentences WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000005'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND meaning_id IS NOT DISTINCT FROM '3d64c3c9-ede0-5ffd-b1ef-278f6b70e486'
      AND user_word_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000003'
      AND sentence_text IS NOT DISTINCT FROM 'I caught up with Maya and heard about her new job.'
      AND normalized_sentence_text IS NOT DISTINCT FROM 'i caught up with maya and heard about her new job.'
      AND source IS NOT DISTINCT FROM 'word_detail'
      AND status IS NOT DISTINCT FROM 'feedback_ready'
      AND submitted_at IS NOT DISTINCT FROM '2026-10-02 10:06:00+00'
      AND deleted_at IS NOT DISTINCT FROM NULL
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:06:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:06:01+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_learner_sentences_1';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.ai_feedback_attempts) <> 1 THEN
    RAISE EXCEPTION 'rehearsal_count_ai_feedback_attempts';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.ai_feedback_attempts WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000006'
      AND learner_sentence_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000005'
      AND status IS NOT DISTINCT FROM 'succeeded'
      AND provider IS NOT DISTINCT FROM 'synthetic_rehearsal'
      AND model IS NOT DISTINCT FROM 'fixed-fixture-no-provider-call'
      AND prompt_version IS NOT DISTINCT FROM 'restore-rehearsal-v1'
      AND request_hash IS NOT DISTINCT FROM 'restore-rehearsal-v1-sentence-1'
      AND feedback_json IS NOT DISTINCT FROM '{"status":"correct","target_word_used_correctly":true,"grammar_acceptable":true,"meaning_clear":true,"naturalness":"natural","headline":"Good use of catch up.","explanation":"Caught up clearly describes sharing news after time apart."}'::jsonb
      AND feedback_text IS NOT DISTINCT FROM 'Caught up clearly describes sharing news after time apart.'
      AND error_code IS NOT DISTINCT FROM NULL
      AND error_message IS NOT DISTINCT FROM NULL
      AND started_at IS NOT DISTINCT FROM '2026-10-02 10:06:00+00'
      AND completed_at IS NOT DISTINCT FROM '2026-10-02 10:06:01+00'
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:06:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:06:01+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_ai_feedback_attempts_1';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.daily_mission_snapshots) <> 1 THEN
    RAISE EXCEPTION 'rehearsal_count_daily_mission_snapshots';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.daily_mission_snapshots WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000007'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND local_date IS NOT DISTINCT FROM '2026-10-02'
      AND timezone IS NOT DISTINCT FROM 'UTC'
      AND review_target IS NOT DISTINCT FROM 5
      AND reviews_completed IS NOT DISTINCT FROM 1
      AND new_word_target IS NOT DISTINCT FROM 1
      AND new_words_completed IS NOT DISTINCT FROM 1
      AND sentence_practice_target IS NOT DISTINCT FROM 1
      AND sentence_practices_completed IS NOT DISTINCT FROM 1
      AND policy_version IS NOT DISTINCT FROM 'p4-mission-policy-v1'
      AND status IS NOT DISTINCT FROM 'open'
      AND completed_at IS NOT DISTINCT FROM NULL
      AND grace_applied IS NOT DISTINCT FROM false
      AND grace_day_id IS NOT DISTINCT FROM NULL
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:06:01+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_daily_mission_snapshots_1';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.daily_activity_summaries) <> 1 THEN
    RAISE EXCEPTION 'rehearsal_count_daily_activity_summaries';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.daily_activity_summaries WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000008'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND local_date IS NOT DISTINCT FROM '2026-10-02'
      AND timezone IS NOT DISTINCT FROM 'UTC'
      AND reviews_attempted IS NOT DISTINCT FROM 1
      AND reviews_correct IS NOT DISTINCT FROM 1
      AND reviews_skipped IS NOT DISTINCT FROM 0
      AND words_discovered IS NOT DISTINCT FROM 1
      AND words_added IS NOT DISTINCT FROM 1
      AND sentences_submitted IS NOT DISTINCT FROM 1
      AND ai_feedback_received IS NOT DISTINCT FROM 1
      AND confidence_points_earned IS NOT DISTINCT FROM 12
      AND confidence_points_spent IS NOT DISTINCT FROM 0
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:06:01+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_daily_activity_summaries_1';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.streak_states) <> 1 THEN
    RAISE EXCEPTION 'rehearsal_count_streak_states';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.streak_states WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000009'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND current_streak_count IS NOT DISTINCT FROM 0
      AND longest_streak_count IS NOT DISTINCT FROM 0
      AND last_completed_local_date IS NOT DISTINCT FROM NULL
      AND last_activity_local_date IS NOT DISTINCT FROM '2026-10-02'
      AND timezone IS NOT DISTINCT FROM 'UTC'
      AND status IS NOT DISTINCT FROM 'active'
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:06:01+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_streak_states_1';
  END IF;

  -- Exact nonempty row count prevents a missing, duplicate or reassigned record passing.
  IF (SELECT count(*) FROM public.confidence_point_ledger) <> 4 THEN
    RAISE EXCEPTION 'rehearsal_count_confidence_point_ledger';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.confidence_point_ledger WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000010'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND amount IS NOT DISTINCT FROM 2
      AND balance_after IS NOT DISTINCT FROM 2
      AND reason IS NOT DISTINCT FROM 'word_added'
      AND source_type IS NOT DISTINCT FROM 'user_word'
      AND source_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000003'
      AND idempotency_key IS NOT DISTINCT FROM 'restore-rehearsal-v1-word_added'
      AND metadata IS NOT DISTINCT FROM '{"fixture":"restore-rehearsal-v1"}'::jsonb
      AND occurred_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:00:00+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_confidence_point_ledger_1';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.confidence_point_ledger WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000011'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND amount IS NOT DISTINCT FROM 5
      AND balance_after IS NOT DISTINCT FROM 7
      AND reason IS NOT DISTINCT FROM 'review_correct'
      AND source_type IS NOT DISTINCT FROM 'review_attempt'
      AND source_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000004'
      AND idempotency_key IS NOT DISTINCT FROM 'restore-rehearsal-v1-review_correct'
      AND metadata IS NOT DISTINCT FROM '{"fixture":"restore-rehearsal-v1"}'::jsonb
      AND occurred_at IS NOT DISTINCT FROM '2026-10-02 10:05:00+00'
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:05:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:05:00+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_confidence_point_ledger_2';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.confidence_point_ledger WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000012'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND amount IS NOT DISTINCT FROM 3
      AND balance_after IS NOT DISTINCT FROM 10
      AND reason IS NOT DISTINCT FROM 'sentence_submitted'
      AND source_type IS NOT DISTINCT FROM 'learner_sentence'
      AND source_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000005'
      AND idempotency_key IS NOT DISTINCT FROM 'restore-rehearsal-v1-sentence_submitted'
      AND metadata IS NOT DISTINCT FROM '{"fixture":"restore-rehearsal-v1"}'::jsonb
      AND occurred_at IS NOT DISTINCT FROM '2026-10-02 10:06:00+00'
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:06:00+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:06:00+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_confidence_point_ledger_3';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM public.confidence_point_ledger WHERE
      id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000013'
      AND user_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000001'
      AND amount IS NOT DISTINCT FROM 2
      AND balance_after IS NOT DISTINCT FROM 12
      AND reason IS NOT DISTINCT FROM 'ai_feedback_received'
      AND source_type IS NOT DISTINCT FROM 'ai_feedback_attempt'
      AND source_id IS NOT DISTINCT FROM 'd0c0a001-0000-4000-8000-000000000006'
      AND idempotency_key IS NOT DISTINCT FROM 'restore-rehearsal-v1-ai_feedback_received'
      AND metadata IS NOT DISTINCT FROM '{"fixture":"restore-rehearsal-v1"}'::jsonb
      AND occurred_at IS NOT DISTINCT FROM '2026-10-02 10:06:01+00'
      AND created_at IS NOT DISTINCT FROM '2026-10-02 10:06:01+00'
      AND updated_at IS NOT DISTINCT FROM '2026-10-02 10:06:01+00'
  ) THEN
    RAISE EXCEPTION 'rehearsal_state_confidence_point_ledger_4';
  END IF;

  -- Relational accounting independently ties projections to the actual event rows.
  IF NOT EXISTS (
    SELECT 1 FROM public.user_words uw
    JOIN public.review_attempts r ON r.user_word_id=uw.id AND r.user_id=uw.user_id AND r.meaning_id=uw.meaning_id
    JOIN public.learner_sentences s ON s.user_word_id=uw.id AND s.user_id=uw.user_id AND s.meaning_id=uw.meaning_id
    JOIN public.ai_feedback_attempts f ON f.learner_sentence_id=s.id
    JOIN public.daily_activity_summaries a ON a.user_id=uw.user_id AND a.local_date=(r.answered_at AT TIME ZONE a.timezone)::date
    JOIN public.daily_mission_snapshots m ON m.user_id=a.user_id AND m.local_date=a.local_date
    WHERE uw.next_review_at=r.answered_at + interval '1 hour'
      AND uw.review_step=r.review_step_after
      AND uw.total_review_count=(SELECT count(*) FROM public.review_attempts WHERE user_word_id=uw.id)
      AND uw.correct_review_count=(SELECT count(*) FROM public.review_attempts WHERE user_word_id=uw.id AND result='correct')
      AND a.reviews_attempted=(SELECT count(*) FROM public.review_attempts WHERE user_id=uw.user_id)
      AND a.reviews_correct=(SELECT count(*) FROM public.review_attempts WHERE user_id=uw.user_id AND result='correct')
      AND a.words_added=(SELECT count(*) FROM public.user_words WHERE user_id=uw.user_id)
      AND a.sentences_submitted=(SELECT count(*) FROM public.learner_sentences WHERE user_id=uw.user_id)
      AND a.ai_feedback_received=(SELECT count(*) FROM public.ai_feedback_attempts WHERE learner_sentence_id=s.id AND status='succeeded')
      AND m.reviews_completed=a.reviews_attempted-a.reviews_skipped
      AND m.new_words_completed=a.words_added
      AND m.sentence_practices_completed=a.sentences_submitted
      AND a.confidence_points_earned=(SELECT sum(amount) FROM public.confidence_point_ledger WHERE user_id=uw.user_id AND amount>0)
      AND a.confidence_points_spent=0
  ) THEN
    RAISE EXCEPTION 'rehearsal_learning_relationships_or_accounting';
  END IF;
  IF EXISTS (
    SELECT 1 FROM (
      SELECT balance_after, sum(amount) OVER (PARTITION BY user_id ORDER BY occurred_at,id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS running_balance
      FROM public.confidence_point_ledger
    ) balances WHERE balance_after IS DISTINCT FROM running_balance
  ) THEN
    RAISE EXCEPTION 'rehearsal_point_running_balance';
  END IF;
  -- No authentication material belongs to this data-only rehearsal; no grace use was earned.
  IF EXISTS (SELECT 1 FROM public.sessions) OR EXISTS (SELECT 1 FROM public.external_identities)
     OR EXISTS (SELECT 1 FROM public.password_credentials) OR EXISTS (SELECT 1 FROM public.magic_links)
     OR EXISTS (SELECT 1 FROM public.oauth_states)
     OR EXISTS (SELECT 1 FROM public.password_registration_links)
     OR EXISTS (SELECT 1 FROM public.password_reset_links)
     OR EXISTS (SELECT 1 FROM public.email_change_links)
     OR EXISTS (SELECT 1 FROM public.grace_day_ledger) THEN
    RAISE EXCEPTION 'rehearsal_unexpected_auth_or_grace_records';
  END IF;

  IF (SELECT count(*) FROM public.idempotency_keys) <> 1 OR NOT EXISTS (
    SELECT 1 FROM public.idempotency_keys k
    JOIN public.user_words uw ON uw.user_id=k.user_id
      AND k.fingerprint=uw.meaning_id::text || '|' || uw.source
    WHERE k.id='d0c0a001-0000-4000-8000-000000000015'
      AND k.user_id='d0c0a001-0000-4000-8000-000000000001'
      AND k.operation='user_words:save' AND k.key='restore-rehearsal-v1-save'
      AND k.created_at='2026-10-02 10:00:00+00'::timestamptz
  ) THEN
    RAISE EXCEPTION 'rehearsal_idempotency_claim';
  END IF;

  -- Verify selected restored structural protections as well as row contents.
  IF (SELECT count(*) FROM pg_catalog.pg_trigger
      WHERE (tgrelid,tgname) IN (
        ('public.confidence_point_ledger'::regclass,'confidence_point_ledger_append_only'),
        ('public.grace_day_ledger'::regclass,'grace_day_ledger_append_only'))
      AND NOT tgisinternal AND tgenabled='O'
      -- pg_trigger.tgtype: ROW=1, BEFORE=2, DELETE=8, UPDATE=16.
      -- Exact bits reject statement/AFTER/UPDATE-only triggers on either ledger.
      AND tgtype = (1 | 2 | 8 | 16)
      -- Both migration triggers are unconditional, cover every column and take no args.
      AND tgattr = ''::int2vector AND tgqual IS NULL AND tgnargs=0
      AND tgfoid='public.vocanova_reject_learning_ledger_mutation()'::regprocedure) <> 2 THEN
    RAISE EXCEPTION 'rehearsal_append_only_triggers';
  END IF;
  IF (SELECT count(*) FROM pg_catalog.pg_constraint
      WHERE (conrelid,conname) IN (
        ('public.review_attempts'::regclass,'review_attempts_user_word_owner_meaning_fk'),
        ('public.user_words'::regclass,'user_words_last_result_rating_consistent'),
        ('public.ai_feedback_attempts'::regclass,'ai_feedback_attempts_feedback_json_required_on_success'),
        ('public.daily_mission_snapshots'::regclass,'daily_mission_snapshots_completed_at_only_on_done'),
        ('public.idempotency_keys'::regclass,'idempotency_keys_user_id_fkey'),
        ('public.daily_activity_summaries'::regclass,'daily_activity_summaries_confidence_point_counters_nonnegative'))) <> 6 THEN
    RAISE EXCEPTION 'rehearsal_required_constraints';
  END IF;

  -- Functional write: one real restored settings row changes, then outer ROLLBACK undoes it.
  UPDATE public.user_settings SET daily_review_target=6
  WHERE id='d0c0a001-0000-4000-8000-000000000002';
  GET DIAGNOSTICS changed = ROW_COUNT;
  IF changed <> 1 OR NOT EXISTS (SELECT 1 FROM public.user_settings
      WHERE id='d0c0a001-0000-4000-8000-000000000002' AND daily_review_target=6) THEN
    RAISE EXCEPTION 'rehearsal_restored_write_probe';
  END IF;

  -- A missing trigger must fail this test; catch only its actual rejection code.
  BEGIN
    UPDATE public.confidence_point_ledger SET amount=amount+1
    WHERE id='d0c0a001-0000-4000-8000-000000000010';
    RAISE EXCEPTION 'rehearsal_ledger_update_was_not_rejected';
  EXCEPTION WHEN SQLSTATE '55000' THEN
    NULL;
  END;
  BEGIN
    DELETE FROM public.confidence_point_ledger
    WHERE id='d0c0a001-0000-4000-8000-000000000010';
    RAISE EXCEPTION 'rehearsal_ledger_delete_was_not_rejected';
  EXCEPTION WHEN SQLSTATE '55000' THEN
    NULL;
  END;
  BEGIN
    UPDATE public.user_words SET last_result='incorrect',last_rating='good'
    WHERE id='d0c0a001-0000-4000-8000-000000000003';
    RAISE EXCEPTION 'rehearsal_invalid_rating_was_not_rejected';
  EXCEPTION WHEN check_violation THEN
    GET STACKED DIAGNOSTICS violated_constraint = CONSTRAINT_NAME;
    IF violated_constraint <> 'user_words_last_result_rating_consistent' THEN
      RAISE EXCEPTION 'rehearsal_unexpected_rating_constraint';
    END IF;
  END;
  BEGIN
    -- Another valid canonical meaning passes the simple FK but must fail ownership binding.
    UPDATE public.review_attempts SET meaning_id='3265c02c-5751-5465-b94b-e767aa871d8d'
    WHERE id='d0c0a001-0000-4000-8000-000000000004';
    RAISE EXCEPTION 'rehearsal_mismatched_review_meaning_was_not_rejected';
  EXCEPTION WHEN foreign_key_violation THEN
    GET STACKED DIAGNOSTICS violated_constraint = CONSTRAINT_NAME;
    IF violated_constraint <> 'review_attempts_user_word_owner_meaning_fk' THEN
      RAISE EXCEPTION 'rehearsal_unexpected_review_foreign_key';
    END IF;
  END;
  -- Each remaining named guard must reject a write that only it should prohibit.
  BEGIN
    UPDATE public.ai_feedback_attempts SET feedback_json=NULL
    WHERE id='d0c0a001-0000-4000-8000-000000000006';
    RAISE EXCEPTION 'rehearsal_missing_success_feedback_was_not_rejected';
  EXCEPTION WHEN check_violation THEN
    GET STACKED DIAGNOSTICS violated_constraint = CONSTRAINT_NAME;
    IF violated_constraint <> 'ai_feedback_attempts_feedback_json_required_on_success' THEN
      RAISE EXCEPTION 'rehearsal_unexpected_feedback_constraint';
    END IF;
  END;
  BEGIN
    UPDATE public.daily_mission_snapshots SET completed_at='2026-10-02 10:06:01+00'
    WHERE id='d0c0a001-0000-4000-8000-000000000007';
    RAISE EXCEPTION 'rehearsal_open_mission_completion_was_not_rejected';
  EXCEPTION WHEN check_violation THEN
    GET STACKED DIAGNOSTICS violated_constraint = CONSTRAINT_NAME;
    IF violated_constraint <> 'daily_mission_snapshots_completed_at_only_on_done' THEN
      RAISE EXCEPTION 'rehearsal_unexpected_mission_completion_constraint';
    END IF;
  END;
  BEGIN
    -- The exact single-user fixture above guarantees this different user is absent.
    UPDATE public.idempotency_keys SET user_id='d0c0a001-0000-4000-8000-999999999999'
    WHERE id='d0c0a001-0000-4000-8000-000000000015';
    RAISE EXCEPTION 'rehearsal_orphan_idempotency_claim_was_not_rejected';
  EXCEPTION WHEN foreign_key_violation THEN
    GET STACKED DIAGNOSTICS violated_constraint = CONSTRAINT_NAME;
    IF violated_constraint <> 'idempotency_keys_user_id_fkey' THEN
      RAISE EXCEPTION 'rehearsal_unexpected_idempotency_foreign_key';
    END IF;
  END;
  BEGIN
    UPDATE public.daily_activity_summaries SET confidence_points_earned=-1
    WHERE id='d0c0a001-0000-4000-8000-000000000008';
    RAISE EXCEPTION 'rehearsal_negative_points_earned_was_not_rejected';
  EXCEPTION WHEN check_violation THEN
    GET STACKED DIAGNOSTICS violated_constraint = CONSTRAINT_NAME;
    IF violated_constraint <> 'daily_activity_summaries_confidence_point_counters_nonnegative' THEN
      RAISE EXCEPTION 'rehearsal_unexpected_points_earned_constraint';
    END IF;
  END;
  BEGIN
    UPDATE public.daily_activity_summaries SET confidence_points_spent=-1
    WHERE id='d0c0a001-0000-4000-8000-000000000008';
    RAISE EXCEPTION 'rehearsal_negative_points_spent_was_not_rejected';
  EXCEPTION WHEN check_violation THEN
    GET STACKED DIAGNOSTICS violated_constraint = CONSTRAINT_NAME;
    IF violated_constraint <> 'daily_activity_summaries_confidence_point_counters_nonnegative' THEN
      RAISE EXCEPTION 'rehearsal_unexpected_points_spent_constraint';
    END IF;
  END;
END;
$check$;

ROLLBACK;

-- Confirm that successful and rejected probes left the reference state unchanged.
BEGIN TRANSACTION READ ONLY;
DO $rolled_back$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM public.user_settings
      WHERE id='d0c0a001-0000-4000-8000-000000000002' AND daily_review_target=5)
    OR NOT EXISTS (SELECT 1 FROM public.confidence_point_ledger
      WHERE id='d0c0a001-0000-4000-8000-000000000010' AND amount=2 AND balance_after=2)
    OR NOT EXISTS (SELECT 1 FROM public.user_words
      WHERE id='d0c0a001-0000-4000-8000-000000000003' AND last_result='correct' AND last_rating='good')
    OR NOT EXISTS (SELECT 1 FROM public.review_attempts
      WHERE id='d0c0a001-0000-4000-8000-000000000004' AND meaning_id='3d64c3c9-ede0-5ffd-b1ef-278f6b70e486')
    OR NOT EXISTS (SELECT 1 FROM public.ai_feedback_attempts
      WHERE id='d0c0a001-0000-4000-8000-000000000006' AND status='succeeded'
        AND feedback_json='{"status":"correct","target_word_used_correctly":true,"grammar_acceptable":true,"meaning_clear":true,"naturalness":"natural","headline":"Good use of catch up.","explanation":"Caught up clearly describes sharing news after time apart."}'::jsonb)
    OR NOT EXISTS (SELECT 1 FROM public.daily_mission_snapshots
      WHERE id='d0c0a001-0000-4000-8000-000000000007' AND status='open' AND completed_at IS NULL)
    OR NOT EXISTS (SELECT 1 FROM public.idempotency_keys
      WHERE id='d0c0a001-0000-4000-8000-000000000015' AND user_id='d0c0a001-0000-4000-8000-000000000001')
    OR NOT EXISTS (SELECT 1 FROM public.daily_activity_summaries
      WHERE id='d0c0a001-0000-4000-8000-000000000008' AND confidence_points_earned=12 AND confidence_points_spent=0) THEN
    RAISE EXCEPTION 'rehearsal_probe_rollback_state';
  END IF;
END;
$rolled_back$;

SELECT json_build_object('canonical_seed',true,'synthetic_identity',true,'settings_and_onboarding',true,'saved_word_schedule',true,'review_history',true,'sentence_and_synthetic_feedback',true,'mission_and_activity',true,'point_ledger',true,'streak_state',true,'learning_relationships',true,'no_auth_material',true,'idempotency_claim',true,'schema_guards',true,'restored_write_and_protections',true);
COMMIT;
