/**
 * Admin API Client
 * Handles all admin-related API calls for user and model management
 */

const JAN_BASE_URL = process.env.NEXT_PUBLIC_JAN_BASE_URL || 'http://localhost:8000';

// ============================================================================
// Types
// ============================================================================

export interface UserProfile {
  id: string;
  email?: string;
  username?: string;
  first_name?: string;
  last_name?: string;
  name?: string;
  picture?: string;
  object: string;
  role?: string;
  is_admin?: boolean;
  enabled?: boolean;
  active?: boolean;
  created_at?: string;
  updated_at?: string;
  groups?: Group[];
  roles?: string[];
}

export interface Group {
  id: string;
  name: string;
  path?: string;
  feature_flags?: string[];
  created_at?: string;
  updated_at?: string;
}

export interface FeatureFlag {
  id: string;
  key: string;
  name: string;
  description?: string;
  enabled?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface Provider {
  id: string; // Backend uses 'id' as the public_id
  name: string; // Backend uses 'name' not 'display_name'
  vendor: string; // Backend uses 'vendor' not 'kind'
  base_url?: string;
  endpoints?: Endpoint[];
  active: boolean;
  category?: string;
  default_provider_image_generate?: boolean;
  default_provider_image_edit?: boolean;
  model_count?: number; // Backend uses 'model_count'
  model_active_count?: number; // Backend uses 'model_active_count'
  metadata?: Record<string, any>;
  created_at?: string;
  updated_at?: string;
}

export interface Endpoint {
  url: string;
  weight?: number;
  priority?: number;
  healthy?: boolean;
}

export interface ProviderModel {
  id: string; // Backend uses 'id' as public_id
  provider_id: string; // Backend uses 'provider_id'
  model_display_name: string;
  model_id: string;
  model_public_id: string; // Public model identifier
  pricing: {
    prompt?: number;
    completion?: number;
    image?: number;
    request?: number;
  };
  category: string;
  category_order?: number; // Legacy field name
  model_order?: number; // Legacy field name
  category_order_number?: number; // New field name used by backend
  model_order_number?: number; // New field name used by backend
  active: boolean;
  created_at?: string;
  updated_at?: string;
  catalog?: ModelCatalog;
  supports_audio?: boolean;
  supports_embeddings?: boolean;
  supports_images?: boolean;
  supports_reasoning?: boolean;
  supports_instruct?: boolean; // From catalog - model can use an instruct backup
  supports_video?: boolean;
  supports_tools?: boolean;
  supports_browser?: boolean; // Model supports browser/web browsing functionality
  token_limits?: {
    context_length?: number;
    max_completion_tokens?: number;
  };
  instruct_model_public_id?: string; // Public ID of the instruct model to use when enable_thinking=false
}

export interface ModelCatalog {
  id: string; // Backend uses 'id' as public_id
  model_display_name?: string; // Display name of the model
  description?: string; // Model description
  supported_parameters?: any; // Backend returns object with names array and default object
  architecture?:
    | string
    | {
        modality?: string;
        input_modalities?: string[] | null;
        output_modalities?: string[] | null;
        tokenizer?: string;
        instruct_type?: string | null;
      };
  supports_images?: boolean;
  supports_embeddings?: boolean;
  supports_reasoning?: boolean;
  supports_instruct?: boolean; // Model can use an instruct backup (shows backup dropdown)
  supports_audio?: boolean;
  supports_video?: boolean;
  supports_tools?: boolean;
  supports_browser?: boolean; // Model supports browser/web browsing functionality
  family?: string;
  status?: string;
  is_moderated?: boolean;
  active?: boolean;
  experimental?: boolean;
  requires_feature_flag?: string | null;
  notes?: string;
  context_length?: number; // Maximum context length in tokens
  tags?: string[]; // Tags for categorization
  created_at?: string;
  updated_at?: string;
}

export interface PromptTemplate {
  id: string;
  public_id: string;
  name: string;
  description?: string;
  category: string;
  template_key: string;
  content: string;
  variables?: string[];
  metadata?: Record<string, any>;
  is_active: boolean;
  is_system: boolean;
  version: number;
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
}

export interface CreatePromptTemplateRequest {
  name: string;
  description?: string;
  category: string;
  template_key: string;
  content: string;
  variables?: string[];
  metadata?: Record<string, any>;
  is_active?: boolean;
}

export interface UpdatePromptTemplateRequest {
  name?: string;
  description?: string;
  category?: string;
  content?: string;
  variables?: string[];
  metadata?: Record<string, any>;
  is_active?: boolean;
}

export interface DuplicatePromptTemplateRequest {
  new_name?: string;
}

// ============================================================================
// Model Prompt Template Types (Model-Specific Template Assignments)
// ============================================================================

export interface ModelPromptTemplate {
  id: string;
  model_catalog_id: string;
  template_key: string;
  prompt_template_id: string;
  priority: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  prompt_template?: {
    public_id: string;
    name: string;
    description?: string;
    category: string;
    template_key: string;
    is_active: boolean;
  };
}

export interface AssignTemplateRequest {
  template_key: string;
  prompt_template_id: string;
  priority?: number;
  is_active?: boolean;
}

export interface UpdateAssignmentRequest {
  prompt_template_id?: string;
  priority?: number;
  is_active?: boolean;
}

export interface EffectiveTemplate {
  template: PromptTemplate | null;
  source: 'model_specific' | 'global_default' | 'hardcoded';
}

export interface EffectiveTemplatesResponse {
  templates: Record<string, EffectiveTemplate>;
}

// ============================================================================
// MCP Tool Types
// ============================================================================

export interface MCPTool {
  id: string;
  public_id: string;
  tool_key: string;
  name: string;
  description: string;
  category: string;
  is_active: boolean;
  metadata?: Record<string, any>;
  disallowed_keywords?: string[];
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
}

export interface UpdateMCPToolRequest {
  description?: string;
  category?: string;
  is_active?: boolean;
  metadata?: Record<string, any>;
  disallowed_keywords?: string[];
}

export interface ListResponse<T> {
  data: T[];
  total?: number;
  limit?: number;
  offset?: number;
}

export interface SyncProviderResponse {
  synced_models_count: number;
  message?: string;
}

export interface BatchUpdateResponse {
  updated_count: number;
  message?: string;
}

// ============================================================================
// Helper Functions
// ============================================================================

async function fetchWithAuth(url: string, options: RequestInit = {}, token?: string) {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(`API Error (${response.status}): ${error}`);
  }

