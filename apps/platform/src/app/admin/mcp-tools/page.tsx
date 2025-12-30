'use client';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
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
import type { MCPTool } from '@/lib/admin/api';
import { createAdminAPIClient } from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
import { Edit, Loader2, MoreHorizontal, Search, Wrench } from 'lucide-react';
import { useEffect, useState } from 'react';
import { MCPToolModal } from './components/mcp-tool-modal';

export default function MCPToolsPage() {
  const [tools, setTools] = useState<MCPTool[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedTool, setSelectedTool] = useState<MCPTool | null>(null);

  const categories = ['search', 'scrape', 'file_search', 'code_execution', 'memory'];

  useEffect(() => {
    loadTools();
  }, [categoryFilter, statusFilter]);

  async function loadTools() {
    try {
      setLoading(true);
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) {
        throw new Error('No authentication token');
      }

      const client = createAdminAPIClient(token);
      const params: Record<string, any> = {};

      if (categoryFilter && categoryFilter !== 'all') params.category = categoryFilter;
      if (statusFilter && statusFilter !== 'all') params.is_active = statusFilter === 'active';
      if (searchTerm) params.search = searchTerm;

      const response = await client.mcpTools.listMCPTools(params);
      setTools(response.data || []);
    } catch (error) {
      console.error('Failed to load MCP tools:', error);
      alert('Failed to load MCP tools');
    } finally {
      setLoading(false);
    }
  }

  async function handleToggleActive(tool: MCPTool) {
    try {
      const authService = getSharedAuthService();
      const token = await authService.getValidAccessToken();

      if (!token) throw new Error('No authentication token');

      const client = createAdminAPIClient(token);

      if (tool.is_active) {
        await client.mcpTools.deactivateTool(tool.public_id);
      } else {
        await client.mcpTools.activateTool(tool.public_id);
      }

      alert(`Tool ${tool.is_active ? 'deactivated' : 'activated'} successfully`);
      loadTools();
    } catch (error) {
      console.error('Failed to toggle tool:', error);
      alert('Failed to update tool status');
    }
  }

  function handleEdit(tool: MCPTool) {
    setSelectedTool(tool);
    setIsModalOpen(true);
  }

  function handleModalClose(refresh?: boolean) {
    setIsModalOpen(false);
    setSelectedTool(null);
    if (refresh) {
      loadTools();
    }
  }

  const filteredTools = tools.filter((tool) =>
    searchTerm
      ? tool.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        tool.tool_key.toLowerCase().includes(searchTerm.toLowerCase())
      : true,
  );

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">MCP Tools</h1>
          <p className="text-muted-foreground mt-1">
            Manage MCP tool descriptions and keyword filters
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search by name or tool key..."
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
                {cat.replace('_', ' ').replace(/\b\w/g, (l) => l.toUpperCase())}
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
                <TableHead>Tool Key</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Keywords Filtered</TableHead>
                <TableHead>Last Updated</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredTools.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-12 text-muted-foreground">
                    <Wrench className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>No MCP tools found</p>
                  </TableCell>
                </TableRow>
              ) : (
                filteredTools.map((tool) => (
                  <TableRow key={tool.public_id}>
                    <TableCell className="font-medium">{tool.name}</TableCell>
                    <TableCell>
                      <code className="text-xs bg-muted px-2 py-1 rounded">{tool.tool_key}</code>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{tool.category}</Badge>
                    </TableCell>
                    <TableCell className="max-w-[300px] truncate">{tool.description}</TableCell>
                    <TableCell>
                      <Badge variant={tool.is_active ? 'default' : 'secondary'}>
                        {tool.is_active ? 'Active' : 'Inactive'}
                      </Badge>
                    </TableCell>
                    <TableCell>{tool.disallowed_keywords?.length || 0}</TableCell>
                    <TableCell>{new Date(tool.updated_at).toLocaleDateString()}</TableCell>
                    <TableCell className="text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreHorizontal className="w-4 h-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleEdit(tool)}>
                            <Edit className="w-4 h-4 mr-2" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem onClick={() => handleToggleActive(tool)}>
                            {tool.is_active ? 'Deactivate' : 'Activate'}
                          </DropdownMenuItem>
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

      {/* Modal */}
      <MCPToolModal open={isModalOpen} onClose={handleModalClose} tool={selectedTool} />
    </div>
  );
}
