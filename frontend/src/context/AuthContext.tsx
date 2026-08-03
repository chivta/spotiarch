import { createContext, useCallback, useContext, useEffect, useState } from "react";
import type { ReactNode } from "react";
import { getMe, logout as logoutRequest } from "../api";
import type { User } from "../types/models";

interface AuthContextValue {
  user: User | null;
  loading: boolean;
  refresh: () => Promise<User>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const refresh = useCallback(async () => {
    const next = await getMe();
    setUser(next);
    return next;
  }, []);
  useEffect(() => { refresh().finally(() => setLoading(false)); }, [refresh]);
  const logout = useCallback(async () => {
    await logoutRequest();
    await refresh();
  }, [refresh]);
  return <AuthContext.Provider value={{ user, loading, refresh, logout }}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("AuthContext unavailable");
  return value;
}