  // Handle responses without body (204 No Content, or empty 200)
  if (response.status === 204) {
    return null;
  }

  // Check content-length to avoid parsing empty responses
  const contentLength = response.headers.get('content-length');
  if (contentLength === '0') {
    return null;
  }

  // Try to parse JSON, but handle empty responses
  const text = await response.text();
  if (!text || text.trim().length === 0) {
    return null;
  }

  return JSON.parse(text);
}

function buildQueryString(params: Record<string, any>): string {
  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      query.append(key, String(value));
    }
  });
  const queryString = query.toString();
  return queryString ? `?${queryString}` : '';
}

// ============================================================================
// User Management API
// ============================================================================

export class UserManagementAPI {
  constructor(private token?: string) {}

  /**
   * Get current user profile with admin status
   */
  async getMe(): Promise<UserProfile> {
    // Use versioned auth endpoint so request matches Kong's /v1 CORS route
    return fetchWithAuth(`${JAN_BASE_URL}/auth/me`, {}, this.token);
  }

  /**
   * List all users
   */
  async listUsers(params?: {
    limit?: number;
    offset?: number;
    search?: string;
    enabled?: boolean;
    exclude_guests?: boolean;
  }): Promise<ListResponse<UserProfile>> {
    const query = params ? buildQueryString(params) : '';
    return fetchWithAuth(`${JAN_BASE_URL}/v1/admin/users${query}`, {}, this.token);
  }

