'use client';

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type {
  CreatePromptTemplateRequest,
  PromptTemplate,
  UpdatePromptTemplateRequest,
} from '@/lib/admin/api';
import { createAdminAPIClient } from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
import { Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';

interface PromptTemplateModalProps {
  open: boolean;
  onClose: (refresh?: boolean) => void;
  template?: PromptTemplate | null;
}

export function PromptTemplateModal({ open, onClose, template }: PromptTemplateModalProps) {
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('basic');
  const [name, setName] = useState('');
  const [templateKey, setTemplateKey] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState('orchestration');
  const [content, setContent] = useState('');
  const [variables, setVariables] = useState('[]');
  const [metadata, setMetadata] = useState('{}');
  const [isActive, setIsActive] = useState(true);

  const categories = ['orchestration', 'system', 'tool', 'reasoning'];

  useEffect(() => {
    if (template) {
      setName(template.name);
      setTemplateKey(template.template_key);
      setDescription(template.description || '');
      setCategory(template.category);
      setContent(template.content);
      setVariables(JSON.stringify(template.variables || [], null, 2));
      setMetadata(JSON.stringify(template.metadata || {}, null, 2));
      setIsActive(template.is_active);
    } else {
      resetForm();
    }
  }, [template, open]);

  function resetForm() {
    setName('');
    setTemplateKey('');
    setDescription('');
    setCategory('orchestration');
    setContent('');
    setVariables('[]');
    setMetadata('{}');
    setIsActive(true);
    setActiveTab('basic');
  }

  async function handleSubmit() {
    // Validation
    if (!name.trim()) {
      alert('Name is required');
      return;
    }

    if (!templateKey.trim()) {
      alert('Template Key is required');
      return;
    }

    if (!content.trim()) {
      alert('Content is required');
      return;
    }

    // Parse JSON fields
    let parsedVariables: string[];
    let parsedMetadata: Record<string, any>;

    try {
      parsedVariables = JSON.parse(variables);
      if (!Array.isArray(parsedVariables)) {
        throw new Error('Variables must be an array');
      }
    } catch (error) {
      alert('Variables must be valid JSON array');
      return;
    }

    try {
      parsedMetadata = JSON.parse(metadata);
      if (typeof parsedMetadata !== 'object' || Array.isArray(parsedMetadata)) {
        throw new Error('Metadata must be an object');
      }
    } catch (error) {
      alert('Metadata must be valid JSON object');
      return;
    }

    setLoading(true);

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) throw new Error('No authentication token');

      const client = createAdminAPIClient(token);

      if (template) {
        // Update existing template
        const updateData: UpdatePromptTemplateRequest = {
          name: name.trim(),
          description: description.trim() || undefined,
          content: content.trim(),
          variables: parsedVariables,
          metadata: parsedMetadata,
          is_active: isActive,
        };

        await client.promptTemplates.updatePromptTemplate(template.public_id, updateData);
        alert('Template updated successfully');
      } else {
        // Create new template
        const createData: CreatePromptTemplateRequest = {
          name: name.trim(),
          template_key: templateKey.trim(),
          description: description.trim() || undefined,
          category,
          content: content.trim(),
          variables: parsedVariables,
          metadata: parsedMetadata,
          is_active: isActive,
        };

        await client.promptTemplates.createPromptTemplate(createData);
        alert('Template created successfully');
      }

      onClose(true);
    } catch (error: any) {
      console.error('Failed to save template:', error);
      alert(error.message || 'Failed to save template');
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{template ? 'Edit Prompt Template' : 'Create Prompt Template'}</DialogTitle>
          <DialogDescription>
            {template
              ? 'Update the prompt template configuration'
              : 'Create a new reusable prompt template'}
          </DialogDescription>
        </DialogHeader>

        {/* Simple Tab Navigation */}
        <div className="border-b mb-4">
          <div className="flex gap-4">
            <button
              className={`pb-2 px-1 border-b-2 transition-colors ${
                activeTab === 'basic'
                  ? 'border-primary text-primary font-medium'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setActiveTab('basic')}
            >
              Basic Info
            </button>
            <button
              className={`pb-2 px-1 border-b-2 transition-colors ${
                activeTab === 'content'
                  ? 'border-primary text-primary font-medium'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setActiveTab('content')}
            >
              Content
            </button>
            <button
              className={`pb-2 px-1 border-b-2 transition-colors ${
                activeTab === 'advanced'
                  ? 'border-primary text-primary font-medium'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
              onClick={() => setActiveTab('advanced')}
            >
              Advanced
            </button>
          </div>
        </div>

        {/* Basic Info Tab */}
        {activeTab === 'basic' && (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name *</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Deep Research Orchestration"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="templateKey">Template Key *</Label>
              <Input
                id="templateKey"
                value={templateKey}
                onChange={(e) => setTemplateKey(e.target.value)}
                placeholder="e.g., deep_research"
                disabled={!!template}
              />
              {template && (
                <p className="text-xs text-muted-foreground">
                  Template key cannot be changed after creation
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="category">Category *</Label>
              <Select value={category} onValueChange={setCategory} disabled={!!template}>
                <SelectTrigger>
                  <SelectValue placeholder="Select category" />
                </SelectTrigger>
                <SelectContent>
                  {categories.map((cat) => (
                    <SelectItem key={cat} value={cat}>
                      {cat.charAt(0).toUpperCase() + cat.slice(1)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {template && (
                <p className="text-xs text-muted-foreground">
                  Category cannot be changed after creation
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Brief description of this template's purpose"
                rows={3}
                className="w-full px-3 py-2 text-sm border rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="isActive"
                checked={isActive}
                onChange={(e) => setIsActive(e.target.checked)}
                className="w-4 h-4"
              />
              <Label htmlFor="isActive">Active</Label>
            </div>
          </div>
        )}

        {/* Content Tab */}
        {activeTab === 'content' && (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="content">Prompt Content *</Label>
              <textarea
                id="content"
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder="Enter the prompt template content. Use {{variable}} syntax for placeholders."
                rows={15}
                className="w-full px-3 py-2 text-sm border rounded-md font-mono focus:outline-none focus:ring-2 focus:ring-primary"
              />
              <p className="text-xs text-muted-foreground">
                Use <code className="bg-muted px-1 py-0.5 rounded">{'{{variable}}'}</code> syntax
                for variables
              </p>
            </div>
          </div>
        )}

        {/* Advanced Tab */}
        {activeTab === 'advanced' && (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="variables">Variables (JSON Array)</Label>
              <textarea
                id="variables"
                value={variables}
                onChange={(e) => setVariables(e.target.value)}
                placeholder='["query", "context", "history"]'
                rows={6}
                className="w-full px-3 py-2 text-sm border rounded-md font-mono focus:outline-none focus:ring-2 focus:ring-primary"
              />
              <p className="text-xs text-muted-foreground">
                List of variable names used in the template
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="metadata">Metadata (JSON Object)</Label>
              <textarea
                id="metadata"
                value={metadata}
                onChange={(e) => setMetadata(e.target.value)}
                placeholder='{"author": "system", "tags": ["research", "deep"]}'
                rows={6}
                className="w-full px-3 py-2 text-sm border rounded-md font-mono focus:outline-none focus:ring-2 focus:ring-primary"
              />
              <p className="text-xs text-muted-foreground">Additional metadata in JSON format</p>
            </div>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onClose()} disabled={loading}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={loading}>
            {loading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
            {template ? 'Update' : 'Create'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
