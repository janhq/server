-- Add model_display_name, category, category_order_number, and model_order_number columns to provider_models table
-- These fields support grouping and sorting models in the UI

-- Add model_display_name column (user-friendly display name, defaults to model_public_id)
ALTER TABLE llm_api.provider_models
    ADD COLUMN IF NOT EXISTS model_display_name VARCHAR(255) NOT NULL DEFAULT '';

-- Add category column (for grouping models, e.g., "Chat", "Embedding", "Image")
ALTER TABLE llm_api.provider_models
    ADD COLUMN IF NOT EXISTS category VARCHAR(128) NOT NULL DEFAULT '';

-- Add category_order_number column (for sorting categories)
ALTER TABLE llm_api.provider_models
    ADD COLUMN IF NOT EXISTS category_order_number INTEGER NOT NULL DEFAULT 0;

-- Add model_order_number column (for sorting models within a category)
ALTER TABLE llm_api.provider_models
    ADD COLUMN IF NOT EXISTS model_order_number INTEGER NOT NULL DEFAULT 0;

-- Add indexes for efficient sorting and filtering
CREATE INDEX IF NOT EXISTS idx_provider_models_category 
    ON llm_api.provider_models(category);

CREATE INDEX IF NOT EXISTS idx_provider_models_category_order 
    ON llm_api.provider_models(category_order_number);

CREATE INDEX IF NOT EXISTS idx_provider_models_model_order 
    ON llm_api.provider_models(model_order_number);

-- Composite index for efficient category-based sorting
CREATE INDEX IF NOT EXISTS idx_provider_models_category_sorting 
    ON llm_api.provider_models(category_order_number, model_order_number);

-- Add comments for documentation
COMMENT ON COLUMN llm_api.provider_models.model_display_name IS 'User-friendly display name for the model (defaults to model_public_id if empty)';
COMMENT ON COLUMN llm_api.provider_models.category IS 'Category for grouping models (e.g., Chat, Embedding, Image)';
COMMENT ON COLUMN llm_api.provider_models.category_order_number IS 'Order number for sorting categories (lower numbers appear first)';
COMMENT ON COLUMN llm_api.provider_models.model_order_number IS 'Order number for sorting models within a category (lower numbers appear first)';
