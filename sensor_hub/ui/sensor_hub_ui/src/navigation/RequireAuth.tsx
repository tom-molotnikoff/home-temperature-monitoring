import { Navigate } from 'react-router';
import { useAuth } from '../providers/AuthContext.tsx';

export default function RequireAuth({ children }: { children: React.ReactElement }) {
  const { user } = useAuth();
  if (user === undefined) return null;
  if (user === null) return <Navigate to="/login" replace />;
  return children;
}
