SET search_path TO llm_api;

DROP INDEX IF EXISTS llm_api.idx_model_catalogs_experimental;
ALTER TABLE llm_api.model_catalogs DROP COLUMN IF EXISTS experimental;

DROP INDEX IF EXISTS llm_api.idx_audit_logs_resource_type;
DROP INDEX IF EXISTS llm_api.idx_audit_logs_admin_user_id;
DROP INDEX IF EXISTS llm_api.idx_audit_logs_created_at;
DROP TABLE IF EXISTS llm_api.audit_logs;

DROP INDEX IF EXISTS llm_api.idx_feature_flags_category;
DROP INDEX IF EXISTS llm_api.idx_feature_flags_key;
DROP TABLE IF EXISTS llm_api.feature_flags;
