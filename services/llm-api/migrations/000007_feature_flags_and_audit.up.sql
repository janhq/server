-- Feature flags definitions and audit logging for admin actions
SET search_path TO llm_api;

-- Feature flag definitions (metadata only)
CREATE TABLE IF NOT EXISTS llm_api.feature_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feature_flags_key ON llm_api.feature_flags(key);
CREATE INDEX IF NOT EXISTS idx_feature_flags_category ON llm_api.feature_flags(category);

INSERT INTO llm_api.feature_flags (key, name, description, category)
VALUES ('experimental_models', 'Experimental Models', 'Access to experimental/beta models in model catalog', 'model_access')
ON CONFLICT (key) DO NOTHING;

-- Audit log for admin actions
CREATE TABLE IF NOT EXISTS llm_api.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id VARCHAR(255) NOT NULL,
    admin_email VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255),
    payload JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    status_code INTEGER,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON llm_api.audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_admin_user_id ON llm_api.audit_logs(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON llm_api.audit_logs(resource_type);

-- Experimental flag on model catalogs
ALTER TABLE llm_api.model_catalogs ADD COLUMN IF NOT EXISTS experimental BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_model_catalogs_experimental ON llm_api.model_catalogs(experimental);
