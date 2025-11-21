-- Migration: Create memory tables with pgvector support
-- Version: 001
-- Date: 2025-11-20

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- User Memory Items Table
CREATE TABLE IF NOT EXISTS user_memory_items (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    scope VARCHAR(50) NOT NULL, -- 'core', 'preference', 'context'
    key VARCHAR(255) NOT NULL,
    text TEXT NOT NULL,
    score INTEGER NOT NULL DEFAULT 3, -- 1-5, importance level
    embedding vector(1024), -- BGE-M3 embeddings
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT user_memory_scope_check CHECK (scope IN ('core', 'preference', 'context')),
    CONSTRAINT user_memory_score_check CHECK (score >= 1 AND score <= 5)
);

-- Indexes for user_memory_items
CREATE INDEX IF NOT EXISTS idx_user_memory_user_id ON user_memory_items(user_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_user_memory_scope ON user_memory_items(scope) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_user_memory_score ON user_memory_items(score DESC) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_user_memory_updated_at ON user_memory_items(updated_at DESC);

-- Vector similarity index (IVFFlat for fast approximate search)
CREATE INDEX IF NOT EXISTS idx_user_memory_embedding ON user_memory_items 
USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100)
WHERE is_deleted = FALSE;

-- Project Facts Table
CREATE TABLE IF NOT EXISTS project_facts (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    kind VARCHAR(50) NOT NULL, -- 'decision', 'requirement', 'constraint', 'context'
    title VARCHAR(500) NOT NULL,
    text TEXT NOT NULL,
    confidence REAL NOT NULL DEFAULT 0.8, -- 0.0-1.0
    embedding vector(1024), -- BGE-M3 embeddings
    source_conversation_id VARCHAR(255),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT project_fact_kind_check CHECK (kind IN ('decision', 'requirement', 'constraint', 'context')),
    CONSTRAINT project_fact_confidence_check CHECK (confidence >= 0.0 AND confidence <= 1.0)
);

-- Indexes for project_facts
CREATE INDEX IF NOT EXISTS idx_project_facts_project_id ON project_facts(project_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_project_facts_kind ON project_facts(kind) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_project_facts_confidence ON project_facts(confidence DESC) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_project_facts_updated_at ON project_facts(updated_at DESC);

-- Vector similarity index
CREATE INDEX IF NOT EXISTS idx_project_facts_embedding ON project_facts 
USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100)
WHERE is_deleted = FALSE;

-- Update constraints to allow all supported kinds/scopes
ALTER TABLE user_memory_items DROP CONSTRAINT IF EXISTS user_memory_scope_check;
ALTER TABLE user_memory_items ADD CONSTRAINT user_memory_scope_check CHECK (scope IN ('core', 'preference', 'context', 'profile', 'skill'));

ALTER TABLE project_facts DROP CONSTRAINT IF EXISTS project_fact_kind_check;
ALTER TABLE project_facts ADD CONSTRAINT project_fact_kind_check CHECK (kind IN ('decision', 'requirement', 'constraint', 'context', 'assumption', 'risk', 'fact'));

-- Episodic Events Table
CREATE TABLE IF NOT EXISTS episodic_events (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255),
    conversation_id VARCHAR(255) NOT NULL,
    time TIMESTAMP NOT NULL,
    text TEXT NOT NULL,
    kind VARCHAR(50) NOT NULL, -- 'interaction', 'decision', 'milestone'
    embedding vector(1024), -- BGE-M3 embeddings
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT episodic_event_kind_check CHECK (kind IN ('interaction', 'decision', 'milestone'))
);

-- Indexes for episodic_events
CREATE INDEX IF NOT EXISTS idx_episodic_events_user_id ON episodic_events(user_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_episodic_events_project_id ON episodic_events(project_id) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_episodic_events_conversation_id ON episodic_events(conversation_id);
CREATE INDEX IF NOT EXISTS idx_episodic_events_time ON episodic_events(time DESC) WHERE is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_episodic_events_kind ON episodic_events(kind) WHERE is_deleted = FALSE;

-- Vector similarity index
CREATE INDEX IF NOT EXISTS idx_episodic_events_embedding ON episodic_events 
USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100)
WHERE is_deleted = FALSE;

-- Conversation Items Table (for storing raw conversation history)
CREATE TABLE IF NOT EXISTS conversation_items (
    id VARCHAR(255) PRIMARY KEY,
    conversation_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL, -- 'user', 'assistant', 'system'
    content TEXT NOT NULL,
    tool_calls TEXT, -- JSON array of tool calls
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT conversation_item_role_check CHECK (role IN ('user', 'assistant', 'system'))
);

-- Indexes for conversation_items
CREATE INDEX IF NOT EXISTS idx_conversation_items_conversation_id ON conversation_items(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversation_items_created_at ON conversation_items(created_at ASC);

-- Comments for documentation
COMMENT ON TABLE user_memory_items IS 'User-specific memory items (preferences, context, core facts)';
COMMENT ON TABLE project_facts IS 'Project-level facts, decisions, and requirements';
COMMENT ON TABLE episodic_events IS 'Time-bound events and interactions';
COMMENT ON TABLE conversation_items IS 'Raw conversation history for memory extraction';

COMMENT ON COLUMN user_memory_items.embedding IS 'BGE-M3 1024-dimensional embedding vector';
COMMENT ON COLUMN project_facts.embedding IS 'BGE-M3 1024-dimensional embedding vector';
COMMENT ON COLUMN episodic_events.embedding IS 'BGE-M3 1024-dimensional embedding vector';

COMMENT ON COLUMN user_memory_items.score IS 'Importance level: 1=low, 2=medium, 3=normal, 4=high, 5=critical';
COMMENT ON COLUMN project_facts.confidence IS 'Confidence level: 0.0-1.0, higher is more confident';
