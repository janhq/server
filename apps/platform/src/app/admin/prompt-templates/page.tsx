'use client';

import { createAdminAPIClient } from '@/lib/admin/api';
import type { PromptTemplate } from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Copy, Edit, Eye, Loader2, MoreHorizontal, Plus, Search, Trash, FileText } from 'lucide-react';
import { useEffect, useState } from 'react';
import { PromptTemplateModal } from './components/prompt-template-modal';
import { PreviewModal } from './components/preview-modal';

export default function PromptTemplatesPage() {
  const [templates, setTemplates] = useState<PromptTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isPreviewOpen, setIsPreviewOpen] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<PromptTemplate | null>(null);

  const categories = ['orchestration', 'system', 'tool', 'reasoning'];

  useEffect(() => {
    loadTemplates();
  }, [categoryFilter, statusFilter]);

  async function loadTemplates() {
    try {
      setLoading(true);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) {
        throw new Error('No authentication token');
      }

      const client = createAdminAPIClient(token);
      const params: any = {};

      if (categoryFilter && categoryFilter !== 'all') params.category = categoryFilter;
      if (statusFilter && statusFilter !== 'all') params.is_active = statusFilter === 'active';
      if (searchTerm) params.search = searchTerm;

      const response = await client.promptTemplates.listPromptTemplates(params);
      setTemplates(response.data || []);
    } catch (error) {
      console.error('Failed to load prompt templates:', error);
      alert('Failed to load prompt templates');
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(template: PromptTemplate) {
    if (template.is_system) {
      alert('System templates cannot be deleted, only edited');
      return;
    }

    if (!confirm(`Are you sure you want to delete "${template.name}"? This action cannot be undone.`)) {
      return;
    }

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) throw new Error('No authentication token');

      const client = createAdminAPIClient(token);
      await client.promptTemplates.deletePromptTemplate(template.public_id);

      alert('Template deleted successfully');
      loadTemplates();
    } catch (error) {
      console.error('Failed to delete template:', error);
      alert('Failed to delete template');
    }
  }

  async function handleDuplicate(template: PromptTemplate) {
    const newName = `${template.name} - Copy`;

    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) throw new Error('No authentication token');

      const client = createAdminAPIClient(token);
      await client.promptTemplates.duplicatePromptTemplate(template.public_id, {
        new_name: newName,
      });

      alert('Template duplicated successfully');
      loadTemplates();
    } catch (error) {
      console.error('Failed to duplicate template:', error);
      alert('Failed to duplicate template');
    }
  }

  async function handleToggleActive(template: PromptTemplate) {
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) throw new Error('No authentication token');

      const client = createAdminAPIClient(token);
      
      if (template.is_active) {
        await client.promptTemplates.deactivateTemplate(template.public_id);
      } else {
        await client.promptTemplates.activateTemplate(template.public_id);
      }

      alert(`Template ${template.is_active ? 'deactivated' : 'activated'} successfully`);
      loadTemplates();
    } catch (error) {
      console.error('Failed to toggle template:', error);
      alert('Failed to update template status');
    }
  }

  function handleEdit(template: PromptTemplate) {
    setSelectedTemplate(template);
    setIsModalOpen(true);
  }

  function handlePreview(template: PromptTemplate) {
    setSelectedTemplate(template);
    setIsPreviewOpen(true);
  }

  function handleAddNew() {
    setSelectedTemplate(null);
    setIsModalOpen(true);
  }

  function handleModalClose(refresh?: boolean) {
    setIsModalOpen(false);
    setSelectedTemplate(null);
    if (refresh) {
      loadTemplates();
    }
  }

  const filteredTemplates = templates.filter((template) =>
    searchTerm
      ? template.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        template.template_key.toLowerCase().includes(searchTerm.toLowerCase())
      : true
  );

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Prompt Templates</h1>
          <p className="text-muted-foreground mt-1">
            Manage reusable prompt templates for orchestration features
          </p>
        </div>
        <Button onClick={handleAddNew}>
          <Plus className="w-4 h-4 mr-2" />
          Add Template
        </Button>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search by name or key..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-9"
          />
        </div>
        <Select value={categoryFilter} onValueChange={setCategoryFilter}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="All Categories" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Categories</SelectItem>
            {categories.map((cat) => (
              <SelectItem key={cat} value={cat}>
                {cat.charAt(0).toUpperCase() + cat.slice(1)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder="All Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="inactive">Inactive</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Table */}
      {loading ? (
        <div className="flex justify-center items-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-primary" />
        </div>
      ) : (
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Template Key</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>System</TableHead>
                <TableHead>Version</TableHead>
                <TableHead>Last Updated</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredTemplates.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-12 text-muted-foreground">
                    <FileText className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>No prompt templates found</p>
                  </TableCell>
                </TableRow>
              ) : (
                filteredTemplates.map((template) => (
                  <TableRow key={template.public_id}>
                    <TableCell className="font-medium">{template.name}</TableCell>
                    <TableCell>
                      <code className="text-xs bg-muted px-2 py-1 rounded">
                        {template.template_key}
                      </code>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {template.category}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={template.is_active ? 'default' : 'secondary'}>
                        {template.is_active ? 'Active' : 'Inactive'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {template.is_system ? (
                        <Badge variant="destructive">System</Badge>
                      ) : (
                        <Badge variant="outline">Custom</Badge>
                      )}
                    </TableCell>
                    <TableCell>v{template.version}</TableCell>
                    <TableCell>
                      {new Date(template.updated_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreHorizontal className="w-4 h-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handlePreview(template)}>
                            <Eye className="w-4 h-4 mr-2" />
                            Preview
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleEdit(template)}>
                            <Edit className="w-4 h-4 mr-2" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleDuplicate(template)}>
                            <Copy className="w-4 h-4 mr-2" />
                            Duplicate
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem onClick={() => handleToggleActive(template)}>
                            {template.is_active ? 'Deactivate' : 'Activate'}
                          </DropdownMenuItem>
                          {!template.is_system && (
                            <>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                onClick={() => handleDelete(template)}
                                className="text-destructive"
                              >
                                <Trash className="w-4 h-4 mr-2" />
                                Delete
                              </DropdownMenuItem>
                            </>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Modals */}
      <PromptTemplateModal
        open={isModalOpen}
        onClose={handleModalClose}
        template={selectedTemplate}
      />

      <PreviewModal
        open={isPreviewOpen}
        onClose={() => {
          setIsPreviewOpen(false);
          setSelectedTemplate(null);
        }}
        template={selectedTemplate}
      />
    </div>
  );
}
