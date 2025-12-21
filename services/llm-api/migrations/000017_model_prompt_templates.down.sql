-- Rollback Model-Specific Prompt Templates
-- Migration 17 Down: Drop model_prompt_templates table

SET search_path TO llm_api;

DROP TABLE IF EXISTS llm_api.model_prompt_templates;
