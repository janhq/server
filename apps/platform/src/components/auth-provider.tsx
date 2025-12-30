'use client';

import { AUTH_EVENTS } from '@/lib/auth/const';
import { getSharedAuthService, JanAuthService } from '@/lib/auth/service';
import { User } from '@/lib/auth/types';
import { useRouter } from 'next/navigation';
import { createContext, ReactNode, useContext, useEffect, useState } from 'react';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: () => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [authService] = useState<JanAuthService>(() => getSharedAuthService());
  const router = useRouter();

  useEffect(() => {
    let mounted = true;

    const initAuth = async () => {
      try {
        await authService.initialize();
        const currentUser = await authService.getCurrentUser();
        if (mounted) {
          setUser(currentUser);
        }
      } catch (error) {
        console.error('Auth initialization failed:', error);
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    };

    initAuth();

    // Subscribe to auth events
    const cleanup = authService.onAuthEvent(async (event) => {
      if (event.data.type === AUTH_EVENTS.LOGIN) {
        const currentUser = await authService.getCurrentUser(true);
        if (mounted) {
          setUser(currentUser);
          router.push('/profile');
        }
      } else if (event.data.type === AUTH_EVENTS.LOGOUT) {
        if (mounted) setUser(null);
      }
    });

    return () => {
      mounted = false;
      cleanup();
    };
  }, [authService, router]);

  const login = async () => {
    // For now, we only support Keycloak
    await authService.loginWithProvider('keycloak');
  };

  const logout = async () => {
    await authService.logout();
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
