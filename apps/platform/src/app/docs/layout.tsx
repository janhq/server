import { baseOptions } from '@/lib/layout.shared';
import { source } from '@/lib/source';
import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import './docs-fix.css';

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <div className="pt-4">
        <DocsLayout
          tree={source.pageTree}
          {...baseOptions()}
          sidebar={{
            enabled: true,
            collapsible: false,
            className: 'bg-gray border-0 top-14 h-[calc(100vh-3.5rem)]',
          }}
          nav={{
            enabled: false,
          }}
          containerProps={{
            className: 'relative',
          }}
        >
          <div className="relative">
            {children}
          </div>
        </DocsLayout>
      </div>
    </>
  );
}
