ALTER TABLE change_runtime_states
    ADD COLUMN IF NOT EXISTS artifact_state jsonb NOT NULL DEFAULT '{}'::jsonb;
