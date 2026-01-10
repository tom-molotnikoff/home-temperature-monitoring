import {API_BASE} from '../environment/Environment';
import { getCsrfToken } from './Csrf';

export type ApiError = { status: number; message: string, error: string | null };
export type ApiMessage = { message: string };

async function request<T>(path: string, opts: RequestInit = {}): Promise<T> {
  const csrf = getCsrfToken();
  const headers: Record<string,string> = { 'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest', ...(opts.headers as Record<string,string> || {}) };
  if (csrf) {
    headers['X-CSRF-Token'] = csrf;
  }
  const url = `${API_BASE}${path}`;
  console.debug('[Client] request', opts.method || 'GET', url, 'headers=', headers);
  let res: Response;
  try {
    res = await fetch(url, { ...opts, headers, credentials: 'include' });
  } catch (err) {
    console.error('[Client] fetch failed', err, url);
    throw { status: 0, message: 'Network error', error: String(err) } as ApiError;
  }
  const text = await res.text();
  const data = text ? JSON.parse(text) : null;
  if (!res.ok) {
    throw {status: res.status, message: data?.message || res.statusText, error: data?.error || null};
  }
  return data as T;
}

export const headResponse = async (path: string): Promise<Response> => {
  const csrf = getCsrfToken();
  const headers: Record<string,string> = { 'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest' };
  if (csrf) headers['X-CSRF-Token'] = csrf;
  console.debug('[Client] HEAD', `${API_BASE}${path}`);
  return fetch(`${API_BASE}${path}`, { method: 'HEAD', headers, credentials: 'include' });
};

export const headRequest = async (path: string): Promise<boolean> => {
  const res = await headResponse(path);
  if (res.ok) return true;
  if (res.status === 404) return false;
  throw { status: res.status, message: res.statusText } as ApiError;
};

export const get = <T>(path: string) => request<T>(path, { method: 'GET' });
export const post = <T>(path: string, body?: unknown) => request<T>(path, { method: 'POST', body: JSON.stringify(body) });
export const put = <T>(path: string, body?: unknown) => request<T>(path, { method: 'PUT', body: JSON.stringify(body) });
export const del = <T>(path: string) => request<T>(path, { method: 'DELETE' });
export const patch = <T>(path: string, body?: unknown) => request<T>(path, { method: 'PATCH', body: JSON.stringify(body) });
export const head = (path: string) => headRequest(path);