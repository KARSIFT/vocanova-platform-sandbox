-- atlas:txmode file
-- Only shared public canonical content is cached; no learner text or identity.
CREATE TABLE pronunciation_audio_cache (
 cache_key text PRIMARY KEY CHECK(cache_key ~ '^[0-9a-f]{64}$'),
 status text NOT NULL CHECK(status IN('pending','ready','failed')),
 generation_id uuid NOT NULL,
 attempted_day date NOT NULL,
 lease_until timestamptz NOT NULL,
 audio bytea,
 created_at timestamptz NOT NULL,
 updated_at timestamptz NOT NULL,
 CONSTRAINT pronunciation_audio_ready CHECK ((status='ready' AND audio IS NOT NULL AND octet_length(audio) BETWEEN 3 AND 1048576) OR(status<>'ready' AND audio IS NULL))
);
CREATE TABLE pronunciation_audio_budget (
 utc_day date PRIMARY KEY,
 reserved_characters integer NOT NULL CHECK(reserved_characters BETWEEN 0 AND 1000)
);
