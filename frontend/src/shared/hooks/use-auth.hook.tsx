import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';

interface User {
  login: string;
  avatar: string;
}

interface AuthContextValue {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Use fetch directly instead of api.service to avoid the 401 → redirect loop.
    // A 401 here simply means "not logged in", not an auth error to redirect for.
    fetch('/api/me')
      .then(res => (res.ok ? (res.json() as Promise<User>) : null))
      .then(data => setUser(data ?? null))
      .catch(() => setUser(null))
      .finally(() => setIsLoading(false));
  }, []);

  return <AuthContext value={{ user, isLoading, isAuthenticated: !!user }}>{children}</AuthContext>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
