-- atlas:txmode file
-- Shared public teaching media only. Original files and learner data never enter
-- this table; URLs are built from the fixed public origin in application code.
CREATE TABLE meaning_image_assets (
 id uuid PRIMARY KEY,
 meaning_id uuid NOT NULL REFERENCES word_meanings(id) ON DELETE RESTRICT,
 meaning_revision text NOT NULL CHECK (meaning_revision ~ '^[0-9a-f]{64}$'),
 sha256 text NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
 alt text NOT NULL CHECK (char_length(alt) >= 1 AND char_length(alt) <= 500 AND alt=btrim(alt) AND alt !~ E'[\r\n]'),
 width integer NOT NULL CHECK (width BETWEEN 1 AND 960),
 height integer NOT NULL CHECK (height BETWEEN 1 AND 640),
 approved boolean NOT NULL DEFAULT false,
 created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
 UNIQUE (meaning_id,meaning_revision)
);
