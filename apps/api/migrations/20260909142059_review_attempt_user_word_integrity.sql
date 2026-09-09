-- atlas:txmode file
-- Issue #1388: review history repeats the owning learner and meaning so those
-- values can be queried without joining user_words. Bind the repeated values
-- to the referenced saved-word row so immutable history cannot become
-- cross-learner or describe a different meaning.

CREATE UNIQUE INDEX user_words_id_user_id_meaning_id_key
  ON user_words (id, user_id, meaning_id);

ALTER TABLE review_attempts
  ADD CONSTRAINT review_attempts_user_word_owner_meaning_fk
  FOREIGN KEY (user_word_id, user_id, meaning_id)
  REFERENCES user_words (id, user_id, meaning_id)
  ON DELETE RESTRICT
  NOT VALID;
