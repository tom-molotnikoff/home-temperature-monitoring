import createClient from 'openapi-fetch';
import type { paths } from './schema';
import { getCsrfToken } from '../api/Csrf';

export const apiClient = createClient<paths>({
  baseUrl: import.meta.env.VITE_API_BASE || '/api',
  credentials: 'include',
});

apiClient.use({
  async onRequest({ request }) {
    const token = getCsrfToken();
    if (token) request.headers.set('X-CSRF-Token', token);
    request.headers.set('X-Requested-With', 'XMLHttpRequest');
    return request;
  },
});
