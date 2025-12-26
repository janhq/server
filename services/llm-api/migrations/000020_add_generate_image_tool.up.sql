-- Seed generate_image MCP tool into admin_mcp_tools
SET search_path TO llm_api;

INSERT INTO llm_api.admin_mcp_tools (
    public_id,
    tool_key,
    name,
    description,
    category,
    is_active,
    disallowed_keywords
)
VALUES (
    'mcp_generate_image_001',
    'generate_image',
    'generate_image',
    'Generate images from text prompts via /v1/images/generations (params: prompt, size, n, num_inference_steps, cfg_scale).',
    'image_generation',
    true,
    ARRAY[]::TEXT[]
)
ON CONFLICT (tool_key) DO NOTHING;
