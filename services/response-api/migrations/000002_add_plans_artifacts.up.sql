-- ============================================================================
-- Migration: Add Plans and Artifacts for Agent Response Architecture
-- ============================================================================

SET search_path TO response_api;

-- ============================================================================
-- PLANS - Multi-step agent execution plans
-- ============================================================================
CREATE TABLE response_api.plans (
    id SERIAL PRIMARY KEY,
    public_id VARCHAR(64) NOT NULL UNIQUE,
    
    -- Relationships
    conversation_id INTEGER REFERENCES response_api.conversations(id),
    response_id INTEGER REFERENCES response_api.responses(id),
    user_id VARCHAR(64) NOT NULL,
    
    -- Plan metadata
    title VARCHAR(500),
    description TEXT,
    agent_type VARCHAR(100),  -- 'slide_generator', 'deep_research', etc.
    
    -- Status tracking
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
        -- pending, planning, in_progress, wait_for_user, completed, failed, cancelled, expired
    progress INTEGER DEFAULT 0,  -- 0-100 percentage
    
    -- Timing
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    
    -- Result
    artifact_id VARCHAR(100),  -- art_* reference if plan produces artifact
    error_code VARCHAR(50),
    error_message TEXT,
    
    -- User selection data (for wait_for_user state)
    user_selection JSONB,
    
    -- Metadata (flexible JSON for agent-specific data)
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_plans_conversation ON response_api.plans(conversation_id);
CREATE INDEX idx_plans_response ON response_api.plans(response_id);
CREATE INDEX idx_plans_user ON response_api.plans(user_id);
CREATE INDEX idx_plans_status ON response_api.plans(status);
CREATE INDEX idx_plans_agent_type ON response_api.plans(agent_type);

-- ============================================================================
-- PLAN_TASKS - Tasks within a plan
-- ============================================================================
CREATE TABLE response_api.plan_tasks (
    id SERIAL PRIMARY KEY,
    public_id VARCHAR(64) NOT NULL UNIQUE,
    plan_id INTEGER NOT NULL REFERENCES response_api.plans(id) ON DELETE CASCADE,
    
    -- Task definition
    sequence INTEGER NOT NULL,  -- Order within plan (1, 2, 3...)
    title VARCHAR(500) NOT NULL,
    description TEXT,
    task_type VARCHAR(100),  -- 'research', 'synthesis', 'generation', 'storage'
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
        -- pending, in_progress, completed, failed, skipped
    progress INTEGER DEFAULT 0,
    
    -- Timing
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,
    
    -- Result/Error
    result_summary TEXT,
    error_code VARCHAR(50),
    error_message TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_plan_tasks_plan ON response_api.plan_tasks(plan_id);
CREATE INDEX idx_plan_tasks_sequence ON response_api.plan_tasks(plan_id, sequence);
CREATE INDEX idx_plan_tasks_status ON response_api.plan_tasks(status);

-- ============================================================================
-- PLAN_STEPS - Steps within a task
-- ============================================================================
CREATE TABLE response_api.plan_steps (
    id SERIAL PRIMARY KEY,
    public_id VARCHAR(64) NOT NULL UNIQUE,
    task_id INTEGER NOT NULL REFERENCES response_api.plan_tasks(id) ON DELETE CASCADE,
    plan_id INTEGER NOT NULL REFERENCES response_api.plans(id) ON DELETE CASCADE,
    
    -- Step definition
    sequence INTEGER NOT NULL,  -- Order within task
    action VARCHAR(100) NOT NULL,  -- 'web_search', 'scrape', 'reasoning', 'generate', 'wait_for_user'
    description TEXT,
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
        -- pending, in_progress, completed, failed, skipped
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    
    -- Timing
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,
    
    -- Input/Output
    input_params JSONB,  -- Parameters passed to this step
    output_data JSONB,   -- Result of this step
    
    -- Error handling
    error_code VARCHAR(50),
    error_message TEXT,
    error_severity VARCHAR(20),  -- retryable, fallback, skippable, fatal
    
    -- Links to conversation items (tool calls, results, reasoning)
    conversation_item_ids TEXT[],  -- Array of linked conversation_item public IDs
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_plan_steps_task ON response_api.plan_steps(task_id);
CREATE INDEX idx_plan_steps_plan ON response_api.plan_steps(plan_id);
CREATE INDEX idx_plan_steps_sequence ON response_api.plan_steps(task_id, sequence);
CREATE INDEX idx_plan_steps_action ON response_api.plan_steps(action);
CREATE INDEX idx_plan_steps_status ON response_api.plan_steps(status);

-- ============================================================================
-- PLAN_STEP_DETAILS - Links steps to their outputs
-- ============================================================================
CREATE TABLE response_api.plan_step_details (
    id SERIAL PRIMARY KEY,
    public_id VARCHAR(64) NOT NULL UNIQUE,
    step_id INTEGER NOT NULL REFERENCES response_api.plan_steps(id) ON DELETE CASCADE,
    
    -- Detail type and reference
    detail_type VARCHAR(50) NOT NULL,
        -- 'tool_call', 'tool_result', 'reasoning', 'response', 'artifact', 'user_input', 'error'
    
    -- References (one will be set based on type)
    conversation_item_id INTEGER REFERENCES response_api.conversation_items(id),
    tool_call_id VARCHAR(100),
    artifact_id VARCHAR(100),
    
    -- Inline content (for reasoning/response that aren't separate items)
    content TEXT,
    
    -- Order within step
    sequence INTEGER NOT NULL DEFAULT 0,
    
    -- Timing
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_plan_step_details_step ON response_api.plan_step_details(step_id);
CREATE INDEX idx_plan_step_details_type ON response_api.plan_step_details(detail_type);
CREATE INDEX idx_plan_step_details_conv_item ON response_api.plan_step_details(conversation_item_id);

-- ============================================================================
-- ARTIFACTS - Generated content (presentations, documents, etc.)
-- ============================================================================
CREATE TABLE response_api.artifacts (
    id VARCHAR(100) PRIMARY KEY,  -- art_pres_abc123
    
    -- Versioning
    version INTEGER NOT NULL DEFAULT 1,
    parent_id VARCHAR(100) REFERENCES response_api.artifacts(id),
    is_latest BOOLEAN NOT NULL DEFAULT true,
    
    -- Ownership
    user_id VARCHAR(64) NOT NULL,
    conversation_id INTEGER REFERENCES response_api.conversations(id),
    plan_id INTEGER REFERENCES response_api.plans(id),
    
    -- Content
    type VARCHAR(50) NOT NULL,  -- presentation, document, code, image
    format VARCHAR(20) NOT NULL,  -- pptx, pdf, docx, png
    title VARCHAR(500),
    content_hash VARCHAR(64),  -- SHA256 for deduplication
    size_bytes BIGINT,
    
    -- Storage
    storage_path TEXT NOT NULL,  -- s3://bucket/path
    storage_region VARCHAR(50),
    
    -- Preview
    preview_path TEXT,  -- Path to preview/thumbnail
    thumbnail_paths TEXT[],  -- Array of thumbnail paths
    
    -- Lifecycle
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,  -- Soft delete after this
    deleted_at TIMESTAMPTZ,  -- Soft delete timestamp
    
    -- Retention policy
    retention_policy VARCHAR(50) DEFAULT 'standard',
        -- 'temporary': 24 hours
        -- 'standard': 90 days
        -- 'extended': 1 year
        -- 'permanent': never expires
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_artifacts_parent ON response_api.artifacts(parent_id);
CREATE INDEX idx_artifacts_latest ON response_api.artifacts(user_id, type, is_latest) WHERE is_latest = true;
CREATE INDEX idx_artifacts_expires ON response_api.artifacts(expires_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_artifacts_user ON response_api.artifacts(user_id);
CREATE INDEX idx_artifacts_conversation ON response_api.artifacts(conversation_id);
CREATE INDEX idx_artifacts_plan ON response_api.artifacts(plan_id);
CREATE INDEX idx_artifacts_type ON response_api.artifacts(type);

-- ============================================================================
-- IDEMPOTENCY_KEYS - Prevent duplicate requests
-- ============================================================================
CREATE TABLE response_api.idempotency_keys (
    key VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    response_id INTEGER REFERENCES response_api.responses(id),
    plan_id INTEGER REFERENCES response_api.plans(id),
    status VARCHAR(50) NOT NULL,
    request_hash VARCHAR(64),  -- SHA256 of request body
    
    CONSTRAINT unique_user_key UNIQUE(user_id, key)
);

CREATE INDEX idx_idempotency_expires ON response_api.idempotency_keys(expires_at);
CREATE INDEX idx_idempotency_user ON response_api.idempotency_keys(user_id);

-- ============================================================================
-- ALTER CONVERSATION_ITEMS - Add tree structure and tool fields
-- ============================================================================

-- Tree structure for nested tools
ALTER TABLE response_api.conversation_items 
    ADD COLUMN IF NOT EXISTS parent_item_id INTEGER REFERENCES response_api.conversation_items(id),
    ADD COLUMN IF NOT EXISTS root_item_id INTEGER REFERENCES response_api.conversation_items(id),
    ADD COLUMN IF NOT EXISTS depth INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS path TEXT;

-- Tool-specific fields (extend existing)
ALTER TABLE response_api.conversation_items
    ADD COLUMN IF NOT EXISTS tool_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS tool_call_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS tool_arguments JSONB,
    ADD COLUMN IF NOT EXISTS tool_result JSONB;

-- Error handling fields
ALTER TABLE response_api.conversation_items
    ADD COLUMN IF NOT EXISTS error_code VARCHAR(50),
    ADD COLUMN IF NOT EXISTS error_message TEXT,
    ADD COLUMN IF NOT EXISTS error_details JSONB;

-- Performance metrics
ALTER TABLE response_api.conversation_items
    ADD COLUMN IF NOT EXISTS latency_ms INTEGER,
    ADD COLUMN IF NOT EXISTS tokens_used INTEGER;

-- Provider tracking
ALTER TABLE response_api.conversation_items
    ADD COLUMN IF NOT EXISTS provider VARCHAR(50);

-- Create indexes for tree traversal
CREATE INDEX IF NOT EXISTS idx_conv_items_parent ON response_api.conversation_items(parent_item_id);
CREATE INDEX IF NOT EXISTS idx_conv_items_root ON response_api.conversation_items(root_item_id);
CREATE INDEX IF NOT EXISTS idx_conv_items_path ON response_api.conversation_items(path);
CREATE INDEX IF NOT EXISTS idx_conv_items_tool_call ON response_api.conversation_items(tool_call_id);
CREATE INDEX IF NOT EXISTS idx_conv_items_tool_name ON response_api.conversation_items(tool_name);

-- ============================================================================
-- ALTER RESPONSES - Add plan reference
-- ============================================================================
ALTER TABLE response_api.responses
    ADD COLUMN IF NOT EXISTS plan_id INTEGER REFERENCES response_api.plans(id);

CREATE INDEX IF NOT EXISTS idx_responses_plan ON response_api.responses(plan_id);
