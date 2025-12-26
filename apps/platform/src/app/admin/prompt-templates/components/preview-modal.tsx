'use client';

import type { PromptTemplate } from '@/lib/admin/api';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';

interface PreviewModalProps {
  open: boolean;
  onClose: () => void;
  template: PromptTemplate | null;
}

export function PreviewModal({ open, onClose, template }: PreviewModalProps) {
  if (!template) return null;

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {template.name}
            <Badge variant={template.is_active ? 'default' : 'secondary'}>
              {template.is_active ? 'Active' : 'Inactive'}
            </Badge>
            {template.is_system && (
              <Badge variant="destructive">System</Badge>
            )}
          </DialogTitle>
          <DialogDescription>
            {template.description || 'No description provided'}
          </DialogDescription>
        </DialogHeader>

        <div className="max-h-[calc(90vh-200px)] overflow-y-auto">
          <div className="space-y-6 pr-4">
            {/* Metadata Section */}
            <div>
              <h3 className="text-sm font-semibold mb-3">Template Information</h3>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground">Template Key:</span>
                  <code className="ml-2 bg-muted px-2 py-1 rounded text-xs">
                    {template.template_key}
                  </code>
                </div>
                <div>
                  <span className="text-muted-foreground">Category:</span>
                  <Badge variant="outline" className="ml-2">
                    {template.category}
                  </Badge>
                </div>
                <div>
                  <span className="text-muted-foreground">Version:</span>
                  <span className="ml-2">v{template.version}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Last Updated:</span>
                  <span className="ml-2">
                    {new Date(template.updated_at).toLocaleDateString()}
                  </span>
                </div>
              </div>
            </div>

            <Separator />

            {/* Content Section */}
            <div>
              <h3 className="text-sm font-semibold mb-3">Prompt Content</h3>
              <div className="bg-muted rounded-lg p-4">
                <pre className="text-sm whitespace-pre-wrap font-mono leading-relaxed">
                  {template.content}
                </pre>
              </div>
            </div>

            {/* Variables Section */}
            {template.variables && template.variables.length > 0 && (
              <>
                <Separator />
                <div>
                  <h3 className="text-sm font-semibold mb-3">Variables</h3>
                  <div className="flex flex-wrap gap-2">
                    {template.variables.map((variable, index) => (
                      <code key={index} className="bg-muted px-3 py-1 rounded text-xs">
                        {'{{'}{variable}{'}}'}
                      </code>
                    ))}
                  </div>
                </div>
              </>
            )}

            {/* Metadata Section */}
            {template.metadata && Object.keys(template.metadata).length > 0 && (
              <>
                <Separator />
                <div>
                  <h3 className="text-sm font-semibold mb-3">Additional Metadata</h3>
                  <div className="bg-muted rounded-lg p-4">
                    <pre className="text-sm font-mono">
                      {JSON.stringify(template.metadata, null, 2)}
                    </pre>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
