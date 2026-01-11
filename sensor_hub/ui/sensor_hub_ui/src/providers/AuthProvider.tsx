import React, { useEffect, useState } from 'react';
import { me } from '../api/Auth';
import { AuthContext } from './AuthContext.tsx';

type User = { id: number; username: string; email?: string; roles: string[]; permissions?: string[] } | null;

export default function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | undefined>(undefined);

  const refresh = async () => {
    try {
      const res = await me();
      setUser(res.user || null);
    } catch {
      setUser(null);
    }
  }

  useEffect(() => { refresh(); }, []);

  return (
    <AuthContext.Provider value={{ user, refresh }}>
      {children}
    </AuthContext.Provider>
  );
}
