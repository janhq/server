ALTER TABLE llm_api.model_catalogs
    DROP COLUMN IF EXISTS model_display_name;

ALTER TABLE llm_api.model_catalogs
    DROP COLUMN IF EXISTS context_length;
