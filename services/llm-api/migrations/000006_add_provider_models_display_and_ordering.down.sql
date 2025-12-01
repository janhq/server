-- Remove model_display_name, category, category_order_number, and model_order_number columns from provider_models table

-- Drop indexes first
DROP INDEX IF EXISTS llm_api.idx_provider_models_category_sorting;
DROP INDEX IF EXISTS llm_api.idx_provider_models_model_order;
DROP INDEX IF EXISTS llm_api.idx_provider_models_category_order;
DROP INDEX IF EXISTS llm_api.idx_provider_models_category;

-- Drop columns
ALTER TABLE llm_api.provider_models
    DROP COLUMN IF EXISTS model_order_number;

ALTER TABLE llm_api.provider_models
    DROP COLUMN IF EXISTS category_order_number;

ALTER TABLE llm_api.provider_models
    DROP COLUMN IF EXISTS category;

ALTER TABLE llm_api.provider_models
    DROP COLUMN IF EXISTS model_display_name;