  /**
   * Create user
   */
  async createUser(data: {
    email: string;
    username: string;
    first_name?: string;
    last_name?: string;
    enabled?: boolean;
  }): Promise<UserProfile> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/users`,
      {
        method: 'POST',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Get user by ID
   */
  async getUser(userId: string): Promise<UserProfile> {
    return fetchWithAuth(`${JAN_BASE_URL}/v1/admin/users/${userId}`, {}, this.token);
  }

  /**
   * Update user
   */
  async updateUser(userId: string, data: Partial<UserProfile>): Promise<UserProfile> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/users/${userId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Deactivate user
   */
  async deactivateUser(userId: string): Promise<void> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/users/${userId}/deactivate`,
      {
        method: 'POST',
      },
      this.token,
    );
  }

  /**
   * Activate user
   */
  async activateUser(userId: string): Promise<void> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/users/${userId}/activate`,
      {
        method: 'POST',
      },
      this.token,
    );
  }

  /**
   * Assign admin role to user
   */
  async assignAdminRole(userId: string): Promise<void> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/users/${userId}/roles/admin`,
      {
        method: 'POST',
      },
      this.token,
    );
  }

  /**
   * List all groups
   */
  async listGroups(): Promise<ListResponse<Group>> {
    return fetchWithAuth(`${JAN_BASE_URL}/v1/admin/groups`, {}, this.token);
  }

  /**
   * Create group
   */
  async createGroup(name: string): Promise<Group> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/groups`,
      {
        method: 'POST',
        body: JSON.stringify({ name }),
      },
      this.token,
    );
  }

  /**
   * Delete group
   */
  async deleteGroup(groupId: string): Promise<void> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/groups/${groupId}`,
      {
        method: 'DELETE',
      },
      this.token,
    );
  }

  /**
   * Add user to group
   */
  async addUserToGroup(userId: string, groupId: string): Promise<void> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/users/${userId}/groups/${groupId}`,
      {
        method: 'POST',
      },
      this.token,
    );
  }

  /**
   * Remove user from group
   */
  async removeUserFromGroup(userId: string, groupId: string): Promise<void> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/users/${userId}/groups/${groupId}`,
      {
        method: 'DELETE',
      },
      this.token,
    );
  }

  /**
   * Get group feature flags
   */
  async getGroupFeatureFlags(groupId: string): Promise<ListResponse<FeatureFlag>> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/groups/${groupId}/feature-flags`,
      {},
      this.token,
    );
  }

  /**
   * Set group feature flags
   */
  async setGroupFeatureFlags(groupId: string, flags: string[]): Promise<void> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/groups/${groupId}/feature-flags`,
      {
        method: 'PATCH',
        body: JSON.stringify({ flags }),
      },
      this.token,
    );
  }

  /**
   * List all feature flags
   */
  async listFeatureFlags(): Promise<ListResponse<FeatureFlag>> {
    return fetchWithAuth(`${JAN_BASE_URL}/v1/admin/feature-flags`, {}, this.token);
  }

  /**
   * Create feature flag
   */
  async createFeatureFlag(data: {
    key: string;
    name: string;
    description?: string;
  }): Promise<FeatureFlag> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/feature-flags`,
      {
        method: 'POST',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Update feature flag
   */
  async updateFeatureFlag(flagId: string, data: Partial<FeatureFlag>): Promise<FeatureFlag> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/feature-flags/${flagId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Delete feature flag
   */
  async deleteFeatureFlag(flagId: string): Promise<void> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/feature-flags/${flagId}`,
      {
        method: 'DELETE',
      },
      this.token,
    );
  }
}

// ============================================================================
// Provider Management API
// ============================================================================

export class ProviderManagementAPI {
  constructor(private token?: string) {}

