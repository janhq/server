'use client';

import {
  AssignTemplateRequest,
  createAdminAPIClient,
  EffectiveTemplatesResponse,
  ModelPromptTemplate,
  PromptTemplate,
} from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
import {
  AlertTriangle,
  Check,
  ChevronDown,
  FileText,
  Loader2,
  Plus,
  RefreshCw,
  Trash2,
  X,
} from 'lucide-react';
import { useEffect, useState } from 'react';

interface ModelPromptTemplatesTabProps {
  modelCatalogId: string;
}

// Common template keys
const TEMPLATE_KEYS = [
  { key: 'deep_research', label: 'Deep Research', description: 'Used for deep research mode' },
  { key: 'timing', label: 'Timing', description: 'Date and time context' },
  { key: 'memory', label: 'Memory', description: 'User memory injection' },
  { key: 'tool_instructions', label: 'Tool Instructions', description: 'Tool usage guidance' },
  { key: 'code_assistant', label: 'Code Assistant', description: 'Code assistance prompts' },
  { key: 'chain_of_thought', label: 'Chain of Thought', description: 'Step-by-step reasoning' },
  { key: 'user_profile', label: 'User Profile', description: 'User profile personalization' },
];

export default function ModelPromptTemplatesTab({ modelCatalogId }: ModelPromptTemplatesTabProps) {
  const [assignments, setAssignments] = useState<ModelPromptTemplate[]>([]);
  const [effective, setEffective] = useState<EffectiveTemplatesResponse | null>(null);
  const [availableTemplates, setAvailableTemplates] = useState<PromptTemplate[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [selectedTemplateKey, setSelectedTemplateKey] = useState<string>('');
  const [selectedTemplateId, setSelectedTemplateId] = useState<string>('');
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadData();
  }, [modelCatalogId]);

  async function loadData() {
    try {
      setIsLoading(true);
      setError(null);

      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) {
        setError('No authentication token');
        return;
      }

      const adminClient = createAdminAPIClient(token);

      // Load all data in parallel
      const [assignmentsRes, effectiveRes, templatesRes] = await Promise.all([
        adminClient.modelCatalogs.listModelPromptTemplates(modelCatalogId),
        adminClient.modelCatalogs.getEffectiveTemplates(modelCatalogId),
        adminClient.promptTemplates.listPromptTemplates({ limit: 100, is_active: true }),
      ]);

      setAssignments(assignmentsRes.data || []);
      setEffective(effectiveRes);
      setAvailableTemplates(templatesRes.data || []);
    } catch (err) {
      console.error('Failed to load model prompt templates:', err);
      setError('Failed to load prompt template assignments');
    } finally {
      setIsLoading(false);
    }
  }

  async function handleAssignTemplate() {
    if (!selectedTemplateKey || !selectedTemplateId) return;

    try {
      setIsSaving(true);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.modelCatalogs.assignPromptTemplate(modelCatalogId, {
        template_key: selectedTemplateKey,
        prompt_template_id: selectedTemplateId,
      });

      setShowAssignModal(false);
      setSelectedTemplateKey('');
      setSelectedTemplateId('');
      loadData();
    } catch (err) {
      console.error('Failed to assign template:', err);
      alert('Failed to assign template');
    } finally {
      setIsSaving(false);
    }
  }

  async function handleUnassignTemplate(templateKey: string) {
    if (!confirm(`Remove custom template for "${templateKey}"? The model will revert to the global default.`)) {
      return;
    }

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();
      if (!token) return;

      const adminClient = createAdminAPIClient(token);
      await adminClient.modelCatalogs.unassignPromptTemplate(modelCatalogId, templateKey);
      loadData();
    } catch (err) {
      console.error('Failed to unassign template:', err);
      alert('Failed to remove template assignment');
    }
  }

  function getSourceBadge(source: string) {
    switch (source) {
      case 'model_specific':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400">
            <Check className="w-3 h-3" />
            Custom
          </span>
        );
      case 'global_default':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300">
            Global Default
          </span>
        );
      case 'hardcoded':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400">
            <AlertTriangle className="w-3 h-3" />
            Hardcoded
          </span>
        );
      default:
        return null;
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-8 gap-2">
        <p className="text-destructive">{error}</p>
        <button
          onClick={loadData}
          className="px-3 py-1.5 text-sm border rounded-md hover:bg-accent flex items-center gap-1"
        >
          <RefreshCw className="w-4 h-4" />
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium">Prompt Template Overrides</h3>
          <p className="text-sm text-muted-foreground">
            Customize prompt templates for this model. Overrides will take precedence over global defaults.
          </p>
        </div>
        <button
          onClick={() => setShowAssignModal(true)}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
        >
          <Plus className="w-4 h-4" />
          Add Override
        </button>
      </div>

      {/* Effective Templates Table */}
      <div className="border rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-muted/50">
            <tr>
              <th className="text-left px-4 py-3 font-medium">Template Key</th>
              <th className="text-left px-4 py-3 font-medium">Active Template</th>
              <th className="text-left px-4 py-3 font-medium">Source</th>
              <th className="text-right px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {TEMPLATE_KEYS.map(({ key, label, description }) => {
              const effectiveTemplate = effective?.templates?.[key];
              const hasCustom = effectiveTemplate?.source === 'model_specific';

              return (
                <tr key={key} className="hover:bg-muted/30">
                  <td className="px-4 py-3">
                    <div>
                      <div className="font-medium">{label}</div>
                      <div className="text-xs text-muted-foreground">{description}</div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    {effectiveTemplate?.template ? (
                      <div className="flex items-center gap-2">
                        <FileText className="w-4 h-4 text-muted-foreground" />
                        <span>{effectiveTemplate.template.name}</span>
                      </div>
                    ) : (
                      <span className="text-muted-foreground italic">Not configured</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    {effectiveTemplate ? getSourceBadge(effectiveTemplate.source) : null}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-2">
                      {hasCustom ? (
                        <button
                          onClick={() => handleUnassignTemplate(key)}
                          className="p-1.5 text-destructive hover:bg-destructive/10 rounded-md"
                          title="Remove custom override"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      ) : (
                        <button
                          onClick={() => {
                            setSelectedTemplateKey(key);
                            setShowAssignModal(true);
                          }}
                          className="p-1.5 text-muted-foreground hover:bg-accent rounded-md"
                          title="Add custom override"
                        >
                          <Plus className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Info Text */}
      <p className="text-xs text-muted-foreground">
        <strong>Custom:</strong> Model uses a specific template override.{' '}
        <strong>Global Default:</strong> Model uses the system-wide template.{' '}
        <strong>Hardcoded:</strong> No template configured, using built-in fallback.
      </p>

      {/* Assign Template Modal */}
      {showAssignModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[60] p-4">
          <div className="bg-card rounded-lg border max-w-md w-full p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Assign Custom Template</h3>
              <button
                onClick={() => {
                  setShowAssignModal(false);
                  setSelectedTemplateKey('');
                  setSelectedTemplateId('');
                }}
                className="p-1 hover:bg-accent rounded-md"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-4">
              {/* Template Key Selection */}
              <div>
                <label className="block text-sm font-medium mb-1.5">Template Key</label>
                <div className="relative">
                  <select
                    value={selectedTemplateKey}
                    onChange={(e) => setSelectedTemplateKey(e.target.value)}
                    className="w-full px-3 py-2 border rounded-md bg-background text-sm appearance-none pr-10"
                  >
                    <option value="">Select template key...</option>
                    {TEMPLATE_KEYS.map(({ key, label }) => (
                      <option key={key} value={key}>
                        {label} ({key})
                      </option>
                    ))}
                  </select>
                  <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
                </div>
              </div>

              {/* Template Selection */}
              <div>
                <label className="block text-sm font-medium mb-1.5">Prompt Template</label>
                <div className="relative">
                  <select
                    value={selectedTemplateId}
                    onChange={(e) => setSelectedTemplateId(e.target.value)}
                    className="w-full px-3 py-2 border rounded-md bg-background text-sm appearance-none pr-10"
                  >
                    <option value="">Select a template...</option>
                    {availableTemplates.map((template) => (
                      <option key={template.public_id} value={template.public_id}>
                        {template.name} ({template.template_key})
                      </option>
                    ))}
                  </select>
                  <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  This template will override the global default for this model.
                </p>
              </div>

              {/* Actions */}
              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => {
                    setShowAssignModal(false);
                    setSelectedTemplateKey('');
                    setSelectedTemplateId('');
                  }}
                  className="px-4 py-2 text-sm border rounded-md hover:bg-accent"
                >
                  Cancel
                </button>
                <button
                  onClick={handleAssignTemplate}
                  disabled={!selectedTemplateKey || !selectedTemplateId || isSaving}
                  className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                  {isSaving && <Loader2 className="w-4 h-4 animate-spin" />}
                  Assign Template
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
