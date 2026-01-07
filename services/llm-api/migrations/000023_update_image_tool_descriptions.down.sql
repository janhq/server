SET search_path TO llm_api;

UPDATE llm_api.admin_mcp_tools
SET description = 'Generate images from text prompts (params: prompt, size, n, num_inference_steps, cfg_scale). Always keep the image url from tool_result, do not change or reformat as it is presigned url'
WHERE tool_key = 'generate_image';

UPDATE llm_api.admin_mcp_tools
SET description = 'Edit images with a prompt and input image (params: prompt, image, mask, size, strength, steps, seed, cfg_scale). Always keep the image url from tool_result, do not change or reformat as it is presigned url'
WHERE tool_key = 'edit_image';