  /**
   * List all providers
   */
  async listProviders(params?: {
    limit?: number;
    offset?: number;
    kind?: string;
    active?: boolean;
  }): Promise<ListResponse<Provider>> {
    const query = params ? buildQueryString(params) : '';
    const response = await fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/providers${query}`,
      {},
      this.token,
    );

    // Backend returns plain array, transform to ListResponse format
    if (Array.isArray(response)) {
      return {
        data: response,
        total: response.length,
      };
    }
    return response;
  }

  /**
   * Get provider by public ID
   */
  async getProvider(publicId: string): Promise<Provider> {
    return fetchWithAuth(`${JAN_BASE_URL}/v1/admin/providers/${publicId}`, {}, this.token);
  }

  /**
   * Update provider
   */
  async updateProvider(
    publicId: string,
    data: Partial<Provider> & { url?: string; endpoints?: Endpoint[] },
  ): Promise<Provider> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/providers/${publicId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Sync provider models (not yet implemented on backend)
   */
  async syncProviderModels(
    publicId: string,
    autoEnableNewModels: boolean = false,
  ): Promise<SyncProviderResponse> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/providers/${publicId}/sync`,
      {
        method: 'POST',
        body: JSON.stringify({ auto_enable_new_models: autoEnableNewModels }),
      },
      this.token,
    );
  }

  /**
   * Create provider
   */
  async createProvider(data: {
    name: string;
    vendor: string;
    category?: string;
    base_url?: string;
    url?: string;
    endpoints?: Endpoint[];
    api_key?: string;
    metadata?: Record<string, string>;
    active?: boolean;
    default_provider_image_generate?: boolean;
    default_provider_image_edit?: boolean;
  }): Promise<Provider> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/providers`,
      {
        method: 'POST',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Delete provider by public ID
   */
  async deleteProvider(publicId: string): Promise<void> {
    await fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/providers/${publicId}`,
      {
        method: 'DELETE',
      },
      this.token,
    );
  }
}

// ============================================================================
// Provider Model Management API
// ============================================================================

export class ProviderModelManagementAPI {
  constructor(private token?: string) {}

  /**
   * List all provider models
   */
  async listProviderModels(params?: {
    limit?: number;
    offset?: number;
    provider_id?: string; // Now public ID
    search?: string; // Search term for filtering models
    active?: boolean;
    supports_images?: boolean;
  }): Promise<ListResponse<ProviderModel>> {
    const query = params ? buildQueryString(params) : '';
    return fetchWithAuth(`${JAN_BASE_URL}/v1/admin/models/provider-models${query}`, {}, this.token);
  }

  /**
   * Get provider model by public ID
   */
  async getProviderModel(publicId: string): Promise<ProviderModel> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/models/provider-models/${publicId}`,
      {},
      this.token,
    );
  }

  /**
   * Update provider model
   */
  async updateProviderModel(
    publicId: string,
    data: Partial<ProviderModel>,
  ): Promise<ProviderModel> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/models/provider-models/${publicId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Activate model
   */
  async activateModel(publicId: string): Promise<ProviderModel> {
    return this.updateProviderModel(publicId, { active: true });
  }

  /**
   * Deactivate model
   */
  async deactivateModel(publicId: string): Promise<ProviderModel> {
    return this.updateProviderModel(publicId, { active: false });
  }

  /**
   * Batch update active status
   */
  async batchUpdateActive(params: {
    enable: boolean;
    provider_id?: string;
    except_models?: string[];
  }): Promise<BatchUpdateResponse> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/models/provider-models/bulk-toggle`,
      {
        method: 'POST',
        body: JSON.stringify(params),
      },
      this.token,
    );
  }
}

// ============================================================================
// Model Catalog Management API
// ============================================================================

export class ModelCatalogManagementAPI {
  constructor(private token?: string) {}

  /**
   * List all model catalogs
   */
  async listModelCatalogs(params?: {
    limit?: number;
    offset?: number;
    family?: string;
    status?: string;
    is_moderated?: boolean;
    active?: boolean;
    supports_embeddings?: boolean;
    supports_images?: boolean;
    supports_reasoning?: boolean;
    supports_audio?: boolean;
    supports_video?: boolean;
    supports_tools?: boolean;
    supports_browser?: boolean;
    experimental?: boolean;
    requires_feature_flag?: string;
  }): Promise<ListResponse<ModelCatalog>> {
    const query = params ? buildQueryString(params) : '';
    return fetchWithAuth(`${JAN_BASE_URL}/v1/admin/models/catalogs${query}`, {}, this.token);
  }

