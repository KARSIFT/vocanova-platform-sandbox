-- Public source inventory only; learner assessments and discovery receipts reuse
-- user_word_knowledge, user_words and word_knowledge_actions lifecycle handling.
CREATE UNIQUE INDEX word_meanings_catalog_owner_idx ON word_meanings(id,word_id);
CREATE TABLE knowledge_catalog_entries (
 rank integer PRIMARY KEY CHECK (rank BETWEEN 1 AND 10000),
 word_id uuid NOT NULL UNIQUE REFERENCES canonical_words(id),
 meaning_id uuid NOT NULL UNIQUE REFERENCES word_meanings(id),
 source_rank integer NOT NULL CHECK (source_rank > 0),
 source_count bigint NOT NULL CHECK (source_count > 0),
 source_sense_key text NOT NULL,
 source_name text NOT NULL,
 catalog_version text NOT NULL CHECK (catalog_version = 'english-10000-v1'),
 CONSTRAINT knowledge_catalog_meaning_owner FOREIGN KEY (meaning_id,word_id) REFERENCES word_meanings(id,word_id)
);