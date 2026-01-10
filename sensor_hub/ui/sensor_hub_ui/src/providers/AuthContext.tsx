import { createContext, useContext } from 'react';

type User = { id: number; username: string; email?: string; roles: string[] } | null;

export const AuthContext = createContext<{ user: User | undefined; refresh: () => Promise<void> }>({ user: undefined, refresh: async () => {} });

export const useAuth = () => useContext(AuthContext);
