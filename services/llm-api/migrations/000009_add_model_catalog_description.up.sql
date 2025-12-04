-- Add description column to model_catalogs to separate from notes
ALTER TABLE llm_api.model_catalogs
    ADD COLUMN IF NOT EXISTS description TEXT;
