-- Add display name and context length to model_catalogs
ALTER TABLE llm_api.model_catalogs
    ADD COLUMN IF NOT EXISTS model_display_name VARCHAR(255) NOT NULL DEFAULT '';

ALTER TABLE llm_api.model_catalogs
    ADD COLUMN IF NOT EXISTS context_length INTEGER;
