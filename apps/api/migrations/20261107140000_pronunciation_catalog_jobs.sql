-- Explicit operator jobs are separate from public on-demand generation budgets.
CREATE TABLE pronunciation_catalog_jobs (
 id uuid PRIMARY KEY,
 provider_identity text NOT NULL,
 max_generations integer NOT NULL CHECK(max_generations BETWEEN 1 AND 1000),
 max_characters integer NOT NULL CHECK(max_characters BETWEEN 1 AND 500000),
 max_audio_seconds double precision NOT NULL CHECK(max_audio_seconds > 0 AND max_audio_seconds <= 60000),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE pronunciation_catalog_attempts (
 id uuid PRIMARY KEY,
 job_id uuid NOT NULL REFERENCES pronunciation_catalog_jobs(id),
 cache_key text NOT NULL CHECK(cache_key ~ '^[0-9a-f]{64}$'),
 meaning_id uuid NOT NULL REFERENCES word_meanings(id),
 meaning_revision text NOT NULL CHECK(meaning_revision ~ '^[0-9a-f]{64}$'),
 characters integer NOT NULL CHECK(characters BETWEEN 1 AND 500),
 status text NOT NULL CHECK(status IN ('reserved','ready','failed','changed')),
 audio_seconds double precision NOT NULL DEFAULT 0 CHECK(audio_seconds >= 0 AND audio_seconds <= 60),
 created_at timestamptz NOT NULL DEFAULT now(),
 finished_at timestamptz,
 UNIQUE(job_id,cache_key)
);
