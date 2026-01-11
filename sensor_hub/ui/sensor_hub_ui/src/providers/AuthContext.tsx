import { createContext, useContext } from 'react';

export type AuthUser = { id: number; username: string; email?: string; roles: string[]; permissions?: string[] } | null;

export const AuthContext = createContext<{ user: AuthUser | undefined; refresh: () => Promise<void> }>({ user: undefined, refresh: async () => {} });

export const useAuth = () => useContext(AuthContext);
