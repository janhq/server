'use client';

import { createAdminAPIClient, Endpoint, Provider } from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
import {
  Activity,
  Box,
  ChevronLeft,
  ChevronRight,
  Database,
  Edit2,
  Loader2,
  Plus,
  RefreshCw,
  Trash,
  X,
} from 'lucide-react';
import { useEffect, useState } from 'react';

export default function ProvidersManagementPage() {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [kindFilter, setKindFilter] = useState<string>('');
  const [activeFilter, setActiveFilter] = useState<boolean | undefined>(undefined);
  const [page, setPage] = useState(0);
  const [totalProviders, setTotalProviders] = useState(0);
  const [syncingProvider, setSyncingProvider] = useState<string | null>(null);
  const [deletingProvider, setDeletingProvider] = useState<string | null>(null);
  const [selectedProvider, setSelectedProvider] = useState<Provider | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const limit = 20;

  useEffect(() => {
    loadProviders();
  }, [page, kindFilter, activeFilter]);

  async function loadProviders() {
    try {
      setIsLoading(true);
      setError(null);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) {
        setError('No authentication token found');
        return;
      }

      const adminClient = createAdminAPIClient(token);
      const response = await adminClient.providers.listProviders({
        limit,
        offset: page * limit,
        kind: kindFilter || undefined,
        active: activeFilter,
      });

      // Backend already returns providers with counts
      setProviders(response.data);
      setTotalProviders(response.total || 0);
    } catch (err) {
      console.error('Failed to load providers:', err);
      setError('Failed to load providers');
    } finally {
      setIsLoading(false);
    }
  }

  async function handleSyncProvider(providerId: string, autoEnable: boolean = false) {
    if (!confirm('Are you sure you want to sync models for this provider?')) {
      return;
    }

    try {
      setSyncingProvider(providerId);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      const result = await adminClient.providers.syncProviderModels(providerId, autoEnable);

      alert(`Successfully synced ${result.synced_models_count} models`);
      loadProviders();
    } catch (err) {
      console.error('Failed to sync provider:', err);
      alert('Failed to sync provider models');
    } finally {
      setSyncingProvider(null);
    }
  }

  async function handleDeleteProvider(provider: Provider) {
    const confirmation = prompt(
      `Type "Approved" to delete provider "${provider.name}". This will also remove all provider models.`,
    );

    if (confirmation !== 'Approved') {
      alert('Deletion cancelled. You must type "Approved" exactly to proceed.');
      return;
    }

    try {
      setDeletingProvider(provider.id);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.providers.deleteProvider(provider.id);

      alert(`Provider "${provider.name}" has been deleted.`);
      loadProviders();
    } catch (err) {
      console.error('Failed to delete provider:', err);
      alert('Failed to delete provider');
    } finally {
      setDeletingProvider(null);
    }
  }

  function parseEndpointsInput(raw: string): Endpoint[] {
    const trimmed = raw?.trim();
    if (!trimmed) return [];

    // Try JSON array first
    if (trimmed.startsWith('[')) {
      const parsed = JSON.parse(trimmed);
      if (Array.isArray(parsed)) {
        return parsed
          .map((item) => {
            if (typeof item === 'string') {
              const url = item.trim();
              return url ? { url } : null;
            }
            if (item && typeof item.url === 'string') {
              const url = item.url.trim();
              if (!url) return null;
              return {
                url,
                weight: typeof item.weight === 'number' ? item.weight : undefined,
                priority: typeof item.priority === 'number' ? item.priority : undefined,
              } as Endpoint;
            }
            return null;
          })
          .filter((ep): ep is Endpoint => !!ep);
      }
      throw new Error('Endpoints JSON must be an array');
    }

    // Fallback: split on commas or newlines
    return trimmed
      .split(/[\n,]/)
      .map((part) => part.trim())
      .filter(Boolean)
      .map((url) => ({ url }));
  }

  async function handleCreateProvider(data: {
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
  }) {
    try {
      setIsSubmitting(true);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) {
        alert('Authentication token missing. Please sign in again.');
        return;
      }

      const adminClient = createAdminAPIClient(token);
      await adminClient.providers.createProvider(data);

      setShowCreateModal(false);
      loadProviders();
    } catch (err) {
      console.error('Failed to create provider:', err);
      alert('Failed to create provider');
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleEditProvider(data: {
    name?: string;
    vendor?: string;
    category?: string;
    base_url?: string;
    url?: string;
    endpoints?: Endpoint[];
    active?: boolean;
    default_provider_image_generate?: boolean;
    default_provider_image_edit?: boolean;
    metadata?: Record<string, string>;
  }) {
    if (!selectedProvider) return;

    try {
      setIsSubmitting(true);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) {
        alert('Authentication token missing. Please sign in again.');
        return;
      }

      const adminClient = createAdminAPIClient(token);
      await adminClient.providers.updateProvider(selectedProvider.id, data);

      setShowEditModal(false);
      setSelectedProvider(null);
      loadProviders();
    } catch (err) {
      console.error('Failed to update provider:', err);
      alert('Failed to update provider');
    } finally {
      setIsSubmitting(false);
    }
  }

  const totalPages = Math.ceil(totalProviders / limit);

  // Get unique kinds for filter
  const availableKinds = ['openrouter', 'openai', 'anthropic', 'google', 'local'];

  function normalizeMetadataValue(value: unknown): string | undefined {
    if (value === null || value === undefined) return undefined;
    if (typeof value === 'string') return value;
    try {
      return JSON.stringify(value);
    } catch {
      return String(value);
    }
  }

  function mergeMetadata(
    base: Record<string, any> | undefined,
    updates: { imageEditPath?: string },
  ): Record<string, string> | undefined {
    const merged: Record<string, string> = {};
    if (base) {
      Object.entries(base).forEach(([key, value]) => {
        const normalized = normalizeMetadataValue(value);
        if (normalized) merged[key] = normalized;
      });
    }

    if (updates.imageEditPath !== undefined) {
      const trimmed = updates.imageEditPath.trim();
      if (trimmed) {
        merged.image_edit_path = trimmed;
      } else {
        delete merged.image_edit_path;
      }
    }

    return Object.keys(merged).length > 0 ? merged : undefined;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Database className="w-8 h-8" />
            Model Providers
          </h1>
          <p className="text-muted-foreground mt-2">
            Manage model providers and sync their available models
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
        >
          <Plus className="w-4 h-4" />
          Add Provider
        </button>
      </div>

      {/* Filters */}
      <div className="bg-card rounded-lg border p-4">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Kind Filter */}
          <select
            value={kindFilter}
            onChange={(e) => {
              setKindFilter(e.target.value);
              setPage(0);
            }}
            className="px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <option value="">All Provider Types</option>
            {availableKinds.map((kind) => (
              <option key={kind} value={kind}>
                {kind.charAt(0).toUpperCase() + kind.slice(1)}
              </option>
            ))}
          </select>

          {/* Active Filter */}
          <select
            value={activeFilter === undefined ? 'all' : activeFilter ? 'active' : 'inactive'}
            onChange={(e) => {
              setActiveFilter(e.target.value === 'all' ? undefined : e.target.value === 'active');
              setPage(0);
            }}
            className="px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <option value="all">All Providers</option>
            <option value="active">Active Only</option>
            <option value="inactive">Inactive Only</option>
          </select>

          {/* Clear Filters */}
          {(kindFilter || activeFilter !== undefined) && (
            <button
              onClick={() => {
                setKindFilter('');
                setActiveFilter(undefined);
                setPage(0);
              }}
              className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground border rounded-md hover:bg-accent transition-colors"
            >
              Clear Filters
            </button>
          )}
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      {/* Providers Grid */}
      <div className="space-y-4">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-6 h-6 animate-spin text-primary" />
          </div>
        ) : providers.length === 0 ? (
          <div className="text-center py-12 bg-card rounded-lg border">
            <Database className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No providers found</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {providers.map((provider) => (
              <div
                key={provider.id}
                className={`bg-card rounded-lg border p-6 hover:shadow-md transition-shadow ${
                  !provider.active ? 'opacity-60 bg-muted/30 grayscale-[0.5]' : ''
                }`}
              >
                {/* Provider Header */}
                <div className="flex items-start justify-between mb-4">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <Database className="w-5 h-5 text-primary" />
                      <h3 className="font-semibold text-lg">{provider.name}</h3>
                    </div>
                    <p className="text-sm text-muted-foreground">{provider.vendor}</p>
                  </div>
                  <div>
                    <span
                      className={`px-2 py-1 rounded-full text-xs font-medium border ${
                        provider.active
                          ? 'bg-green-100 text-green-700 border-green-200 dark:bg-green-900/30 dark:text-green-400 dark:border-green-800'
                          : 'bg-gray-100 text-gray-700 border-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700'
                      }`}
                    >
                      {provider.active ? 'Active' : 'Inactive'}
                    </span>
                  </div>
                </div>

                {/* Provider Stats */}
                <div className="grid grid-cols-2 gap-3 mb-4">
                  <div className="bg-muted/50 rounded-md p-3">
                    <div className="flex items-center gap-2 mb-1">
                      <Box className="w-4 h-4 text-muted-foreground" />
                      <span className="text-xs text-muted-foreground">Total Models</span>
                    </div>
                    <p className="text-2xl font-bold">
                      {provider.model_count !== undefined ? provider.model_count : '-'}
                    </p>
                  </div>
                  <div className="bg-muted/50 rounded-md p-3">
                    <div className="flex items-center gap-2 mb-1">
                      <Activity className="w-4 h-4 text-green-600" />
                      <span className="text-xs text-muted-foreground">Active</span>
                    </div>
                    <p className="text-2xl font-bold text-green-600">
                      {provider.model_active_count !== undefined
                        ? provider.model_active_count
                        : '-'}
                    </p>
                  </div>
                </div>

                {/* Provider ID */}
                <div className="mb-4">
                  <p className="text-xs text-muted-foreground mb-1">Provider ID</p>
                  <code className="text-xs bg-muted px-2 py-1 rounded">{provider.id}</code>
                </div>

                <div className="mb-4 space-y-1">
                  <p className="text-xs text-muted-foreground">Endpoints</p>
                  {provider.endpoints && provider.endpoints.length > 0 ? (
                    <div className="space-y-1">
                      {provider.endpoints.slice(0, 3).map((ep, idx) => (
                        <div key={ep.url + idx} className="text-xs bg-muted px-2 py-1 rounded">
                          {ep.url}
                        </div>
                      ))}
                      {provider.endpoints.length > 3 && (
                        <p className="text-xs text-muted-foreground">
                          +{provider.endpoints.length - 3} more
                        </p>
                      )}
                    </div>
                  ) : provider.base_url ? (
                    <div className="text-xs bg-muted px-2 py-1 rounded">{provider.base_url}</div>
                  ) : (
                    <p className="text-xs text-muted-foreground">Not configured</p>
                  )}
                </div>

                {/* Actions */}
                <div className="flex gap-2">
                  <button
                    onClick={() => {
                      setSelectedProvider(provider);
                      setShowEditModal(true);
                    }}
                    className="flex items-center justify-center gap-2 px-3 py-2 border rounded-md hover:bg-accent transition-colors text-sm font-medium"
                    title="Edit provider"
                  >
                    <Edit2 className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => handleSyncProvider(provider.id, false)}
                    disabled={syncingProvider === provider.id}
                    className="flex-1 flex items-center justify-center gap-2 px-3 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-sm font-medium"
                  >
                    {syncingProvider === provider.id ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Syncing...
                      </>
                    ) : (
                      <>
                        <RefreshCw className="w-4 h-4" />
                        Sync Models
                      </>
                    )}
                  </button>
                  <button
                    onClick={() => handleDeleteProvider(provider)}
                    disabled={deletingProvider === provider.id}
                    className="flex items-center justify-center gap-2 px-3 py-2 border rounded-md text-sm font-medium hover:bg-destructive/10 hover:text-destructive transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    title="Delete provider"
                  >
                    {deletingProvider === provider.id ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Deleting...
                      </>
                    ) : (
                      <>
                        <Trash className="w-4 h-4" />
                        Delete
                      </>
                    )}
                  </button>
                </div>

                {/* Timestamps */}
                {provider.updated_at && (
                  <div className="mt-3 pt-3 border-t">
                    <p className="text-xs text-muted-foreground">
                      Updated: {new Date(provider.updated_at).toLocaleDateString()}
                    </p>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Showing {page * limit + 1} to {Math.min((page + 1) * limit, totalProviders)} of{' '}
            {totalProviders} providers
          </p>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage(Math.max(0, page - 1))}
              disabled={page === 0}
              className="p-2 border rounded-md hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span className="text-sm">
              Page {page + 1} of {totalPages}
            </span>
            <button
              onClick={() => setPage(Math.min(totalPages - 1, page + 1))}
              disabled={page >= totalPages - 1}
              className="p-2 border rounded-md hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      {/* Create Provider Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-card rounded-lg border max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Add New Provider</h2>
              <button
                onClick={() => setShowCreateModal(false)}
                className="p-1 hover:bg-accent rounded-md"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form
              onSubmit={(e) => {
                e.preventDefault();
                const formData = new FormData(e.currentTarget);

                // Construct metadata object
                const metadata: Record<string, string> = {};
                const description = formData.get('metadata_description') as string;
                const environment = formData.get('metadata_environment') as string;
                const imageEditPath = formData.get('metadata_image_edit_path') as string;
                const autoEnable = formData.get('metadata_auto_enable') === 'on';

                if (description) metadata.description = description;
                if (environment) metadata.environment = environment;
                if (imageEditPath) metadata.image_edit_path = imageEditPath;
                metadata.auto_enable_new_models = autoEnable ? 'true' : 'false';

                const endpointsRaw = (formData.get('endpoints') as string) || '';
                let endpoints: Endpoint[] = [];
                try {
                  endpoints = parseEndpointsInput(endpointsRaw);
                } catch (err) {
                  console.error(err);
                  alert('Invalid endpoints format. Use comma-separated URLs or a JSON array.');
                  return;
                }

                const baseUrl = (formData.get('base_url') as string) || '';
                if (!baseUrl && endpoints.length === 0) {
                  alert('Please provide either a Base URL or one or more endpoints.');
                  return;
                }

                handleCreateProvider({
                  name: formData.get('name') as string,
                  vendor: formData.get('vendor') as string,
                  category: (formData.get('category') as string) || undefined,
                  base_url: baseUrl || undefined,
                  endpoints: endpoints.length ? endpoints : undefined,
                  api_key: (formData.get('api_key') as string) || undefined,
                  active: formData.get('active') === 'on',
                  default_provider_image_generate:
                    formData.get('default_provider_image_generate') === 'on',
                  default_provider_image_edit: formData.get('default_provider_image_edit') === 'on',
                  metadata,
                });
              }}
              className="space-y-4"
            >
              <div className="grid grid-cols-2 gap-4">
                <div className="col-span-2">
                  <label className="block text-sm font-medium mb-1">Provider Name *</label>
                  <input
                    type="text"
                    name="name"
                    required
                    placeholder="e.g. Local Jan vLLM Provider"
                    className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Vendor *</label>
                  <input
                    type="text"
                    name="vendor"
                    required
                    placeholder="e.g. custom"
                    className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Category</label>
                  <select
                    name="category"
                    defaultValue="llm"
                    className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                  >
                    <option value="llm">LLM</option>
                    <option value="image">Image</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Environment</label>
                  <input
                    type="text"
                    name="metadata_environment"
                    placeholder="e.g. production"
                    className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>

                <div className="col-span-2">
                  <label className="block text-sm font-medium mb-1">
                    Endpoints (comma-separated or JSON array)
                  </label>
                  <textarea
                    name="endpoints"
                    rows={2}
                    placeholder="http://vllm-1:8101/v1, http://vllm-2:8101/v1"
                    className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary resize-none"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Endpoints take precedence when provided. Leave empty to fall back to Base URL.
                  </p>
                </div>

                <div className="col-span-2">
                  <label className="block text-sm font-medium mb-1">Base URL</label>
                  <input
                    type="url"
                    name="base_url"
                    placeholder="http://10.200.108.71:8080/v1"
                    className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Optional when endpoints are provided. Used for legacy compatibility.
                  </p>
                </div>

                <div className="col-span-2">
                  <label className="block text-sm font-medium mb-1">Image Edit Path (Optional)</label>
                  <input
                    type="text"
                    name="metadata_image_edit_path"
                    placeholder="/v1/images/edits or full URL"
                    className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Overrides the edit endpoint for this provider.
                  </p>
                </div>

                <div className="col-span-2">
                  <label className="block text-sm font-medium mb-1">API Key (Optional)</label>
                  <input
                    type="password"
                    name="api_key"
                    placeholder="sk-..."
                    className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                  />
                </div>

                <div className="col-span-2">
                  <label className="block text-sm font-medium mb-1">Description</label>
                  <textarea
                    name="metadata_description"
                    rows={2}
                    placeholder="Provider description..."
                    className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary resize-none"
                  />
                </div>
              </div>

              <div className="flex flex-col gap-3 pt-2 border-t">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    name="active"
                    defaultChecked
                    className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <span className="text-sm font-medium">Active</span>
                  <span className="text-xs text-muted-foreground ml-auto">
                    Enable this provider immediately
                  </span>
                </label>

                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    name="default_provider_image_generate"
                    className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <span className="text-sm font-medium">Default for Image Generate</span>
                  <span className="text-xs text-muted-foreground ml-auto">
                    Use as default for /v1/images/generations
                  </span>
                </label>

                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    name="default_provider_image_edit"
                    className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <span className="text-sm font-medium">Default for Image Edit</span>
                  <span className="text-xs text-muted-foreground ml-auto">
                    Use as default for /v1/images/edits
                  </span>
                </label>

                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    name="metadata_auto_enable"
                    defaultChecked
                    className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <span className="text-sm font-medium">Auto-enable New Models</span>
                  <span className="text-xs text-muted-foreground ml-auto">
                    Automatically enable new models found during sync
                  </span>
                </label>
              </div>

              <div className="flex justify-end gap-2 pt-4">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 border rounded-md hover:bg-accent transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50 flex items-center gap-2"
                >
                  {isSubmitting && <Loader2 className="w-4 h-4 animate-spin" />}
                  Create Provider
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Provider Modal */}
      {showEditModal && selectedProvider && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-card rounded-lg border max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Edit Provider</h2>
              <button
                onClick={() => {
                  setShowEditModal(false);
                  setSelectedProvider(null);
                }}
                className="p-1 hover:bg-accent rounded-md"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form
              onSubmit={(e) => {
                e.preventDefault();
                const formData = new FormData(e.currentTarget);
                const endpointsRaw = (formData.get('endpoints') as string) || '';
                let endpoints: Endpoint[] | undefined;
                try {
                  const parsed = parseEndpointsInput(endpointsRaw);
                  endpoints = parsed.length ? parsed : undefined;
                } catch (err) {
                  console.error(err);
                  alert('Invalid endpoints format. Use comma-separated URLs or a JSON array.');
                  return;
                }

                const baseUrl = (formData.get('base_url') as string) || '';
                if (!baseUrl && !endpoints) {
                  alert('Please provide either a Base URL or one or more endpoints.');
                  return;
                }
                const imageEditPath = (formData.get('metadata_image_edit_path') as string) || '';
                const metadata = mergeMetadata(selectedProvider.metadata, {
                  imageEditPath,
                });

                handleEditProvider({
                  name: formData.get('name') as string,
                  vendor: formData.get('vendor') as string,
                  category: (formData.get('category') as string) || undefined,
                  base_url: baseUrl || undefined,
                  endpoints,
                  active: formData.get('active') === 'on',
                  default_provider_image_generate:
                    formData.get('default_provider_image_generate') === 'on',
                  default_provider_image_edit: formData.get('default_provider_image_edit') === 'on',
                  metadata,
                });
              }}
              className="space-y-4"
            >
              <div>
                <label className="block text-sm font-medium mb-1">Provider Name</label>
                <input
                  type="text"
                  name="name"
                  defaultValue={selectedProvider.name}
                  className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Vendor</label>
                <input
                  type="text"
                  name="vendor"
                  defaultValue={selectedProvider.vendor}
                  className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Category</label>
                <select
                  name="category"
                  defaultValue={selectedProvider.category || 'llm'}
                  className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="llm">LLM</option>
                  <option value="image">Image</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Base URL</label>
                <input
                  type="url"
                  name="base_url"
                  defaultValue={selectedProvider.base_url || ''}
                  placeholder="https://api.example.com"
                  className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Optional when endpoints are provided.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">
                  Endpoints (comma-separated or JSON array)
                </label>
                <textarea
                  name="endpoints"
                  defaultValue={
                    selectedProvider.endpoints && selectedProvider.endpoints.length > 0
                      ? selectedProvider.endpoints.map((ep) => ep.url).join('\n')
                      : ''
                  }
                  rows={2}
                  placeholder="http://vllm-1:8101/v1, http://vllm-2:8101/v1"
                  className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary resize-none"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Leave blank to keep existing endpoints or use Base URL fallback.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Image Edit Path (Optional)</label>
                <input
                  type="text"
                  name="metadata_image_edit_path"
                  defaultValue={
                    typeof selectedProvider.metadata?.image_edit_path === 'string'
                      ? selectedProvider.metadata?.image_edit_path
                      : ''
                  }
                  placeholder="/v1/images/edits or full URL"
                  className="w-full px-3 py-2 border rounded-md bg-background focus:outline-none focus:ring-2 focus:ring-primary"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Overrides the edit endpoint for this provider.
                </p>
              </div>

              <div className="flex flex-col gap-3 pt-2 border-t">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    name="active"
                    defaultChecked={selectedProvider.active}
                    className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <span className="text-sm font-medium">Active</span>
                  <span className="text-xs text-muted-foreground ml-auto">
                    Enable this provider
                  </span>
                </label>

                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    name="default_provider_image_generate"
                    defaultChecked={!!selectedProvider.default_provider_image_generate}
                    className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <span className="text-sm font-medium">Default for Image Generate</span>
                  <span className="text-xs text-muted-foreground ml-auto">
                    Use as default for /v1/images/generations
                  </span>
                </label>

                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    name="default_provider_image_edit"
                    defaultChecked={!!selectedProvider.default_provider_image_edit}
                    className="w-4 h-4 rounded border-gray-300 text-primary focus:ring-primary"
                  />
                  <span className="text-sm font-medium">Default for Image Edit</span>
                  <span className="text-xs text-muted-foreground ml-auto">
                    Use as default for /v1/images/edits
                  </span>
                </label>
              </div>

              <div className="bg-muted/50 rounded-md p-3">
                <p className="text-xs text-muted-foreground">
                  <strong>Provider ID:</strong> {selectedProvider.id}
                </p>
              </div>

              <div className="flex justify-end gap-2 pt-4">
                <button
                  type="button"
                  onClick={() => {
                    setShowEditModal(false);
                    setSelectedProvider(null);
                  }}
                  className="px-4 py-2 border rounded-md hover:bg-accent transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50 flex items-center gap-2"
                >
                  {isSubmitting && <Loader2 className="w-4 h-4 animate-spin" />}
                  Save Changes
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
