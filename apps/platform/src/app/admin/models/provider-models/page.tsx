'use client';

import { createAdminAPIClient, Provider, ProviderModel } from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
import {
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  Edit2,
  ExternalLink,
  Loader2,
  Search,
  X,
  XCircle,
} from 'lucide-react';
import Link from 'next/link';
import { useEffect, useState } from 'react';

export default function ProviderModelsManagementPage() {
  const [models, setModels] = useState<ProviderModel[]>([]);
  const [allModels, setAllModels] = useState<ProviderModel[]>([]); // All models for instruct model dropdown
  const [providers, setProviders] = useState<Provider[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [providerFilter, setProviderFilter] = useState<string | undefined>(undefined);
  const [activeFilter, setActiveFilter] = useState<boolean | undefined>(undefined);
  const [supportsImagesFilter, setSupportsImagesFilter] = useState<boolean | undefined>(undefined);
  const [page, setPage] = useState(0);
  const [totalModels, setTotalModels] = useState(0);
  const [selectedModels, setSelectedModels] = useState<Set<string>>(new Set());
  const [selectedModelForEdit, setSelectedModelForEdit] = useState<ProviderModel | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [isPerformingBulkAction, setIsPerformingBulkAction] = useState(false);
  const [selectedInstructModel, setSelectedInstructModel] = useState<string>(''); // For edit modal

  const limit = 20;

  useEffect(() => {
    loadProviders();
    loadAllModels(); // Load all models for instruct model dropdown
  }, []);

  useEffect(() => {
    loadModels();
  }, [page, searchQuery, providerFilter, activeFilter, supportsImagesFilter]);

  // Reset selectedInstructModel when edit modal opens
  useEffect(() => {
    if (showEditModal && selectedModelForEdit) {
      setSelectedInstructModel(selectedModelForEdit.instruct_model_public_id || '');
    }
  }, [showEditModal, selectedModelForEdit]);

  async function loadAllModels() {
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      // Fetch a large number of models for the dropdown
      const response = await adminClient.providerModels.listProviderModels({
        limit: 500,
        offset: 0,
      });
      setAllModels(response.data);
    } catch (err) {
      console.error('Failed to load all models:', err);
    }
  }

  async function loadProviders() {
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      const response = await adminClient.providers.listProviders({ limit: 100 });
      setProviders(response.data);
    } catch (err) {
      console.error('Failed to load providers:', err);
    }
  }

  async function loadModels() {
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
      const response = await adminClient.providerModels.listProviderModels({
        limit,
        offset: page * limit,
        search: searchQuery || undefined,
        provider_id: providerFilter,
        active: activeFilter,
        supports_images: supportsImagesFilter,
      });

      setModels(response.data);
      setTotalModels(response.total || 0);
      setSelectedModels(new Set());
    } catch (err) {
      console.error('Failed to load models:', err);
      setError('Failed to load models');
    } finally {
      setIsLoading(false);
    }
  }

  async function handleEditModel(data: Partial<ProviderModel>) {
    if (!selectedModelForEdit) return;
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.providerModels.updateProviderModel(selectedModelForEdit.id, data);

      setShowEditModal(false);
      setSelectedModelForEdit(null);
      loadModels();
    } catch (err) {
      console.error('Failed to update model:', err);
      alert('Failed to update model');
    }
  }

  async function handleToggleModelActive(modelPublicId: string, currentlyActive: boolean) {
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);

      if (currentlyActive) {
        await adminClient.providerModels.deactivateModel(modelPublicId);
      } else {
        await adminClient.providerModels.activateModel(modelPublicId);
      }

      loadModels();
    } catch (err) {
      console.error('Failed to update model status:', err);
      alert('Failed to update model status');
    }
  }

  async function handleBulkActivate() {
    if (selectedModels.size === 0) return;
    if (!confirm(`Activate ${selectedModels.size} selected models?`)) return;

    try {
      setIsPerformingBulkAction(true);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);

      await Promise.all(
        Array.from(selectedModels).map((modelPublicId) =>
          adminClient.providerModels.activateModel(modelPublicId),
        ),
      );

      alert(`Successfully activated ${selectedModels.size} models`);
      loadModels();
    } catch (err) {
      console.error('Failed to bulk activate:', err);
      alert('Failed to activate models');
    } finally {
      setIsPerformingBulkAction(false);
    }
  }

  async function handleBulkDeactivate() {
    if (selectedModels.size === 0) return;
    if (!confirm(`Deactivate ${selectedModels.size} selected models?`)) return;

    try {
      setIsPerformingBulkAction(true);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);

      await Promise.all(
        Array.from(selectedModels).map((modelPublicId) =>
          adminClient.providerModels.deactivateModel(modelPublicId),
        ),
      );

      alert(`Successfully deactivated ${selectedModels.size} models`);
      loadModels();
    } catch (err) {
      console.error('Failed to bulk deactivate:', err);
      alert('Failed to deactivate models');
    } finally {
      setIsPerformingBulkAction(false);
    }
  }

  function toggleModelSelection(modelPublicId: string) {
    const newSelected = new Set(selectedModels);
    if (newSelected.has(modelPublicId)) {
      newSelected.delete(modelPublicId);
    } else {
      newSelected.add(modelPublicId);
    }
    setSelectedModels(newSelected);
  }

  function toggleSelectAll() {
    if (selectedModels.size === models.length) {
      setSelectedModels(new Set());
    } else {
      setSelectedModels(new Set(models.map((m) => m.id)));
    }
  }

  // ... (rest of the component)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Provider Models</h1>
          <p className="text-muted-foreground">Manage models synced from external providers.</p>
        </div>
      </div>

      <div className="flex flex-col gap-4 md:flex-row md:items-center justify-between">
        <div className="flex items-center gap-2 flex-1">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search by model ID, display name, or provider..."
              className="w-full pl-9 pr-10 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                setPage(0);
              }}
            />
            {searchQuery && (
              <button
                onClick={() => {
                  setSearchQuery('');
                  setPage(0);
                }}
                className="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground transition-colors"
                title="Clear search"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
          <select
            className="px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            value={providerFilter || ''}
            onChange={(e) => {
              setProviderFilter(e.target.value || undefined);
              setPage(0);
            }}
          >
            <option value="">All Providers</option>
            {providers.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
          <select
            className="px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            value={activeFilter === undefined ? '' : activeFilter.toString()}
            onChange={(e) => {
              setActiveFilter(e.target.value === '' ? undefined : e.target.value === 'true');
              setPage(0);
            }}
          >
            <option value="">All Status</option>
            <option value="true">Active</option>
            <option value="false">Inactive</option>
          </select>

          {(searchQuery ||
            providerFilter ||
            activeFilter !== undefined ||
            supportsImagesFilter !== undefined) && (
            <button
              onClick={() => {
                setSearchQuery('');
                setProviderFilter(undefined);
                setActiveFilter(undefined);
                setSupportsImagesFilter(undefined);
                setPage(0);
              }}
              className="px-3 py-2 text-sm text-muted-foreground hover:text-foreground border rounded-md hover:bg-accent transition-colors whitespace-nowrap"
            >
              Clear Filters
            </button>
          )}
        </div>

        {selectedModels.size > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-muted-foreground">{selectedModels.size} selected</span>
            <button
              onClick={handleBulkActivate}
              disabled={isPerformingBulkAction}
              className="px-3 py-2 border rounded-md text-sm hover:bg-accent transition-colors"
            >
              Activate
            </button>
            <button
              onClick={handleBulkDeactivate}
              disabled={isPerformingBulkAction}
              className="px-3 py-2 border rounded-md text-sm hover:bg-accent transition-colors"
            >
              Deactivate
            </button>
          </div>
        )}
      </div>

      <div className="bg-card rounded-lg border overflow-hidden">
        <table className="w-full text-sm text-left">
          <thead className="bg-muted/50 text-muted-foreground font-medium border-b">
            <tr>
              <th className="w-10 px-4 py-3">
                <input
                  type="checkbox"
                  checked={models.length > 0 && selectedModels.size === models.length}
                  onChange={toggleSelectAll}
                  className="rounded border-gray-300"
                />
              </th>
              <th className="px-4 py-3">Model</th>
              <th className="px-4 py-3">Provider</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Category</th>
              <th className="px-4 py-3">Pricing (1M tokens)</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {isLoading ? (
              <tr>
                <td colSpan={7} className="p-8 text-center">
                  <Loader2 className="w-6 h-6 animate-spin mx-auto text-muted-foreground" />
                </td>
              </tr>
            ) : models.length === 0 ? (
              <tr>
                <td colSpan={7} className="p-8 text-center text-muted-foreground">
                  No models found
                </td>
              </tr>
            ) : (
              models.map((model) => {
                const provider = providers.find((p) => p.id === model.provider_id);
                return (
                  <tr key={model.id} className="hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3">
                      <input
                        type="checkbox"
                        checked={selectedModels.has(model.id)}
                        onChange={() => toggleModelSelection(model.id)}
                        className="rounded border-gray-300"
                      />
                    </td>
                    <td className="px-4 py-3">
                      <div className="space-y-0.5">
                        <div className="font-medium">{model.model_display_name}</div>
                        <div className="text-xs flex items-center gap-1.5">
                          <code className="bg-muted px-1 py-0.5 rounded text-muted-foreground font-mono">
                            {model.model_public_id}
                          </code>
                          <Link
                            href={`/admin/models/catalogs?search=${encodeURIComponent(model.model_public_id)}`}
                            className="text-muted-foreground hover:text-primary transition-colors"
                            title="View in Catalogs"
                          >
                            <ExternalLink className="w-3 h-3" />
                          </Link>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      {provider ? (
                        <div className="flex items-center gap-2">
                          <div className="font-medium text-sm">{provider.name}</div>
                        </div>
                      ) : (
                        <span className="text-muted-foreground text-sm">Unknown</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`px-2 py-1 rounded-md text-xs font-medium ${
                          model.active
                            ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                            : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
                        }`}
                      >
                        {model.active ? 'Active' : 'Inactive'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className="px-2 py-1 rounded-md bg-muted text-xs font-medium">
                        {model.category}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <div className="flex flex-col gap-1">
                        <div className="flex justify-between gap-4">
                          <span className="text-muted-foreground">Input:</span>
                          <span>${model.pricing?.prompt?.toFixed(6) || '0.00'}</span>
                        </div>
                        <div className="flex justify-between gap-4">
                          <span className="text-muted-foreground">Output:</span>
                          <span>${model.pricing?.completion?.toFixed(6) || '0.00'}</span>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => {
                            setSelectedModelForEdit(model);
                            setShowEditModal(true);
                          }}
                          className="p-1.5 hover:bg-accent rounded-md transition-colors text-muted-foreground hover:text-foreground"
                          title="Edit model"
                        >
                          <Edit2 className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleToggleModelActive(model.id, model.active)}
                          className={`p-1.5 hover:bg-accent rounded-md transition-colors ${
                            model.active ? 'text-gray-600' : 'text-green-600'
                          }`}
                          title={model.active ? 'Deactivate model' : 'Activate model'}
                        >
                          {model.active ? (
                            <XCircle className="w-4 h-4" />
                          ) : (
                            <CheckCircle2 className="w-4 h-4" />
                          )}
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between border-t pt-4">
        <div className="text-sm text-muted-foreground">
          Showing {models.length} of {totalModels} models
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setPage(Math.max(0, page - 1))}
            disabled={page === 0}
            className="p-2 border rounded-md hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
          <span className="text-sm font-medium">
            Page {page + 1} of {Math.max(1, Math.ceil(totalModels / limit))}
          </span>
          <button
            onClick={() => setPage(Math.min(Math.ceil(totalModels / limit) - 1, page + 1))}
            disabled={page >= Math.ceil(totalModels / limit) - 1}
            className="p-2 border rounded-md hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Edit Modal */}
      {showEditModal && selectedModelForEdit && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-card rounded-lg border max-w-lg w-full p-6 max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Edit Provider Model</h2>
              <button
                onClick={() => {
                  setShowEditModal(false);
                  setSelectedModelForEdit(null);
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
                handleEditModel({
                  model_display_name: formData.get('model_display_name') as string,
                  category: formData.get('category') as string,
                  category_order_number: Number(formData.get('category_order_number')) || undefined,
                  model_order_number: Number(formData.get('model_order_number')) || undefined,
                  active: formData.get('active') === 'on',
                  pricing: {
                    ...(selectedModelForEdit.pricing || {}),
                    prompt: Number(formData.get('pricing_prompt')) || 0,
                    completion: Number(formData.get('pricing_completion')) || 0,
                  },
                  token_limits: {
                    ...(selectedModelForEdit.token_limits || {}),
                    context_length: Number(formData.get('context_length')) || undefined,
                    max_completion_tokens:
                      Number(formData.get('max_completion_tokens')) || undefined,
                  },
                  supports_images: formData.get('supports_images') === 'on',
                  supports_audio: formData.get('supports_audio') === 'on',
                  supports_video: formData.get('supports_video') === 'on',
                  supports_reasoning: formData.get('supports_reasoning') === 'on',
                  supports_embeddings: formData.get('supports_embeddings') === 'on',
                  instruct_model_public_id: selectedInstructModel || undefined,
                });
              }}
              className="space-y-4"
            >
              <div>
                <label className="block text-sm font-medium mb-1">Display Name</label>
                <input
                  type="text"
                  name="model_display_name"
                  defaultValue={selectedModelForEdit.model_display_name}
                  className="w-full px-3 py-2 border rounded-md bg-background text-sm"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Category</label>
                <input
                  type="text"
                  name="category"
                  list="category-options"
                  defaultValue={selectedModelForEdit.category}
                  className="w-full px-3 py-2 border rounded-md bg-background text-sm"
                  placeholder="Select or type custom category"
                  required
                />
                <datalist id="category-options">
                  <option value="jan" />
                  <option value="legacy" />
                </datalist>
                <p className="text-xs text-muted-foreground mt-1">
                  Select from predefined options or type a custom category
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">Category Order Number</label>
                  <input
                    type="number"
                    name="category_order_number"
                    defaultValue={selectedModelForEdit.category_order_number ?? selectedModelForEdit.category_order}
                    className="w-full px-3 py-2 border rounded-md bg-background text-sm"
                    placeholder="e.g., 1"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Sort order for the category
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Model Order Number</label>
                  <input
                    type="number"
                    name="model_order_number"
                    defaultValue={selectedModelForEdit.model_order_number ?? selectedModelForEdit.model_order}
                    className="w-full px-3 py-2 border rounded-md bg-background text-sm"
                    placeholder="e.g., 1"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Sort order within the category
                  </p>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">Prompt Price</label>
                  <input
                    type="number"
                    name="pricing_prompt"
                    defaultValue={selectedModelForEdit.pricing?.prompt}
                    step="0.000001"
                    className="w-full px-3 py-2 border rounded-md bg-background text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Completion Price</label>
                  <input
                    type="number"
                    name="pricing_completion"
                    defaultValue={selectedModelForEdit.pricing?.completion}
                    step="0.000001"
                    className="w-full px-3 py-2 border rounded-md bg-background text-sm"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">Context Length</label>
                  <input
                    type="number"
                    name="context_length"
                    defaultValue={selectedModelForEdit.token_limits?.context_length}
                    className="w-full px-3 py-2 border rounded-md bg-background text-sm"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Max Output Tokens</label>
                  <input
                    type="number"
                    name="max_completion_tokens"
                    defaultValue={selectedModelForEdit.token_limits?.max_completion_tokens}
                    className="w-full px-3 py-2 border rounded-md bg-background text-sm"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">Capabilities</label>
                <div className="grid grid-cols-2 gap-2">
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      name="supports_images"
                      defaultChecked={selectedModelForEdit.supports_images}
                      className="rounded border-gray-300"
                    />
                    <span className="text-sm">Vision</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      name="supports_audio"
                      defaultChecked={selectedModelForEdit.supports_audio}
                      className="rounded border-gray-300"
                    />
                    <span className="text-sm">Audio</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      name="supports_video"
                      defaultChecked={selectedModelForEdit.supports_video}
                      className="rounded border-gray-300"
                    />
                    <span className="text-sm">Video</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      name="supports_reasoning"
                      defaultChecked={selectedModelForEdit.supports_reasoning}
                      className="rounded border-gray-300"
                    />
                    <span className="text-sm">Reasoning</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      name="supports_embeddings"
                      defaultChecked={selectedModelForEdit.supports_embeddings}
                      className="rounded border-gray-300"
                    />
                    <span className="text-sm">Embeddings</span>
                  </label>
                </div>
              </div>

              {/* Instruct Model Dropdown - shown when catalog has supports_instruct=true */}
              {selectedModelForEdit.supports_instruct && (
                <div>
                  <label className="block text-sm font-medium mb-1">Instruct Model Backup</label>
                  <select
                    value={selectedInstructModel}
                    onChange={(e) => setSelectedInstructModel(e.target.value)}
                    className="w-full px-3 py-2 border rounded-md bg-background text-sm"
                  >
                    <option value="">None (use this model for all requests)</option>
                    {allModels
                      .filter((m) => m.id !== selectedModelForEdit.id) // Show all models as potential backups
                      .map((m) => (
                        <option key={m.id} value={m.id}>
                          {m.model_display_name} ({m.model_public_id}){!m.active ? ' [inactive]' : ''}
                        </option>
                      ))}
                  </select>
                  <p className="text-xs text-muted-foreground mt-1">
                    When enable_thinking=false is passed in the API request, use this model instead.
                  </p>
                </div>
              )}

              <div className="flex items-center gap-2 pt-2">
                <input
                  type="checkbox"
                  name="active"
                  id="active-toggle"
                  defaultChecked={selectedModelForEdit.active}
                  className="rounded border-gray-300"
                />
                <label htmlFor="active-toggle" className="text-sm font-medium">
                  Active
                </label>
              </div>

              <div className="flex justify-end gap-2 pt-4 border-t">
                <button
                  type="button"
                  onClick={() => {
                    setShowEditModal(false);
                    setSelectedModelForEdit(null);
                  }}
                  className="px-4 py-2 text-sm border rounded-md hover:bg-accent transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
                >
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
