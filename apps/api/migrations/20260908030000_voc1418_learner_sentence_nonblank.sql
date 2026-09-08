-- atlas:txmode file
-- DOC-05 §2/§11 and DOC-09 §6: learner sentence text is meaningful input,
-- and normalized sentence text is the trimmed/collapsed persistence form.
-- Keep this additive rollout NOT VALID so existing retained learner content is
-- not rewritten or blocked while every new write is protected immediately.

ALTER TABLE learner_sentences
  ADD CONSTRAINT learner_sentences_sentence_text_nonblank
    CHECK (sentence_text ~ '[^[:space:]]') NOT VALID,
  ADD CONSTRAINT learner_sentences_normalized_sentence_text_nonblank
    CHECK (normalized_sentence_text ~ '[^[:space:]]') NOT VALID;
