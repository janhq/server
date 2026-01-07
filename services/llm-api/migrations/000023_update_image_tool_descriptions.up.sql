SET search_path TO llm_api;

UPDATE llm_api.admin_mcp_tools
SET description = 'Generate images from text prompts. Use when the user asks to create, generate, or make a new image from a text description.'
WHERE tool_key = 'generate_image';

UPDATE llm_api.admin_mcp_tools
SET description = 'Edit an existing image based on a text prompt. Use when the user wants to modify, change, or add elements to an existing image. Requires the image parameter with id or url of the image to edit.'
WHERE tool_key = 'edit_image';
