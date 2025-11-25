-- ============================================================================
-- USER SETTINGS
-- ============================================================================
-- Stores user preferences and feature toggles for personalization and control

CREATE TABLE llm_api.user_settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    
    -- Memory Feature Controls
    memory_enabled BOOLEAN NOT NULL DEFAULT true,
    memory_auto_inject BOOLEAN NOT NULL DEFAULT false,
    memory_inject_user_core BOOLEAN NOT NULL DEFAULT false,
    memory_inject_project BOOLEAN NOT NULL DEFAULT false,
    memory_inject_conversation BOOLEAN NOT NULL DEFAULT false,
    
    -- Memory Retrieval Preferences
    memory_max_user_items INTEGER NOT NULL DEFAULT 3,
    memory_max_project_items INTEGER NOT NULL DEFAULT 5,
    memory_max_episodic_items INTEGER NOT NULL DEFAULT 3,
    memory_min_similarity NUMERIC(3,2) NOT NULL DEFAULT 0.75,
    
    -- Other Feature Toggles
    enable_trace BOOLEAN NOT NULL DEFAULT false,
    enable_tools BOOLEAN NOT NULL DEFAULT true,
    
    -- Preferences stored as flexible JSON for future extensions
    preferences JSONB NOT NULL DEFAULT '{}',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_user_settings_user FOREIGN KEY (user_id) REFERENCES llm_api.users(id) ON DELETE CASCADE,
    CONSTRAINT ux_user_settings_user_id UNIQUE (user_id)
);

CREATE INDEX idx_user_settings_user_id ON llm_api.user_settings(user_id);
CREATE INDEX idx_user_settings_memory_enabled ON llm_api.user_settings(memory_enabled);

-- Add helpful comment
COMMENT ON TABLE llm_api.user_settings IS 'User preferences and feature toggles including memory controls';
COMMENT ON COLUMN llm_api.user_settings.memory_enabled IS 'Master toggle for memory features (observation and retrieval)';
COMMENT ON COLUMN llm_api.user_settings.memory_auto_inject IS 'Automatically inject memory at conversation start (default: false)';
COMMENT ON COLUMN llm_api.user_settings.memory_inject_user_core IS 'Inject user profile core (language, role, etc.) when enabled';
COMMENT ON COLUMN llm_api.user_settings.memory_inject_project IS 'Inject project context when enabled';
COMMENT ON COLUMN llm_api.user_settings.memory_inject_conversation IS 'Inject conversation history summaries when enabled';
