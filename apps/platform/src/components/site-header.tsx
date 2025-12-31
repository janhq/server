'use client';

import { useAuth } from '@/components/auth-provider';
import { Loader2, LogOut, Shield, User as UserIcon } from 'lucide-react';
import Link from 'next/link';
import { useEffect, useRef, useState } from 'react';

export function SiteHeader() {
  const { user, isLoading, login, logout } = useAuth();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsMenuOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 max-w-screen-2xl items-center">
        <div className="mr-4 hidden md:flex">
          <Link href="/" className="mr-6 flex items-center space-x-2">
            <img src="/jan_logo.svg" alt="Jan Logo" className="h-8 w-8" />
            <span className="hidden font-bold sm:inline-block">Jan Platform</span>
          </Link>
          <nav className="flex items-center gap-6 text-sm">
            <Link
              href="/docs/quickstart"
              className="transition-colors hover:text-foreground/80 text-foreground/60"
            >
              Docs
            </Link>
            <Link
              href="/docs/api-reference"
              className="transition-colors hover:text-foreground/80 text-foreground/60"
            >
              API Reference
            </Link>
          </nav>
        </div>
        <div className="flex flex-1 items-center justify-between space-x-2 md:justify-end">
          <div className="w-full flex-1 md:w-auto md:flex-none">{/* Search could go here */}</div>
          <nav className="flex items-center gap-2">
            {isLoading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : user ? (
              <div className="relative" ref={menuRef}>
                <button
                  onClick={() => setIsMenuOpen(!isMenuOpen)}
                  className="flex items-center gap-2 focus:outline-none"
                >
                  {user.picture ? (
                    <img
                      src={user.picture}
                      alt={user.name || 'User'}
                      className="h-8 w-8 rounded-full border border-border"
                    />
                  ) : (
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-200 text-gray-700">
                      <span className="text-xs font-bold">{user.name?.charAt(0) || 'U'}</span>
                    </div>
                  )}
                </button>

                {isMenuOpen && (
                  <div className="site-header-dropdown absolute right-0 mt-2 w-56 rounded-md border bg-popover p-1 shadow-md z-50">
                    <div className="px-2 py-1.5 text-sm font-semibold">
                      {user.name}
                      <div className="text-xs font-normal text-muted-foreground">{user.email}</div>
                    </div>
                    <div className="h-px bg-border my-1" />
                    <Link
                      href="/profile"
                      className="flex w-full items-center rounded-sm px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground"
                      onClick={() => setIsMenuOpen(false)}
                    >
                      <UserIcon className="mr-2 h-4 w-4" />
                      Profile
                    </Link>
                    {user.is_admin && (
                      <Link
                        href="/admin/users"
                        className="flex w-full items-center rounded-sm px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground"
                        onClick={() => setIsMenuOpen(false)}
                      >
                        <Shield className="mr-2 h-4 w-4" />
                        Admin
                      </Link>
                    )}
                    <div className="h-px bg-border my-1" />
                    <button
                      onClick={() => {
                        logout();
                        setIsMenuOpen(false);
                      }}
                      className="flex w-full items-center rounded-sm px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground text-red-500"
                    >
                      <LogOut className="mr-2 h-4 w-4" />
                      Log out
                    </button>
                  </div>
                )}
              </div>
            ) : (
              <button
                onClick={() => login()}
                className="inline-flex h-9 items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 cursor-pointer"
              >
                Sign in
              </button>
            )}
          </nav>
        </div>
      </div>
    </header>
  );
}
