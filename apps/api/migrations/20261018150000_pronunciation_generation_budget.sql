ALTER TABLE pronunciation_audio_budget
    ADD COLUMN reserved_generations integer NOT NULL DEFAULT 0,
    ADD CONSTRAINT pronunciation_audio_budget_generation_cap
        CHECK (reserved_generations BETWEEN 0 AND 20);
