'use client';

import { createAdminAPIClient, FeatureFlag } from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
import { Flag, Loader2, Pencil, Plus, Search, Trash2, X } from 'lucide-react';
import { useEffect, useState } from 'react';

export default function FeatureFlagsPage() {
  const [flags, setFlags] = useState<FeatureFlag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [selectedFlag, setSelectedFlag] = useState<FeatureFlag | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Form states
  const [formData, setFormData] = useState({
    key: '',
    name: '',
    description: '',
    category: 'model_access',
  });

  useEffect(() => {
    loadFeatureFlags();
  }, []);

  async function loadFeatureFlags() {
    setIsLoading(true);
    setError(null);

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) {
        throw new Error('No access token');
      }

      const adminClient = createAdminAPIClient(token);
      const response = await adminClient.users.listFeatureFlags();
      setFlags(response.data || []);
    } catch (err) {
      console.error('Failed to load feature flags:', err);
      setError('Failed to load feature flags. Please try again.');
    } finally {
      setIsLoading(false);
    }
  }

  async function handleCreateFlag(e: React.FormEvent) {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) {
        throw new Error('No access token');
      }

      const adminClient = createAdminAPIClient(token);
      await adminClient.users.createFeatureFlag({
        key: formData.key,
        name: formData.name,
        description: formData.description || undefined,
      });

      await loadFeatureFlags();
      setShowCreateModal(false);
      resetForm();
    } catch (err) {
      console.error('Failed to create feature flag:', err);
      alert('Failed to create feature flag. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleUpdateFlag(e: React.FormEvent) {
    e.preventDefault();
    if (!selectedFlag) return;

    setIsSubmitting(true);

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) {
        throw new Error('No access token');
      }

      const adminClient = createAdminAPIClient(token);
      await adminClient.users.updateFeatureFlag(selectedFlag.id, {
        key: formData.key,
        name: formData.name,
        description: formData.description,
      });

      await loadFeatureFlags();
      setShowEditModal(false);
      setSelectedFlag(null);
      resetForm();
    } catch (err) {
      console.error('Failed to update feature flag:', err);
      alert('Failed to update feature flag. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleDeleteFlag() {
    if (!selectedFlag) return;

    setIsSubmitting(true);

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) {
        throw new Error('No access token');
      }

      const adminClient = createAdminAPIClient(token);
      await adminClient.users.deleteFeatureFlag(selectedFlag.id);

      await loadFeatureFlags();
      setShowDeleteModal(false);
      setSelectedFlag(null);
    } catch (err) {
      console.error('Failed to delete feature flag:', err);
      alert('Failed to delete feature flag. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  }

  function openEditModal(flag: FeatureFlag) {
    setSelectedFlag(flag);
    setFormData({
      key: flag.key,
      name: flag.name,
      description: flag.description || '',
      category: 'model_access',
    });
    setShowEditModal(true);
  }

  function openDeleteModal(flag: FeatureFlag) {
    setSelectedFlag(flag);
    setShowDeleteModal(true);
  }

  function resetForm() {
    setFormData({
      key: '',
      name: '',
      description: '',
      category: 'model_access',
    });
  }

  function closeCreateModal() {
    setShowCreateModal(false);
    resetForm();
  }

  function closeEditModal() {
    setShowEditModal(false);
    setSelectedFlag(null);
    resetForm();
  }

  function closeDeleteModal() {
    setShowDeleteModal(false);
    setSelectedFlag(null);
  }

  const filteredFlags = flags.filter((flag) => {
    if (!searchQuery) return true;
    const query = searchQuery.toLowerCase();
    return (
      flag.key.toLowerCase().includes(query) ||
      flag.name.toLowerCase().includes(query) ||
      (flag.description && flag.description.toLowerCase().includes(query))
    );
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-primary" />
          <p className="text-sm text-muted-foreground">Loading feature flags...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <Flag className="w-8 h-8" />
            Feature Flags
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage feature flags for access control and experimental features
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
        >
          <Plus className="w-4 h-4" />
          Create Flag
        </button>
      </div>

      {error && (
        <div className="p-4 bg-destructive/10 border border-destructive rounded-md">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      )}

      {/* Search Bar */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search by key, name, or description..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full pl-10 pr-4 py-2 bg-card border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="p-4 bg-card border border-border rounded-lg">
          <div className="text-sm text-muted-foreground">Total Flags</div>
          <div className="text-2xl font-bold">{flags.length}</div>
        </div>
        <div className="p-4 bg-card border border-border rounded-lg">
          <div className="text-sm text-muted-foreground">Search Results</div>
          <div className="text-2xl font-bold">{filteredFlags.length}</div>
        </div>
        <div className="p-4 bg-card border border-border rounded-lg">
          <div className="text-sm text-muted-foreground">Model Access Flags</div>
          <div className="text-2xl font-bold">
            {flags.filter((f) => f.key.includes('model') || f.key.includes('custom')).length}
          </div>
        </div>
      </div>

      {/* Feature Flags Table */}
      <div className="bg-card border border-border rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-muted/50 border-b border-border">
              <tr>
                <th className="text-left p-4 text-sm font-medium text-muted-foreground">Key</th>
                <th className="text-left p-4 text-sm font-medium text-muted-foreground">Name</th>
                <th className="text-left p-4 text-sm font-medium text-muted-foreground">
                  Description
                </th>
                <th className="text-left p-4 text-sm font-medium text-muted-foreground">Created</th>
                <th className="text-right p-4 text-sm font-medium text-muted-foreground">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              {filteredFlags.length === 0 ? (
                <tr>
                  <td colSpan={5} className="p-8 text-center text-muted-foreground">
                    {searchQuery
                      ? 'No feature flags found matching your search.'
                      : 'No feature flags yet. Create one to get started.'}
                  </td>
                </tr>
              ) : (
                filteredFlags.map((flag) => (
                  <tr key={flag.id} className="border-b border-border hover:bg-muted/30">
                    <td className="p-4">
                      <code className="px-2 py-1 bg-muted rounded text-sm font-mono">
                        {flag.key}
                      </code>
                    </td>
                    <td className="p-4">
                      <div className="font-medium">{flag.name}</div>
                    </td>
                    <td className="p-4">
                      <div className="text-sm text-muted-foreground max-w-md truncate">
                        {flag.description || '—'}
                      </div>
                    </td>
                    <td className="p-4">
                      <div className="text-sm text-muted-foreground">
                        {flag.created_at ? new Date(flag.created_at).toLocaleDateString() : '—'}
                      </div>
                    </td>
                    <td className="p-4">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => openEditModal(flag)}
                          className="p-2 text-muted-foreground hover:text-foreground hover:bg-muted rounded-md transition-colors"
                          title="Edit"
                        >
                          <Pencil className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => openDeleteModal(flag)}
                          className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-md transition-colors"
                          title="Delete"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg shadow-lg w-full max-w-md mx-4">
            <div className="flex items-center justify-between p-4 border-b border-border">
              <h3 className="text-lg font-semibold">Create Feature Flag</h3>
              <button
                onClick={closeCreateModal}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreateFlag} className="p-4 space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">
                  Key <span className="text-destructive">*</span>
                </label>
                <input
                  type="text"
                  required
                  value={formData.key}
                  onChange={(e) => setFormData({ ...formData, key: e.target.value })}
                  placeholder="experimental_models"
                  className="w-full px-3 py-2 bg-background border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary font-mono text-sm"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Use lowercase with underscores (e.g., experimental_models)
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">
                  Name <span className="text-destructive">*</span>
                </label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="Experimental Models"
                  className="w-full px-3 py-2 bg-background border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Access to experimental and beta models"
                  rows={3}
                  className="w-full px-3 py-2 bg-background border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary resize-none"
                />
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={closeCreateModal}
                  className="flex-1 px-4 py-2 border border-border rounded-md hover:bg-muted transition-colors"
                  disabled={isSubmitting}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    'Create Flag'
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Modal */}
      {showEditModal && selectedFlag && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg shadow-lg w-full max-w-md mx-4">
            <div className="flex items-center justify-between p-4 border-b border-border">
              <h3 className="text-lg font-semibold">Edit Feature Flag</h3>
              <button
                onClick={closeEditModal}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleUpdateFlag} className="p-4 space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">
                  Key <span className="text-destructive">*</span>
                </label>
                <input
                  type="text"
                  required
                  value={formData.key}
                  onChange={(e) => setFormData({ ...formData, key: e.target.value })}
                  className="w-full px-3 py-2 bg-background border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary font-mono text-sm"
                />
                <p className="text-xs text-destructive mt-1">
                  ⚠️ Changing the key will break existing references
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">
                  Name <span className="text-destructive">*</span>
                </label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 bg-background border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows={3}
                  className="w-full px-3 py-2 bg-background border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-primary resize-none"
                />
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={closeEditModal}
                  className="flex-1 px-4 py-2 border border-border rounded-md hover:bg-muted transition-colors"
                  disabled={isSubmitting}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Updating...
                    </>
                  ) : (
                    'Update Flag'
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteModal && selectedFlag && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg shadow-lg w-full max-w-md mx-4">
            <div className="flex items-center justify-between p-4 border-b border-border">
              <h3 className="text-lg font-semibold text-destructive">Delete Feature Flag</h3>
              <button
                onClick={closeDeleteModal}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-4 space-y-4">
              <p className="text-sm">
                Are you sure you want to delete the feature flag{' '}
                <code className="px-2 py-1 bg-muted rounded font-mono">{selectedFlag.key}</code>?
              </p>
              <p className="text-sm text-muted-foreground">
                Models with this feature flag requirement will have it set to NULL. This action
                cannot be undone.
              </p>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={closeDeleteModal}
                  className="flex-1 px-4 py-2 border border-border rounded-md hover:bg-muted transition-colors"
                  disabled={isSubmitting}
                >
                  Cancel
                </button>
                <button
                  onClick={handleDeleteFlag}
                  className="flex-1 px-4 py-2 bg-destructive text-destructive-foreground rounded-md hover:bg-destructive/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Deleting...
                    </>
                  ) : (
                    'Delete Flag'
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
