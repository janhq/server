'use client';

import { createAdminAPIClient } from '@/lib/admin/api';
import type { MCPTool, UpdateMCPToolRequest } from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
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
import { Loader2, AlertCircle } from 'lucide-react';
import type { ChangeEvent } from 'react';
import { useEffect, useState } from 'react';

interface MCPToolModalProps {
  open: boolean;
  onClose: (refresh?: boolean) => void;
  tool?: MCPTool | null;
}

export function MCPToolModal({ open, onClose, tool }: MCPToolModalProps) {
  const [loading, setLoading] = useState(false);
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState('');
  const [isActive, setIsActive] = useState(true);
  const [disallowedKeywords, setDisallowedKeywords] = useState('');

  const categories = ['search', 'scrape', 'file_search', 'code_execution', 'memory'];

  useEffect(() => {
    if (tool) {
      setDescription(tool.description || '');
      setCategory(tool.category);
      setIsActive(tool.is_active);
      setDisallowedKeywords(tool.disallowed_keywords?.join('\n') || '');
    } else {
      resetForm();
    }
  }, [tool, open]);

  function resetForm() {
    setDescription('');
    setCategory('');
    setIsActive(true);
    setDisallowedKeywords('');
  }

  async function handleSubmit() {
    if (!tool) return;

    // Validation
    if (!description.trim()) {
      alert('Description is required');
      return;
    }

    // Parse disallowed keywords (one per line)
    const keywordsArray = disallowedKeywords
      .split('\n')
      .map(k => k.trim())
      .filter(k => k.length > 0);

    // Validate regex patterns
    // Note: Go regex supports inline flags like (?i) for case-insensitive matching
    // JavaScript RegExp doesn't support these, so we strip them for validation only
    const goInlineFlags = /^\(\?[imsU]+\)/;
    for (const pattern of keywordsArray) {
      try {
        // Remove Go-style inline flags for JS validation
        const jsPattern = pattern.replace(goInlineFlags, '');
        new RegExp(jsPattern);
      } catch (error) {
        alert(`Invalid regex pattern: ${pattern}`);
        return;
      }
    }

    setLoading(true);

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) throw new Error('No authentication token');

      const client = createAdminAPIClient(token);

      const updateData: UpdateMCPToolRequest = {
        description: description.trim(),
        category,
        is_active: isActive,
        disallowed_keywords: keywordsArray,
      };

      await client.mcpTools.updateMCPTool(tool.public_id, updateData);
      alert('Tool updated successfully');
      onClose(true);
    } catch (error) {
      console.error('Failed to save tool:', error);
      alert('Failed to save tool');
    } finally {
      setLoading(false);
    }
  }

  if (!tool) return null;

  return (
    <Dialog open={open} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Edit MCP Tool</DialogTitle>
          <DialogDescription>
            Update tool description, category, and keyword filters
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Read-only fields */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label className="text-muted-foreground">Name (read-only)</Label>
              <Input value={tool.name} disabled className="bg-muted" />
            </div>
            <div className="space-y-2">
              <Label className="text-muted-foreground">Tool Key (read-only)</Label>
              <Input value={tool.tool_key} disabled className="bg-muted" />
            </div>
          </div>

          {/* Editable fields */}
          <div className="space-y-2">
            <Label htmlFor="description">Description *</Label>
            <textarea
              id="description"
              placeholder="Enter tool description..."
              value={description}
              onChange={(e: ChangeEvent<HTMLTextAreaElement>) => setDescription(e.target.value)}
              rows={3}
              className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            />
            <p className="text-xs text-muted-foreground">
              This description is shown in the MCP tools/list response
            </p>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="category">Category</Label>
              <Select value={category} onValueChange={setCategory}>
                <SelectTrigger>
                  <SelectValue placeholder="Select category" />
                </SelectTrigger>
                <SelectContent>
                  {categories.map((cat) => (
                    <SelectItem key={cat} value={cat}>
                      {cat.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Status</Label>
              <div className="flex items-center space-x-2 pt-2">
                <button
                  type="button"
                  role="switch"
                  aria-checked={isActive}
                  onClick={() => setIsActive(!isActive)}
                  className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background ${isActive ? 'bg-primary' : 'bg-input'}`}
                >
                  <span
                    className={`pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform ${isActive ? 'translate-x-5' : 'translate-x-0'}`}
                  />
                </button>
                <span className="text-sm">
                  {isActive ? 'Active' : 'Inactive'}
                </span>
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="disallowedKeywords">Disallowed Keywords (Regex Patterns)</Label>
            <textarea
              id="disallowedKeywords"
              placeholder="Enter regex patterns, one per line...&#10;Example: (?i)menlo&#10;Example: (?i)competitor\s*name"
              value={disallowedKeywords}
              onChange={(e: ChangeEvent<HTMLTextAreaElement>) => setDisallowedKeywords(e.target.value)}
              rows={4}
              className="flex min-h-[100px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            />
            <div className="flex items-start gap-2 text-xs text-muted-foreground">
              <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
              <p>
                Search results matching these patterns will be filtered out.
                Use regex patterns (e.g., <code className="bg-muted px-1">(?i)menlo</code> for case-insensitive match).
                Invalid patterns will be ignored.
              </p>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onClose()} disabled={loading}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={loading}>
            {loading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
            Save Changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
