import { useEffect, useState, useCallback } from 'react';
import { type ApiKey, listApiKeys } from '../api/ApiKeys';
import { useAuth } from '../providers/AuthContext';
import { logger } from '../tools/logger';

export function useApiKeys() {
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [loaded, setLoaded] = useState(false);
  const { user } = useAuth();

  const refresh = useCallback(async () => {
    setLoaded(false);
    try {
      const keys = await listApiKeys();
      setApiKeys(keys);
    } catch (err) {
      logger.error('Failed to load API keys', err);
      setApiKeys([]);
    } finally {
      setLoaded(true);
    }
  }, []);

  useEffect(() => {
    if (user === undefined || user === null) return;
    refresh();
  }, [user, refresh]);

  return { apiKeys, loaded, refresh };
}
