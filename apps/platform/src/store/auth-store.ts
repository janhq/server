import Cookies from 'js-cookie';
import { create } from 'zustand';

interface AuthUser {
  email: string;
  name?: string;
  provider?: 'google' | 'email';
}

interface AuthState {
  isLoggedIn: boolean;
  user: AuthUser | null;
  login: (userData: AuthUser) => void;
  logout: () => void;
}

const COOKIE_NAME = 'auth-storage';

// Helper to get initial state from cookie (only runs on client)
const getInitialState = (): Pick<AuthState, 'isLoggedIn' | 'user'> => {
  // Skip on server-side
  if (typeof window === 'undefined') {
    return { isLoggedIn: false, user: null };
  }

  const cookie = Cookies.get(COOKIE_NAME);
  if (cookie) {
    try {
      const parsed = JSON.parse(cookie);
      return {
        isLoggedIn: parsed.isLoggedIn || false,
        user: parsed.user || null,
      };
    } catch {
      return { isLoggedIn: false, user: null };
    }
  }
  return { isLoggedIn: false, user: null };
};

// Helper to save state to cookie
const saveStateToCookie = (state: Pick<AuthState, 'isLoggedIn' | 'user'>) => {
  Cookies.set(COOKIE_NAME, JSON.stringify(state));
};

export const useAuthStore = create<AuthState>()((set) => ({
  ...getInitialState(),
  login: (userData: AuthUser) => {
    const newState = {
      isLoggedIn: true,
      user: userData,
    };
    set(newState);
    saveStateToCookie(newState);
  },
  logout: () => {
    const newState = {
      isLoggedIn: false,
      user: null,
    };
    set(newState);
    Cookies.remove(COOKIE_NAME);

    // Redirect to login page
    if (typeof window !== 'undefined') {
      window.location.href = '/';
    }
  },
}));
