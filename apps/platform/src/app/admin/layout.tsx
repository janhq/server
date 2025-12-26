'use client';

import { useAuth } from '@/components/auth-provider';
import { createAdminAPIClient } from '@/lib/admin/api';
import { getSharedAuthService } from '@/lib/auth/service';
import { Box, FileText, LayoutDashboard, Loader2, Shield, Users, Wrench } from 'lucide-react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const { user, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const [isAdmin, setIsAdmin] = useState<boolean | null>(null);
  const [isChecking, setIsChecking] = useState(true);

  useEffect(() => {
    async function checkAdminStatus() {
      if (authLoading) return;

      if (!user) {
        router.push('/auth/keycloak');
        return;
      }

      try {
        const authService = getSharedAuthService();
        const token = await authService.getValidAccessToken();

        if (!token) {
          router.push('/auth/keycloak');
          return;
        }

        const adminClient = createAdminAPIClient(token);
        const adminStatus = await adminClient.checkIsAdmin();

        setIsAdmin(adminStatus);

        if (!adminStatus) {
          router.push('/docs');
        }
      } catch (error) {
        console.error('Failed to check admin status:', error);
        setIsAdmin(false);
        router.push('/docs');
      } finally {
        setIsChecking(false);
      }
    }

    checkAdminStatus();
  }, [user, authLoading, router]);

  if (authLoading || isChecking || isAdmin === null) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-primary" />
          <p className="text-sm text-muted-foreground">Verifying admin access...</p>
        </div>
      </div>
    );
  }

  if (!isAdmin) {
    return null;
  }

  const navItems = [
    {
      title: 'Overview',
      href: '/admin',
      icon: LayoutDashboard,
    },
    {
      title: 'User Management',
      href: '/admin/users',
      icon: Users,
      children: [
        { title: 'Users', href: '/admin/users' },
        { title: 'Feature Flags', href: '/admin/users/feature-flags' },
      ],
    },
    {
      title: 'Model Management',
      href: '/admin/models',
      icon: Box,
      children: [
        { title: 'Providers', href: '/admin/models/providers' },
        { title: 'Provider Models', href: '/admin/models/provider-models' },
        { title: 'Model Catalogs', href: '/admin/models/catalogs' },
      ],
    },
    {
      title: 'Prompt Templates',
      href: '/admin/prompt-templates',
      icon: FileText,
    },
    {
      title: 'MCP Tools',
      href: '/admin/mcp-tools',
      icon: Wrench,
    },
  ];

  return (
    <div className="flex min-h-screen">
      {/* Admin Sidebar */}
      <aside className="w-64 bg-card border-r border-border fixed h-full pt-14 overflow-y-auto">
        <div className="p-4">
          <div className="flex items-center gap-2 mb-6 px-2">
            <Shield className="w-5 h-5 text-primary" />
            <h2 className="font-semibold text-lg">Admin Panel</h2>
          </div>

          <nav className="space-y-1">
            {navItems.map((item) => {
              const isActive = pathname === item.href;
              const Icon = item.icon;

              return (
                <div key={item.href}>
                  <Link
                    href={item.href}
                    className={`flex items-center gap-3 px-3 py-2 rounded-md transition-colors ${
                      isActive
                        ? 'bg-primary text-primary-foreground'
                        : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                    }`}
                  >
                    <Icon className="w-4 h-4" />
                    <span className="text-sm font-medium">{item.title}</span>
                  </Link>

                  {item.children && (
                    <div className="ml-7 mt-1 space-y-1">
                      {item.children.map((child) => {
                        const isChildActive = pathname === child.href;
                        return (
                          <Link
                            key={child.href}
                            href={child.href}
                            className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-xs transition-colors ${
                              isChildActive
                                ? 'bg-primary/10 text-primary font-medium'
                                : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                            }`}
                          >
                            {child.title}
                          </Link>
                        );
                      })}
                    </div>
                  )}
                </div>
              );
            })}
          </nav>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 ml-64 pt-14">
        <div className="p-6 max-w-7xl mx-auto">{children}</div>
      </main>
    </div>
  );
}