  /**
   * Get model catalog by public ID (public IDs can contain slashes)
   */
  async getModelCatalog(publicId: string): Promise<ModelCatalog> {
    return fetchWithAuth(`${JAN_BASE_URL}/v1/admin/models/catalogs/${publicId}`, {}, this.token);
  }

  /**
   * Update model catalog
   */
  async updateModelCatalog(publicId: string, data: Partial<ModelCatalog>): Promise<ModelCatalog> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/models/catalogs/${publicId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Batch toggle catalog-linked provider models
   */
  async batchToggle(params: {
    enable: boolean;
    catalog_ids?: string[];
  }): Promise<BatchUpdateResponse> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/models/catalogs/bulk-toggle`,
      {
        method: 'POST',
        body: JSON.stringify(params),
      },
      this.token,
    );
  }

  // ============================================================================
  // Model Prompt Template Methods (Model-Specific Template Assignments)
  // ============================================================================

  /**
   * List prompt template assignments for a model catalog
   */
  async listModelPromptTemplates(modelId: string): Promise<ListResponse<ModelPromptTemplate>> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/models/prompt-templates/list/${modelId}`,
      {},
      this.token,
    );
  }

  /**
   * Assign a prompt template to a model catalog
   * If an assignment already exists for the template_key, it will be updated
   */
  async assignPromptTemplate(
    modelId: string,
    data: AssignTemplateRequest,
  ): Promise<ModelPromptTemplate> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/models/prompt-templates/assign/${modelId}`,
      {
        method: 'POST',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Update an existing prompt template assignment
   */
  async updatePromptTemplateAssignment(
    modelId: string,
    templateKey: string,
    data: UpdateAssignmentRequest,
  ): Promise<ModelPromptTemplate> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/models/prompt-templates/update/${encodeURIComponent(templateKey)}/${modelId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Remove a prompt template assignment from a model catalog
   * The model will revert to using the global default template
   */
  async unassignPromptTemplate(modelId: string, templateKey: string): Promise<void> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/models/prompt-templates/unassign/${encodeURIComponent(templateKey)}/${modelId}`,
      { method: 'DELETE' },
      this.token,
    );
  }

  /**
   * Get effective templates for a model catalog (resolved with fallbacks)
   * Returns all template keys with their resolved templates and sources
   */
  async getEffectiveTemplates(modelId: string): Promise<EffectiveTemplatesResponse> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/models/prompt-templates/effective/${modelId}`,
      {},
      this.token,
    );
  }
}

// ============================================================================
// Prompt Template Management API
// ============================================================================

export class PromptTemplateManagementAPI {
  constructor(private token?: string) {}

  /**
   * List all prompt templates
   */
  async listPromptTemplates(params?: {
    limit?: number;
    offset?: number;
    category?: string;
    is_active?: boolean;
    is_system?: boolean;
    search?: string;
  }): Promise<ListResponse<PromptTemplate>> {
    const query = params ? buildQueryString(params) : '';
    return fetchWithAuth(`${JAN_BASE_URL}/v1/admin/prompt-templates${query}`, {}, this.token);
  }

  /**
   * Get prompt template by public ID
   */
  async getPromptTemplate(publicId: string): Promise<PromptTemplate> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/prompt-templates/${publicId}`,
      {},
      this.token,
    );
  }

  /**
   * Get prompt template by key (public endpoint)
   */
  async getPromptTemplateByKey(templateKey: string): Promise<PromptTemplate> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/prompt-templates/${templateKey}`,
      {},
      this.token,
    );
  }

  /**
   * Create new prompt template
   */
  async createPromptTemplate(data: CreatePromptTemplateRequest): Promise<PromptTemplate> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/prompt-templates`,
      {
        method: 'POST',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Update prompt template
   */
  async updatePromptTemplate(
    publicId: string,
    data: UpdatePromptTemplateRequest,
  ): Promise<PromptTemplate> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/prompt-templates/${publicId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Delete prompt template (only non-system templates)
   */
  async deletePromptTemplate(publicId: string): Promise<void> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/prompt-templates/${publicId}`,
      {
        method: 'DELETE',
      },
      this.token,
    );
  }

  /**
   * Duplicate prompt template
   */
  async duplicatePromptTemplate(
    publicId: string,
    data: DuplicatePromptTemplateRequest,
  ): Promise<PromptTemplate> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/prompt-templates/${publicId}/duplicate`,
      {
        method: 'POST',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Activate template
   */
  async activateTemplate(publicId: string): Promise<PromptTemplate> {
    return this.updatePromptTemplate(publicId, { is_active: true });
  }

  /**
   * Deactivate template
   */
  async deactivateTemplate(publicId: string): Promise<PromptTemplate> {
    return this.updatePromptTemplate(publicId, { is_active: false });
  }
}

// ============================================================================
// MCP Tool Management API
// ============================================================================

export class MCPToolManagementAPI {
  constructor(private token?: string) {}

  /**
   * List all MCP tools
   */
  async listMCPTools(params?: {
    limit?: number;
    offset?: number;
    category?: string;
    is_active?: boolean;
    search?: string;
  }): Promise<ListResponse<MCPTool>> {
    const query = params ? buildQueryString(params) : '';
    return fetchWithAuth(`${JAN_BASE_URL}/v1/admin/mcp-tools${query}`, {}, this.token);
  }

  /**
   * Get MCP tool by public ID
   */
  async getMCPTool(publicId: string): Promise<MCPTool> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/mcp-tools/${publicId}`,
      {},
      this.token,
    );
  }

  /**
   * Update MCP tool (Note: name is read-only)
   */
  async updateMCPTool(
    publicId: string,
    data: UpdateMCPToolRequest,
  ): Promise<MCPTool> {
    return fetchWithAuth(
      `${JAN_BASE_URL}/v1/admin/mcp-tools/${publicId}`,
      {
        method: 'PATCH',
        body: JSON.stringify(data),
      },
      this.token,
    );
  }

  /**
   * Activate tool
   */
  async activateTool(publicId: string): Promise<MCPTool> {
    return this.updateMCPTool(publicId, { is_active: true });
  }

  /**
   * Deactivate tool
   */
  async deactivateTool(publicId: string): Promise<MCPTool> {
    return this.updateMCPTool(publicId, { is_active: false });
  }
}

// ============================================================================
// Main Admin API Client
// ============================================================================

export class AdminAPIClient {
  public users: UserManagementAPI;
  public providers: ProviderManagementAPI;
  public providerModels: ProviderModelManagementAPI;
  public modelCatalogs: ModelCatalogManagementAPI;
  public promptTemplates: PromptTemplateManagementAPI;
  public mcpTools: MCPToolManagementAPI;

  constructor(token?: string) {
    this.users = new UserManagementAPI(token);
    this.providers = new ProviderManagementAPI(token);
    this.providerModels = new ProviderModelManagementAPI(token);
    this.modelCatalogs = new ModelCatalogManagementAPI(token);
    this.promptTemplates = new PromptTemplateManagementAPI(token);
    this.mcpTools = new MCPToolManagementAPI(token);
  }

  /**
   * Check if current user is admin
   */
  async checkIsAdmin(): Promise<boolean> {
    try {
      const profile = await this.users.getMe();
      return (
        profile.is_admin === true ||
        profile.role === 'admin' ||
        (profile.roles && profile.roles.includes('admin')) ||
        false
      );
    } catch (error) {
      console.error('Failed to check admin status:', error);
      return false;
    }
  }
}

/**
 * Create admin API client with token
 */
export function createAdminAPIClient(token?: string): AdminAPIClient {
  return new AdminAPIClient(token);
}
