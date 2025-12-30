'use client';

import { createAdminAPIClient, FeatureFlag, ModelCatalog } from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
import {
  Brain,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  Edit2,
  Eye,
  FileText,
  Flag,
  Globe,
  Image as ImageIcon,
  Layers,
  Loader2,
  Mic,
  Video,
  Wrench,
  X,
  XCircle,
  Zap,
} from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import { useEffect, useState } from 'react';
import ModelPromptTemplatesTab from './components/ModelPromptTemplatesTab';

export default function ModelCatalogsManagementPage() {
  const searchParams = useSearchParams();
  const [catalogs, setCatalogs] = useState<ModelCatalog[]>([]);
  const [featureFlags, setFeatureFlags] = useState<FeatureFlag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState<string>(searchParams.get('search') || '');
  const [familyFilter, setFamilyFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [experimentalFilter, setExperimentalFilter] = useState<boolean | undefined>(undefined);
  const [featureFlagFilter, setFeatureFlagFilter] = useState<string>('');
  const [supportsEmbeddings, setSupportsEmbeddings] = useState<boolean | undefined>(undefined);
  const [supportsImages, setSupportsImages] = useState<boolean | undefined>(undefined);
  const [supportsReasoning, setSupportsReasoning] = useState<boolean | undefined>(undefined);
  const [supportsAudio, setSupportsAudio] = useState<boolean | undefined>(undefined);
  const [supportsVideo, setSupportsVideo] = useState<boolean | undefined>(undefined);
  const [supportsBrowser, setSupportsBrowser] = useState<boolean | undefined>(undefined);
  const [page, setPage] = useState(0);
  const [totalCatalogs, setTotalCatalogs] = useState(0);
  const [selectedCatalog, setSelectedCatalog] = useState<ModelCatalog | null>(null);
  const [showDetailsModal, setShowDetailsModal] = useState(false);
  const [isEditMode, setIsEditMode] = useState(false);
  const [activeTab, setActiveTab] = useState<'details' | 'prompts'>('details');

  const limit = 20;

  useEffect(() => {
    loadCatalogs();
    loadFeatureFlags();
  }, [
    page,
    searchQuery,
    familyFilter,
    statusFilter,
    experimentalFilter,
    featureFlagFilter,
    supportsEmbeddings,
    supportsImages,
    supportsReasoning,
    supportsAudio,
    supportsVideo,
    supportsBrowser,
  ]);

  async function loadFeatureFlags() {
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      const response = await adminClient.users.listFeatureFlags();
      setFeatureFlags(response.data || []);
    } catch (err) {
      console.error('Failed to load feature flags:', err);
    }
  }

  async function loadCatalogs() {
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
      const response = await adminClient.modelCatalogs.listModelCatalogs({
        limit: 1000, // Load more for client-side filtering
        offset: 0,
        family: familyFilter || undefined,
        status: statusFilter || undefined,
        experimental: experimentalFilter,
        requires_feature_flag: featureFlagFilter || undefined,
        supports_embeddings: supportsEmbeddings,
        supports_images: supportsImages,
        supports_reasoning: supportsReasoning,
        supports_audio: supportsAudio,
        supports_video: supportsVideo,
        supports_browser: supportsBrowser,
      });

      // Client-side filtering by search query
      let filteredData = response.data;
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        filteredData = response.data.filter(
          (catalog) =>
            catalog.id.toLowerCase().includes(query) ||
            (catalog.family && catalog.family.toLowerCase().includes(query)),
        );
      }

      // Client-side pagination
      const startIndex = page * limit;
      const endIndex = startIndex + limit;
      const paginatedData = filteredData.slice(startIndex, endIndex);

      setCatalogs(paginatedData);
      setTotalCatalogs(filteredData.length);
    } catch (err) {
      console.error('Failed to load catalogs:', err);
      setError('Failed to load model catalogs');
    } finally {
      setIsLoading(false);
    }
  }

  async function handleUpdateStatus(catalogId: string, newStatus: string) {
    if (!confirm(`Update catalog status to "${newStatus}"?`)) return;

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.modelCatalogs.updateModelCatalog(catalogId, { status: newStatus });

      loadCatalogs();
    } catch (err) {
      console.error('Failed to update catalog status:', err);
      alert('Failed to update catalog status');
    }
  }

  async function handleEditCatalog(data: Partial<ModelCatalog>) {
    if (!selectedCatalog) return;
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.modelCatalogs.updateModelCatalog(selectedCatalog.id, data);

      setShowDetailsModal(false);
      setSelectedCatalog(null);
      setIsEditMode(false);
      loadCatalogs();
    } catch (err) {
      console.error('Failed to update catalog:', err);
      alert('Failed to update catalog. ' + (err instanceof Error ? err.message : 'Unknown error'));
    }
  }

  async function handleToggleActive(catalogId: string, currentlyActive: boolean) {
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.modelCatalogs.batchToggle({
        enable: !currentlyActive,
        catalog_ids: [catalogId],
      });

      loadCatalogs();
    } catch (err) {
      console.error('Failed to toggle catalog active status:', err);
      alert('Failed to toggle catalog active status');
    }
  }

  const totalPages = Math.ceil(totalCatalogs / limit);

  const families = ['gpt-4', 'gpt-3.5', 'claude', 'gemini', 'llama', 'mistral', 'palm'];
  const statuses = ['filled', 'updated', 'pending', 'deprecated'];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
          <Layers className="w-8 h-8" />
          Model Catalogs
        </h1>
        <p className="text-muted-foreground mt-2">
          Browse and manage model catalog entries and their capabilities
        </p>
      </div>

      {/* Filters */}
      <div className="bg-card rounded-lg border p-4">
        <div className="space-y-4">
          {/* Search Bar */}
          <div className="relative max-w-md">
            <div className="absolute left-3 top-1/2 -translate-y-1/2">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="text-muted-foreground"
              >
                <circle cx="11" cy="11" r="8" />
                <path d="m21 21-4.35-4.35" />
              </svg>
            </div>
            <input
              type="text"
              placeholder="Search by catalog ID or family name..."
              className="w-full pl-10 pr-4 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
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
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>

          {/* Row 1: Family and Status */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <select
              value={familyFilter}
              onChange={(e) => {
                setFamilyFilter(e.target.value);
                setPage(0);
              }}
              className="px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            >
              <option value="">All Families</option>
              {families.map((family) => (
                <option key={family} value={family}>
                  {family}
                </option>
              ))}
            </select>

            <select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setPage(0);
              }}
              className="px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            >
              <option value="">All Statuses</option>
              {statuses.map((status) => (
                <option key={status} value={status}>
                  {status.charAt(0).toUpperCase() + status.slice(1)}
                </option>
              ))}
            </select>

            <select
              value={experimentalFilter === undefined ? '' : experimentalFilter.toString()}
              onChange={(e) => {
                const val = e.target.value;
                setExperimentalFilter(val === '' ? undefined : val === 'true');
                setPage(0);
              }}
              className="px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            >
              <option value="">All Types</option>
              <option value="true">Experimental</option>
              <option value="false">Stable</option>
            </select>

            <select
              value={featureFlagFilter}
              onChange={(e) => {
                setFeatureFlagFilter(e.target.value);
                setPage(0);
              }}
              className="px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            >
              <option value="">All Access Levels</option>
              {featureFlags.map((flag) => (
                <option key={flag.id} value={flag.key}>
                  Requires: {flag.name}
                </option>
              ))}
            </select>

            {(searchQuery ||
              familyFilter ||
              statusFilter ||
              experimentalFilter !== undefined ||
              featureFlagFilter ||
              supportsEmbeddings !== undefined ||
              supportsImages !== undefined ||
              supportsReasoning !== undefined ||
              supportsAudio !== undefined ||
              supportsVideo !== undefined ||
              supportsBrowser !== undefined) && (
              <button
                onClick={() => {
                  setSearchQuery('');
                  setFamilyFilter('');
                  setStatusFilter('');
                  setExperimentalFilter(undefined);
                  setFeatureFlagFilter('');
                  setSupportsEmbeddings(undefined);
                  setSupportsImages(undefined);
                  setSupportsReasoning(undefined);
                  setSupportsAudio(undefined);
                  setSupportsVideo(undefined);
                  setSupportsBrowser(undefined);
                  setPage(0);
                }}
                className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground border rounded-md hover:bg-accent transition-colors"
              >
                Clear Filters
              </button>
            )}
          </div>

          {/* Row 2: Capability Filters */}
          <div>
            <p className="text-sm font-medium mb-2">Filter by Capabilities:</p>
            <div className="grid grid-cols-2 md:grid-cols-6 gap-3">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={supportsImages === true}
                  onChange={(e) => {
                    setSupportsImages(e.target.checked ? true : undefined);
                    setPage(0);
                  }}
                  className="rounded border-gray-300"
                />
                <div className="flex items-center gap-1.5">
                  <ImageIcon className="w-4 h-4 text-purple-600" />
                  <span className="text-sm">Vision</span>
                </div>
              </label>

              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={supportsEmbeddings === true}
                  onChange={(e) => {
                    setSupportsEmbeddings(e.target.checked ? true : undefined);
                    setPage(0);
                  }}
                  className="rounded border-gray-300"
                />
                <div className="flex items-center gap-1.5">
                  <Zap className="w-4 h-4 text-green-600" />
                  <span className="text-sm">Embeddings</span>
                </div>
              </label>

              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={supportsReasoning === true}
                  onChange={(e) => {
                    setSupportsReasoning(e.target.checked ? true : undefined);
                    setPage(0);
                  }}
                  className="rounded border-gray-300"
                />
                <div className="flex items-center gap-1.5">
                  <Brain className="w-4 h-4 text-orange-600" />
                  <span className="text-sm">Reasoning</span>
                </div>
              </label>

              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={supportsAudio === true}
                  onChange={(e) => {
                    setSupportsAudio(e.target.checked ? true : undefined);
                    setPage(0);
                  }}
                  className="rounded border-gray-300"
                />
                <div className="flex items-center gap-1.5">
                  <Mic className="w-4 h-4 text-blue-600" />
                  <span className="text-sm">Audio</span>
                </div>
              </label>

              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={supportsVideo === true}
                  onChange={(e) => {
                    setSupportsVideo(e.target.checked ? true : undefined);
                    setPage(0);
                  }}
                  className="rounded border-gray-300"
                />
                <div className="flex items-center gap-1.5">
                  <Video className="w-4 h-4 text-red-600" />
                  <span className="text-sm">Video</span>
                </div>
              </label>

              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={supportsBrowser === true}
                  onChange={(e) => {
                    setSupportsBrowser(e.target.checked ? true : undefined);
                    setPage(0);
                  }}
                  className="rounded border-gray-300"
                />
                <div className="flex items-center gap-1.5">
                  <Globe className="w-4 h-4 text-cyan-600" />
                  <span className="text-sm">Browser</span>
                </div>
              </label>
            </div>
          </div>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      {/* Catalogs Grid */}
      <div className="space-y-4">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="w-6 h-6 animate-spin text-primary" />
          </div>
        ) : catalogs.length === 0 ? (
          <div className="text-center py-12 bg-card rounded-lg border">
            <Layers className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No model catalogs found</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {catalogs.map((catalog) => (
              <div
                key={catalog.id}
                className="bg-card rounded-lg border p-5 hover:shadow-md transition-shadow"
              >
                {/* Header */}
                <div className="flex items-start justify-between mb-4">
                  <div className="flex-1">
                    <h3 className="font-semibold text-lg mb-1">
                      {(catalog as any).model_display_name || catalog.id}
                    </h3>
                    {catalog.family && (
                      <p className="text-sm text-muted-foreground mb-1">Family: {catalog.family}</p>
                    )}
                    <code className="text-xs bg-muted px-2 py-1 rounded">{catalog.id}</code>
                  </div>
                  {catalog.status && (
                    <span
                      className={`px-2 py-1 rounded-md text-xs font-medium ${
                        catalog.status === 'filled'
                          ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                          : catalog.status === 'updated'
                            ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                            : catalog.status === 'pending'
                              ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                              : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
                      }`}
                    >
                      {catalog.status}
                    </span>
                  )}
                  <span
                    className={`ml-2 px-2 py-1 rounded-md text-xs font-medium ${
                      catalog.active
                        ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                        : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
                    }`}
                  >
                    {catalog.active ? 'Active' : 'Inactive'}
                  </span>
                  {catalog.experimental && (
                    <span className="ml-2 px-2 py-1 rounded-md text-xs font-medium bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400 flex items-center gap-1">
                      <Flag className="w-3 h-3" />
                      Experimental
                    </span>
                  )}
                  {catalog.requires_feature_flag && (
                    <span className="ml-2 px-2 py-1 rounded-md text-xs font-medium bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400 flex items-center gap-1">
                      <Flag className="w-3 h-3" />
                      {featureFlags.find((f) => f.key === catalog.requires_feature_flag)?.name ||
                        catalog.requires_feature_flag}
                    </span>
                  )}
                </div>

                {/* Architecture */}
                {catalog.architecture && (
                  <div className="mb-3">
                    <p className="text-xs text-muted-foreground mb-1">Architecture</p>
                    <p className="text-sm font-medium">
                      {typeof catalog.architecture === 'string'
                        ? catalog.architecture
                        : (catalog.architecture as any).modality || 'Complex'}
                    </p>
                  </div>
                )}

                {/* Capabilities */}
                <div className="mb-4">
                  <p className="text-xs text-muted-foreground mb-2">Capabilities</p>
                  <div className="flex flex-wrap gap-2">
                    {catalog.supports_images && (
                      <div className="flex items-center gap-1 px-2 py-1 bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400 rounded-md text-xs">
                        <ImageIcon className="w-3 h-3" />
                        Vision
                      </div>
                    )}
                    {catalog.supports_embeddings && (
                      <div className="flex items-center gap-1 px-2 py-1 bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 rounded-md text-xs">
                        <Zap className="w-3 h-3" />
                        Embeddings
                      </div>
                    )}
                    {catalog.supports_reasoning && (
                      <div className="flex items-center gap-1 px-2 py-1 bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400 rounded-md text-xs">
                        <Brain className="w-3 h-3" />
                        Reasoning
                      </div>
                    )}
                    {catalog.supports_audio && (
                      <div className="flex items-center gap-1 px-2 py-1 bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 rounded-md text-xs">
                        <Mic className="w-3 h-3" />
                        Audio
                      </div>
                    )}
                    {catalog.supports_video && (
                      <div className="flex items-center gap-1 px-2 py-1 bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 rounded-md text-xs">
                        <Video className="w-3 h-3" />
                        Video
                      </div>
                    )}
                    {catalog.supports_browser && (
                      <div className="flex items-center gap-1 px-2 py-1 bg-cyan-100 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-400 rounded-md text-xs">
                        <Globe className="w-3 h-3" />
                        Browser
                      </div>
                    )}
                    {!catalog.supports_images &&
                      !catalog.supports_embeddings &&
                      !catalog.supports_reasoning &&
                      !catalog.supports_audio &&
                      !catalog.supports_video &&
                      !catalog.supports_browser && (
                        <span className="text-xs text-muted-foreground">
                          No capabilities listed
                        </span>
                      )}
                  </div>
                </div>

                {/* Supported Parameters */}
                {catalog.supported_parameters &&
                  typeof catalog.supported_parameters === 'object' && (
                    <div className="mb-4">
                      <p className="text-xs text-muted-foreground mb-1">Supported Parameters</p>
                      <div className="flex flex-wrap gap-1">
                        {Object.keys(catalog.supported_parameters)
                          .slice(0, 4)
                          .map((paramKey) => (
                            <span
                              key={paramKey}
                              className="px-1.5 py-0.5 bg-muted rounded text-xs"
                              title={JSON.stringify(
                                (catalog.supported_parameters as Record<string, any>)[paramKey],
                              )}
                            >
                              {paramKey}
                            </span>
                          ))}
                        {Object.keys(catalog.supported_parameters).length > 4 && (
                          <span className="px-1.5 py-0.5 text-xs text-muted-foreground">
                            +{Object.keys(catalog.supported_parameters).length - 4}
                          </span>
                        )}
                      </div>
                    </div>
                  )}

                {/* Actions */}
                <div className="flex gap-2 pt-3 border-t">
                  <button
                    onClick={() => {
                      setSelectedCatalog(catalog);
                      setShowDetailsModal(true);
                      setIsEditMode(false);
                    }}
                    className="flex-1 flex items-center justify-center gap-2 px-3 py-2 border rounded-md hover:bg-accent transition-colors text-sm"
                    title="View details"
                  >
                    <Eye className="w-4 h-4" />
                    View
                  </button>
                  <button
                    onClick={() => {
                      setSelectedCatalog(catalog);
                      setShowDetailsModal(true);
                      setIsEditMode(true);
                    }}
                    className="flex-1 flex items-center justify-center gap-2 px-3 py-2 border rounded-md hover:bg-accent transition-colors text-sm"
                    title="Edit catalog"
                  >
                    <Edit2 className="w-4 h-4" />
                    Edit
                  </button>
                  <button
                    onClick={() => handleToggleActive(catalog.id, !!catalog.active)}
                    className={`p-2 border rounded-md hover:bg-accent transition-colors ${
                      catalog.active ? 'text-gray-600' : 'text-green-600'
                    }`}
                    title={catalog.active ? 'Deactivate catalog' : 'Activate catalog'}
                  >
                    {catalog.active ? (
                      <XCircle className="w-4 h-4" />
                    ) : (
                      <CheckCircle2 className="w-4 h-4" />
                    )}
                  </button>
                </div>

                {/* Timestamps */}
                {catalog.updated_at && (
                  <div className="mt-3 pt-3 border-t">
                    <p className="text-xs text-muted-foreground">
                      Updated:{' '}
                      {(() => {
                        try {
                          const date = new Date(catalog.updated_at);
                          // Check if date is valid
                          if (isNaN(date.getTime())) return 'Invalid date';
                          // Check if year is before 2000 (likely wrong timestamp)
                          if (date.getFullYear() < 2000) return 'Invalid date';
                          return date.toLocaleDateString('en-GB', {
                            day: '2-digit',
                            month: '2-digit',
                            year: 'numeric',
                          });
                        } catch (e) {
                          return 'Invalid date';
                        }
                      })()}
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
            Showing {page * limit + 1} to {Math.min((page + 1) * limit, totalCatalogs)} of{' '}
            {totalCatalogs} catalogs
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

      {/* Details/Edit Modal */}
      {showDetailsModal && selectedCatalog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-card rounded-lg border max-w-2xl w-full p-6 max-h-[80vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">
                {isEditMode ? 'Edit Catalog' : 'Catalog Details'}
              </h2>
              <div className="flex items-center gap-2">
                {!isEditMode && (
                  <button
                    onClick={() => setIsEditMode(true)}
                    className="p-1.5 hover:bg-accent rounded-md text-muted-foreground hover:text-foreground"
                    title="Edit"
                  >
                    <Edit2 className="w-4 h-4" />
                  </button>
                )}
                <button
                  onClick={() => {
                    setShowDetailsModal(false);
                    setSelectedCatalog(null);
                    setIsEditMode(false);
                  }}
                  className="p-1 hover:bg-accent rounded-md"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>
            </div>

            {isEditMode ? (
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  const formData = new FormData(e.currentTarget);

                  // Parse architecture
                  const inputModalities: string[] = [];
                  if (formData.get('input_text') === 'on') inputModalities.push('text');
                  if (formData.get('input_image') === 'on') inputModalities.push('image');
                  if (formData.get('input_file') === 'on') inputModalities.push('file');
                  if (formData.get('input_audio') === 'on') inputModalities.push('audio');
                  if (formData.get('input_video') === 'on') inputModalities.push('video');

                  const outputModalities: string[] = [];
                  if (formData.get('output_text') === 'on') outputModalities.push('text');
                  if (formData.get('output_image') === 'on') outputModalities.push('image');
                  if (formData.get('output_embedding') === 'on') outputModalities.push('embedding');

                  // Parse supported parameters
                  const paramNames: string[] = [];
                  if (formData.get('param_temperature') === 'on') paramNames.push('temperature');
                  if (formData.get('param_top_p') === 'on') paramNames.push('top_p');
                  if (formData.get('param_top_k') === 'on') paramNames.push('top_k');
                  if (formData.get('param_presence_penalty') === 'on')
                    paramNames.push('presence_penalty');
                  if (formData.get('param_repetition_penalty') === 'on')
                    paramNames.push('repetition_penalty');
                  if (formData.get('param_frequency_penalty') === 'on')
                    paramNames.push('frequency_penalty');
                  if (formData.get('param_max_tokens') === 'on') paramNames.push('max_tokens');
                  if (formData.get('param_logit_bias') === 'on') paramNames.push('logit_bias');
                  if (formData.get('param_logprobs') === 'on') paramNames.push('logprobs');
                  if (formData.get('param_top_logprobs') === 'on') paramNames.push('top_logprobs');
                  if (formData.get('param_seed') === 'on') paramNames.push('seed');
                  if (formData.get('param_response_format') === 'on')
                    paramNames.push('response_format');
                  if (formData.get('param_structured_outputs') === 'on')
                    paramNames.push('structured_outputs');
                  if (formData.get('param_stop') === 'on') paramNames.push('stop');
                  if (formData.get('param_tools') === 'on') paramNames.push('tools');
                  if (formData.get('param_tool_choice') === 'on') paramNames.push('tool_choice');
                  if (formData.get('param_parallel_tool_calls') === 'on')
                    paramNames.push('parallel_tool_calls');
                  if (formData.get('param_include_reasoning') === 'on')
                    paramNames.push('include_reasoning');
                  if (formData.get('param_reasoning') === 'on') paramNames.push('reasoning');
                  if (formData.get('param_web_search_options') === 'on')
                    paramNames.push('web_search_options');
                  if (formData.get('param_verbosity') === 'on') paramNames.push('verbosity');

                  // Helper function to get non-empty numeric value
                  const getNumericValue = (key: string, isInteger = false): number | undefined => {
                    const value = formData.get(key) as string;
                    if (!value || value.trim() === '') return undefined;
                    const parsed = isInteger ? parseInt(value) : parseFloat(value);
                    return isNaN(parsed) ? undefined : parsed;
                  };

                  // Build default parameters object with only defined values
                  const defaultParams: Record<string, number> = {};
                  const temperature = getNumericValue('temperature');
                  if (temperature !== undefined) defaultParams.temperature = temperature;
                  const top_p = getNumericValue('top_p');
                  if (top_p !== undefined) defaultParams.top_p = top_p;
                  const top_k = getNumericValue('top_k', true);
                  if (top_k !== undefined) defaultParams.top_k = top_k;
                  const presence_penalty = getNumericValue('presence_penalty');
                  if (presence_penalty !== undefined)
                    defaultParams.presence_penalty = presence_penalty;
                  const repetition_penalty = getNumericValue('repetition_penalty');
                  if (repetition_penalty !== undefined)
                    defaultParams.repetition_penalty = repetition_penalty;
                  const frequency_penalty = getNumericValue('frequency_penalty');
                  if (frequency_penalty !== undefined)
                    defaultParams.frequency_penalty = frequency_penalty;

                  // Parse tags
                  const tagsStr = formData.get('tags') as string;
                  const tags = tagsStr
                    ? tagsStr
                        .split(',')
                        .map((t) => t.trim())
                        .filter((t) => t)
                    : [];

                  const contextLengthStr = formData.get('context_length') as string;

                  handleEditCatalog({
                    model_display_name: formData.get('model_display_name') as string,
                    description: formData.get('description') as string,
                    family: formData.get('family') as string,
                    status: formData.get('status') as string,
                    notes: formData.get('notes') as string,
                    is_moderated: formData.get('is_moderated') === 'on',
                    experimental: formData.get('experimental') === 'on',
                    requires_feature_flag:
                      (formData.get('requires_feature_flag') as string) || null,
                    supports_images: formData.get('supports_images') === 'on',
                    supports_embeddings: formData.get('supports_embeddings') === 'on',
                    supports_reasoning: formData.get('supports_reasoning') === 'on',
                    supports_instruct: formData.get('supports_instruct') === 'on',
                    supports_audio: formData.get('supports_audio') === 'on',
                    supports_video: formData.get('supports_video') === 'on',
                    supports_tools: formData.get('supports_tools') === 'on',
                    supports_browser: formData.get('supports_browser') === 'on',
                    architecture: {
                      input_modalities: inputModalities.length > 0 ? inputModalities : undefined,
                      output_modalities: outputModalities.length > 0 ? outputModalities : undefined,
                      instruct_type: (formData.get('instruct_type') as string) || undefined,
                      modality: (formData.get('modality') as string) || undefined,
                      tokenizer: (formData.get('tokenizer') as string) || undefined,
                    },
                    context_length: contextLengthStr ? parseFloat(contextLengthStr) : undefined,
                    supported_parameters: {
                      names: paramNames,
                      default: defaultParams,
                    },
                    tags: tags.length > 0 ? tags : undefined,
                  });
                }}
                className="space-y-6"
              >
                {/* Catalog ID (Read-only) */}
                <div className="bg-muted/30 p-3 rounded-md">
                  <label className="block text-sm font-medium mb-1">Catalog ID (Read-only)</label>
                  <code className="block text-sm text-muted-foreground">{selectedCatalog.id}</code>
                </div>

                {/* Basic Information */}
                <div className="border-t pt-4">
                  <h3 className="text-lg font-semibold mb-3">Basic Information</h3>

                  <div className="space-y-3">
                    {/* Model Display Name */}
                    <div>
                      <label className="block text-sm font-medium mb-1">
                        Model Display Name <span className="text-red-500">*</span>
                      </label>
                      <input
                        type="text"
                        name="model_display_name"
                        defaultValue={(selectedCatalog as any).model_display_name || ''}
                        placeholder="e.g., GPT-4 Turbo"
                        className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      />
                    </div>

                    {/* Description */}
                    <div>
                      <label className="block text-sm font-medium mb-1">Description</label>
                      <textarea
                        name="description"
                        defaultValue={(selectedCatalog as any).description || ''}
                        placeholder="Brief description of the model..."
                        rows={2}
                        className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      />
                    </div>

                    {/* Family */}
                    <div>
                      <label className="block text-sm font-medium mb-1">Family</label>
                      <input
                        type="text"
                        name="family"
                        defaultValue={selectedCatalog.family || ''}
                        placeholder="e.g., gpt-4, claude, gemini"
                        className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      />
                    </div>

                    {/* Access Control */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {/* Experimental Flag */}
                      <div className="flex items-center gap-2">
                        <input
                          type="checkbox"
                          name="experimental"
                          id="experimental"
                          defaultChecked={selectedCatalog.experimental}
                          className="rounded border-gray-300"
                        />
                        <label htmlFor="experimental" className="text-sm font-medium">
                          Experimental Model
                        </label>
                      </div>

                      {/* Feature Flag Requirement */}
                      <div>
                        <label className="block text-sm font-medium mb-1">
                          Required Feature Flag
                        </label>
                        <select
                          name="requires_feature_flag"
                          defaultValue={selectedCatalog.requires_feature_flag || ''}
                          className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                        >
                          <option value="">None (Public)</option>
                          {featureFlags.map((flag) => (
                            <option key={flag.id} value={flag.key}>
                              {flag.name} ({flag.key})
                            </option>
                          ))}
                        </select>
                      </div>
                    </div>

                    {/* Status */}
                    <div>
                      <label className="block text-sm font-medium mb-1">Status</label>
                      <select
                        name="status"
                        defaultValue={selectedCatalog.status || ''}
                        className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      >
                        <option value="">Select status</option>
                        <option value="filled">Filled</option>
                        <option value="updated">Updated</option>
                        <option value="pending">Pending</option>
                        <option value="deprecated">Deprecated</option>
                      </select>
                    </div>

                    {/* Context Length */}
                    <div>
                      <label className="block text-sm font-medium mb-1">
                        Context Length (tokens)
                      </label>
                      <input
                        type="number"
                        name="context_length"
                        defaultValue={(selectedCatalog as any).context_length || ''}
                        placeholder="e.g., 128000"
                        className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      />
                    </div>

                    {/* Tags */}
                    <div>
                      <label className="block text-sm font-medium mb-1">
                        Tags (comma-separated)
                      </label>
                      <input
                        type="text"
                        name="tags"
                        defaultValue={((selectedCatalog as any).tags || []).join(', ')}
                        placeholder="e.g., vision, reasoning, multimodal"
                        className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      />
                      <p className="text-xs text-muted-foreground mt-1">
                        Separate tags with commas
                      </p>
                    </div>
                  </div>
                </div>

                {/* Architecture */}
                <div className="border-t pt-4">
                  <h3 className="text-lg font-semibold mb-3">Architecture</h3>

                  <div className="space-y-4">
                    {/* Instruct Type */}
                    <div>
                      <label className="block text-sm font-medium mb-1">Instruct Type</label>
                      <select
                        name="instruct_type"
                        defaultValue={
                          typeof selectedCatalog.architecture === 'object'
                            ? (selectedCatalog.architecture as any).instruct_type || ''
                            : ''
                        }
                        className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      >
                        <option value="">Select instruct type</option>
                        <option value="none">None</option>
                        <option value="airoboros">Airoboros</option>
                        <option value="alpaca">Alpaca</option>
                        <option value="alpaca-modif">Alpaca Modified</option>
                        <option value="chatml">ChatML</option>
                        <option value="claude">Claude</option>
                        <option value="code-llama">Code Llama</option>
                        <option value="gemma">Gemma</option>
                        <option value="llama2">Llama 2</option>
                        <option value="llama3">Llama 3</option>
                        <option value="mistral">Mistral</option>
                        <option value="nemotron">Nemotron</option>
                        <option value="neural">Neural</option>
                        <option value="openchat">OpenChat</option>
                        <option value="phi3">Phi-3</option>
                        <option value="rwkv">RWKV</option>
                        <option value="vicuna">Vicuna</option>
                        <option value="zephyr">Zephyr</option>
                        <option value="deepseek-r1">DeepSeek R1</option>
                        <option value="deepseek-v3.1">DeepSeek V3.1</option>
                        <option value="qwq">QwQ</option>
                        <option value="qwen3">Qwen3</option>
                      </select>
                    </div>

                    {/* Tokenizer */}
                    <div>
                      <label className="block text-sm font-medium mb-1">Tokenizer</label>
                      <select
                        name="tokenizer"
                        defaultValue={
                          typeof selectedCatalog.architecture === 'object'
                            ? (selectedCatalog.architecture as any).tokenizer || ''
                            : ''
                        }
                        className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      >
                        <option value="">Select tokenizer</option>
                        <option value="Router">Router</option>
                        <option value="Media">Media</option>
                        <option value="Other">Other</option>
                        <option value="GPT">GPT</option>
                        <option value="Claude">Claude</option>
                        <option value="Gemini">Gemini</option>
                        <option value="Grok">Grok</option>
                        <option value="Cohere">Cohere</option>
                        <option value="Nova">Nova</option>
                        <option value="Qwen">Qwen</option>
                        <option value="Yi">Yi</option>
                        <option value="DeepSeek">DeepSeek</option>
                        <option value="Mistral">Mistral</option>
                        <option value="Llama2">Llama 2</option>
                        <option value="Llama3">Llama 3</option>
                        <option value="Llama4">Llama 4</option>
                        <option value="PaLM">PaLM</option>
                        <option value="RWKV">RWKV</option>
                        <option value="Qwen3">Qwen3</option>
                      </select>
                    </div>

                    {/* Modality */}
                    <div>
                      <label className="block text-sm font-medium mb-1">Modality</label>
                      <input
                        type="text"
                        name="modality"
                        defaultValue={
                          typeof selectedCatalog.architecture === 'object'
                            ? (selectedCatalog.architecture as any).modality || ''
                            : ''
                        }
                        placeholder="e.g., text-to-text, multimodal"
                        className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      />
                    </div>

                    {/* Input Modalities */}
                    <div>
                      <label className="block text-sm font-medium mb-2">Input Modalities</label>
                      <div className="grid grid-cols-2 gap-2">
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            name="input_text"
                            defaultChecked={
                              typeof selectedCatalog.architecture === 'object' &&
                              Array.isArray(
                                (selectedCatalog.architecture as any).input_modalities,
                              ) &&
                              (selectedCatalog.architecture as any).input_modalities.includes(
                                'text',
                              )
                            }
                            className="rounded border-gray-300"
                          />
                          <span className="text-sm">Text</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            name="input_image"
                            defaultChecked={
                              typeof selectedCatalog.architecture === 'object' &&
                              Array.isArray(
                                (selectedCatalog.architecture as any).input_modalities,
                              ) &&
                              (selectedCatalog.architecture as any).input_modalities.includes(
                                'image',
                              )
                            }
                            className="rounded border-gray-300"
                          />
                          <span className="text-sm">Image</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            name="input_file"
                            defaultChecked={
                              typeof selectedCatalog.architecture === 'object' &&
                              Array.isArray(
                                (selectedCatalog.architecture as any).input_modalities,
                              ) &&
                              (selectedCatalog.architecture as any).input_modalities.includes(
                                'file',
                              )
                            }
                            className="rounded border-gray-300"
                          />
                          <span className="text-sm">File</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            name="input_audio"
                            defaultChecked={
                              typeof selectedCatalog.architecture === 'object' &&
                              Array.isArray(
                                (selectedCatalog.architecture as any).input_modalities,
                              ) &&
                              (selectedCatalog.architecture as any).input_modalities.includes(
                                'audio',
                              )
                            }
                            className="rounded border-gray-300"
                          />
                          <span className="text-sm">Audio</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            name="input_video"
                            defaultChecked={
                              typeof selectedCatalog.architecture === 'object' &&
                              Array.isArray(
                                (selectedCatalog.architecture as any).input_modalities,
                              ) &&
                              (selectedCatalog.architecture as any).input_modalities.includes(
                                'video',
                              )
                            }
                            className="rounded border-gray-300"
                          />
                          <span className="text-sm">Video</span>
                        </label>
                      </div>
                    </div>

                    {/* Output Modalities */}
                    <div>
                      <label className="block text-sm font-medium mb-2">Output Modalities</label>
                      <div className="grid grid-cols-2 gap-2">
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            name="output_text"
                            defaultChecked={
                              typeof selectedCatalog.architecture === 'object' &&
                              Array.isArray(
                                (selectedCatalog.architecture as any).output_modalities,
                              ) &&
                              (selectedCatalog.architecture as any).output_modalities.includes(
                                'text',
                              )
                            }
                            className="rounded border-gray-300"
                          />
                          <span className="text-sm">Text</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            name="output_image"
                            defaultChecked={
                              typeof selectedCatalog.architecture === 'object' &&
                              Array.isArray(
                                (selectedCatalog.architecture as any).output_modalities,
                              ) &&
                              (selectedCatalog.architecture as any).output_modalities.includes(
                                'image',
                              )
                            }
                            className="rounded border-gray-300"
                          />
                          <span className="text-sm">Image</span>
                        </label>
                        <label className="flex items-center gap-2 cursor-pointer">
                          <input
                            type="checkbox"
                            name="output_embedding"
                            defaultChecked={
                              typeof selectedCatalog.architecture === 'object' &&
                              Array.isArray(
                                (selectedCatalog.architecture as any).output_modalities,
                              ) &&
                              (selectedCatalog.architecture as any).output_modalities.includes(
                                'embedding',
                              )
                            }
                            className="rounded border-gray-300"
                          />
                          <span className="text-sm">Embedding</span>
                        </label>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Supported Parameters */}
                <div className="border-t pt-4">
                  <h3 className="text-lg font-semibold mb-3">Supported Parameters</h3>

                  <div className="space-y-4">
                    {/* Parameter Names (checkboxes) */}
                    <div>
                      <label className="block text-sm font-medium mb-2">Available Parameters</label>
                      <div className="grid grid-cols-2 gap-2 max-h-48 overflow-y-auto border rounded-md p-3">
                        {[
                          'temperature',
                          'top_p',
                          'top_k',
                          'presence_penalty',
                          'repetition_penalty',
                          'frequency_penalty',
                          'max_tokens',
                          'logit_bias',
                          'logprobs',
                          'top_logprobs',
                          'seed',
                          'response_format',
                          'structured_outputs',
                          'stop',
                          'tools',
                          'tool_choice',
                          'parallel_tool_calls',
                          'include_reasoning',
                          'reasoning',
                          'web_search_options',
                          'verbosity',
                        ].map((param) => (
                          <label key={param} className="flex items-center gap-2 cursor-pointer">
                            <input
                              type="checkbox"
                              name={`param_${param}`}
                              defaultChecked={
                                selectedCatalog.supported_parameters &&
                                Array.isArray(
                                  (selectedCatalog.supported_parameters as any).names,
                                ) &&
                                (selectedCatalog.supported_parameters as any).names.includes(param)
                              }
                              className="rounded border-gray-300"
                            />
                            <span className="text-xs">{param.replace(/_/g, ' ')}</span>
                          </label>
                        ))}
                      </div>
                    </div>

                    {/* Default Values */}
                    <div>
                      <label className="block text-sm font-medium mb-2">
                        Default Parameter Values
                      </label>
                      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1.5">
                            Temperature (optional)
                          </label>
                          <div className="flex gap-2 items-center">
                            <input
                              type="number"
                              name="temperature"
                              step="0.01"
                              min="0"
                              max="2"
                              defaultValue={
                                (selectedCatalog.supported_parameters &&
                                  (selectedCatalog.supported_parameters as any).default
                                    ?.temperature) ||
                                ''
                              }
                              placeholder="e.g., 0.7"
                              className="flex-1 min-w-0 px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                            />
                            <button
                              type="button"
                              onClick={(e) => {
                                const input = e.currentTarget
                                  .previousElementSibling as HTMLInputElement;
                                if (input) input.value = '';
                              }}
                              className="flex-shrink-0 w-9 h-9 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent border rounded-md transition-colors"
                              title="Clear value"
                            >
                              ✕
                            </button>
                          </div>
                        </div>
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1.5">
                            Top P (optional)
                          </label>
                          <div className="flex gap-2 items-center">
                            <input
                              type="number"
                              name="top_p"
                              step="0.01"
                              min="0"
                              max="1"
                              defaultValue={
                                (selectedCatalog.supported_parameters &&
                                  (selectedCatalog.supported_parameters as any).default?.top_p) ||
                                ''
                              }
                              placeholder="e.g., 0.9"
                              className="flex-1 min-w-0 px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                            />
                            <button
                              type="button"
                              onClick={(e) => {
                                const input = e.currentTarget
                                  .previousElementSibling as HTMLInputElement;
                                if (input) input.value = '';
                              }}
                              className="flex-shrink-0 w-9 h-9 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent border rounded-md transition-colors"
                              title="Clear value"
                            >
                              ✕
                            </button>
                          </div>
                        </div>
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1.5">
                            Top K (optional)
                          </label>
                          <div className="flex gap-2 items-center">
                            <input
                              type="number"
                              name="top_k"
                              step="1"
                              min="0"
                              defaultValue={
                                (selectedCatalog.supported_parameters &&
                                  (selectedCatalog.supported_parameters as any).default?.top_k) ||
                                ''
                              }
                              placeholder="e.g., 40"
                              className="flex-1 min-w-0 px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                            />
                            <button
                              type="button"
                              onClick={(e) => {
                                const input = e.currentTarget
                                  .previousElementSibling as HTMLInputElement;
                                if (input) input.value = '';
                              }}
                              className="flex-shrink-0 w-9 h-9 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent border rounded-md transition-colors"
                              title="Clear value"
                            >
                              ✕
                            </button>
                          </div>
                        </div>
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1.5">
                            Presence Penalty (optional)
                          </label>
                          <div className="flex gap-2 items-center">
                            <input
                              type="number"
                              name="presence_penalty"
                              step="0.01"
                              min="-2"
                              max="2"
                              defaultValue={
                                (selectedCatalog.supported_parameters &&
                                  (selectedCatalog.supported_parameters as any).default
                                    ?.presence_penalty) ||
                                ''
                              }
                              placeholder="e.g., 0.0"
                              className="flex-1 min-w-0 px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                            />
                            <button
                              type="button"
                              onClick={(e) => {
                                const input = e.currentTarget
                                  .previousElementSibling as HTMLInputElement;
                                if (input) input.value = '';
                              }}
                              className="flex-shrink-0 w-9 h-9 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent border rounded-md transition-colors"
                              title="Clear value"
                            >
                              ✕
                            </button>
                          </div>
                        </div>
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1.5">
                            Repetition Penalty (optional)
                          </label>
                          <div className="flex gap-2 items-center">
                            <input
                              type="number"
                              name="repetition_penalty"
                              step="0.01"
                              min="0"
                              defaultValue={
                                (selectedCatalog.supported_parameters &&
                                  (selectedCatalog.supported_parameters as any).default
                                    ?.repetition_penalty) ||
                                ''
                              }
                              placeholder="e.g., 1.0"
                              className="flex-1 min-w-0 px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                            />
                            <button
                              type="button"
                              onClick={(e) => {
                                const input = e.currentTarget
                                  .previousElementSibling as HTMLInputElement;
                                if (input) input.value = '';
                              }}
                              className="flex-shrink-0 w-9 h-9 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent border rounded-md transition-colors"
                              title="Clear value"
                            >
                              ✕
                            </button>
                          </div>
                        </div>
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1.5">
                            Frequency Penalty (optional)
                          </label>
                          <div className="flex gap-2 items-center">
                            <input
                              type="number"
                              name="frequency_penalty"
                              step="0.01"
                              min="-2"
                              max="2"
                              defaultValue={
                                (selectedCatalog.supported_parameters &&
                                  (selectedCatalog.supported_parameters as any).default
                                    ?.frequency_penalty) ||
                                ''
                              }
                              placeholder="e.g., 0"
                              className="flex-1 min-w-0 px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                            />
                            <button
                              type="button"
                              onClick={(e) => {
                                const input = e.currentTarget
                                  .previousElementSibling as HTMLInputElement;
                                if (input) input.value = '';
                              }}
                              className="flex-shrink-0 w-9 h-9 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent border rounded-md transition-colors"
                              title="Clear value"
                            >
                              ✕
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Capabilities */}
                <div className="border-t pt-4">
                  <h3 className="text-lg font-semibold mb-3">Capabilities</h3>
                  <div className="space-y-2">
                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        name="supports_images"
                        defaultChecked={selectedCatalog.supports_images}
                        className="rounded border-gray-300"
                      />
                      <div className="flex items-center gap-1.5">
                        <ImageIcon className="w-4 h-4 text-purple-600" />
                        <span className="text-sm">Vision Support</span>
                      </div>
                    </label>

                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        name="supports_embeddings"
                        defaultChecked={selectedCatalog.supports_embeddings}
                        className="rounded border-gray-300"
                      />
                      <div className="flex items-center gap-1.5">
                        <Zap className="w-4 h-4 text-green-600" />
                        <span className="text-sm">Embeddings Support</span>
                      </div>
                    </label>

                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        name="supports_reasoning"
                        defaultChecked={selectedCatalog.supports_reasoning}
                        className="rounded border-gray-300"
                      />
                      <div className="flex items-center gap-1.5">
                        <Brain className="w-4 h-4 text-orange-600" />
                        <span className="text-sm">Reasoning Support</span>
                      </div>
                    </label>

                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        name="supports_instruct"
                        defaultChecked={selectedCatalog.supports_instruct}
                        className="rounded border-gray-300"
                      />
                      <div className="flex items-center gap-1.5">
                        <Brain className="w-4 h-4 text-purple-600" />
                        <span className="text-sm">Supports Instruct Backup</span>
                      </div>
                    </label>

                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        name="supports_audio"
                        defaultChecked={selectedCatalog.supports_audio}
                        className="rounded border-gray-300"
                      />
                      <div className="flex items-center gap-1.5">
                        <Mic className="w-4 h-4 text-blue-600" />
                        <span className="text-sm">Audio Support</span>
                      </div>
                    </label>

                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        name="supports_video"
                        defaultChecked={selectedCatalog.supports_video}
                        className="rounded border-gray-300"
                      />
                      <div className="flex items-center gap-1.5">
                        <Video className="w-4 h-4 text-red-600" />
                        <span className="text-sm">Video Support</span>
                      </div>
                    </label>

                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        name="supports_tools"
                        defaultChecked={selectedCatalog.supports_tools ?? true}
                        className="rounded border-gray-300"
                      />
                      <div className="flex items-center gap-1.5">
                        <Wrench className="w-4 h-4 text-purple-600" />
                        <span className="text-sm">Tool/Function Calling</span>
                      </div>
                    </label>

                    <label className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        name="supports_browser"
                        defaultChecked={selectedCatalog.supports_browser}
                        className="rounded border-gray-300"
                      />
                      <div className="flex items-center gap-1.5">
                        <Globe className="w-4 h-4 text-cyan-600" />
                        <span className="text-sm">Browser Support</span>
                      </div>
                    </label>
                  </div>
                </div>

                {/* Notes and Moderation */}
                <div className="border-t pt-4">
                  <h3 className="text-lg font-semibold mb-3">Additional Information</h3>

                  <div className="space-y-3">
                    {/* Notes */}
                    <div>
                      <label className="block text-sm font-medium mb-1">Notes</label>
                      <textarea
                        name="notes"
                        defaultValue={selectedCatalog.notes || ''}
                        placeholder="Add any notes about this catalog..."
                        rows={3}
                        className="w-full px-3 py-2 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                      />
                    </div>

                    {/* Is Moderated */}
                    <div className="flex items-center gap-2">
                      <input
                        type="checkbox"
                        name="is_moderated"
                        id="is_moderated"
                        defaultChecked={selectedCatalog.is_moderated}
                        className="rounded border-gray-300"
                      />
                      <label htmlFor="is_moderated" className="text-sm cursor-pointer">
                        Is Moderated
                      </label>
                    </div>
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex justify-end gap-2 pt-4 mt-4 border-t">
                  <button
                    type="button"
                    onClick={() => setIsEditMode(false)}
                    className="px-4 py-2 border rounded-md hover:bg-accent transition-colors text-sm"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors text-sm font-medium"
                  >
                    Save Changes
                  </button>
                </div>
              </form>
            ) : (
              <div className="space-y-4">
                {/* Tab Navigation */}
                <div className="flex border-b">
                  <button
                    onClick={() => setActiveTab('details')}
                    className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                      activeTab === 'details'
                        ? 'border-primary text-primary'
                        : 'border-transparent text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    Details
                  </button>
                  <button
                    onClick={() => setActiveTab('prompts')}
                    className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors flex items-center gap-1.5 ${
                      activeTab === 'prompts'
                        ? 'border-primary text-primary'
                        : 'border-transparent text-muted-foreground hover:text-foreground'
                    }`}
                  >
                    <FileText className="w-4 h-4" />
                    Prompt Templates
                  </button>
                </div>

                {activeTab === 'prompts' ? (
                  <ModelPromptTemplatesTab modelCatalogId={selectedCatalog.id} />
                ) : (
                  <>
                    {/* Basic Info */}
                    <div>
                      <label className="block text-sm font-medium mb-1">Public ID</label>
                      <code className="block text-sm bg-muted px-3 py-2 rounded">
                        {selectedCatalog.id}
                      </code>
                    </div>

                    {selectedCatalog.family && (
                      <div>
                        <label className="block text-sm font-medium mb-1">Family</label>
                        <p className="text-sm">{selectedCatalog.family}</p>
                      </div>
                    )}

                    {selectedCatalog.notes && (
                      <div>
                        <label className="block text-sm font-medium mb-1">Notes</label>
                        <p className="text-sm whitespace-pre-wrap">{selectedCatalog.notes}</p>
                      </div>
                    )}

                    {selectedCatalog.architecture && (
                      <div>
                        <label className="block text-sm font-medium mb-1">Architecture</label>
                        {typeof selectedCatalog.architecture === 'string' ? (
                          <p className="text-sm">{selectedCatalog.architecture}</p>
                        ) : (
                          <div className="bg-muted/50 rounded-lg p-3 max-h-40 overflow-y-auto">
                            <pre className="text-xs whitespace-pre-wrap">
                              {JSON.stringify(selectedCatalog.architecture, null, 2)}
                            </pre>
                          </div>
                        )}
                      </div>
                    )}

                    {/* Capabilities */}
                    <div>
                      <label className="block text-sm font-medium mb-2">Capabilities</label>
                      <div className="grid grid-cols-2 gap-2">
                        <div className="flex items-center gap-2">
                          <div
                            className={`w-2 h-2 rounded-full ${selectedCatalog.supports_images ? 'bg-green-500' : 'bg-gray-300'}`}
                          />
                          <span className="text-sm">Vision Support</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <div
                            className={`w-2 h-2 rounded-full ${selectedCatalog.supports_embeddings ? 'bg-green-500' : 'bg-gray-300'}`}
                          />
                          <span className="text-sm">Embeddings</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <div
                            className={`w-2 h-2 rounded-full ${selectedCatalog.supports_reasoning ? 'bg-green-500' : 'bg-gray-300'}`}
                          />
                          <span className="text-sm">Reasoning</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <div
                            className={`w-2 h-2 rounded-full ${selectedCatalog.supports_audio ? 'bg-green-500' : 'bg-gray-300'}`}
                          />
                          <span className="text-sm">Audio</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <div
                            className={`w-2 h-2 rounded-full ${selectedCatalog.supports_video ? 'bg-green-500' : 'bg-gray-300'}`}
                          />
                          <span className="text-sm">Video</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <div
                            className={`w-2 h-2 rounded-full ${(selectedCatalog.supports_tools ?? true) ? 'bg-green-500' : 'bg-gray-300'}`}
                          />
                          <span className="text-sm">Tool Calling</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <div
                            className={`w-2 h-2 rounded-full ${selectedCatalog.supports_browser ? 'bg-green-500' : 'bg-gray-300'}`}
                          />
                          <span className="text-sm">Browser</span>
                        </div>
                      </div>
                    </div>

                    {/* Supported Parameters */}
                    {selectedCatalog.supported_parameters &&
                      typeof selectedCatalog.supported_parameters === 'object' && (
                        <div>
                          <label className="block text-sm font-medium mb-2">
                            Supported Parameters
                          </label>
                          <div className="bg-muted/50 rounded-lg p-3 max-h-60 overflow-y-auto">
                            <pre className="text-xs whitespace-pre-wrap">
                              {JSON.stringify(selectedCatalog.supported_parameters, null, 2)}
                            </pre>
                          </div>
                        </div>
                      )}

                    {/* Status Update */}
                    {selectedCatalog.status && (
                      <div>
                        <label className="block text-sm font-medium mb-2">Status</label>
                        <div className="flex gap-2">
                          <span className="px-3 py-1.5 bg-primary/10 text-primary rounded-md text-sm">
                            Current: {selectedCatalog.status}
                          </span>
                          <select
                            onChange={(e) => {
                              if (e.target.value) {
                                handleUpdateStatus(selectedCatalog.id, e.target.value);
                                setShowDetailsModal(false);
                                setSelectedCatalog(null);
                              }
                            }}
                            defaultValue=""
                            className="px-3 py-1.5 border rounded-md bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                          >
                            <option value="">Change status...</option>
                            {statuses
                              .filter((s) => s !== selectedCatalog.status)
                              .map((status) => (
                                <option key={status} value={status}>
                                  {status.charAt(0).toUpperCase() + status.slice(1)}
                                </option>
                              ))}
                          </select>
                        </div>
                      </div>
                    )}

                    {/* Timestamps */}
                    <div className="grid grid-cols-2 gap-4 pt-4 border-t">
                      {selectedCatalog.created_at && (
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1">
                            Created
                          </label>
                          <p className="text-sm">
                            {new Date(selectedCatalog.created_at).toLocaleString()}
                          </p>
                        </div>
                      )}
                      {selectedCatalog.updated_at && (
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1">
                            Updated
                          </label>
                          <p className="text-sm">
                            {new Date(selectedCatalog.updated_at).toLocaleString()}
                          </p>
                        </div>
                      )}
                    </div>
                  </>
                )}

                <div className="flex justify-end pt-4 mt-4 border-t">
                  <button
                    onClick={() => {
                      setShowDetailsModal(false);
                      setSelectedCatalog(null);
                      setActiveTab('details');
                    }}
                    className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
                  >
                    Close
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
