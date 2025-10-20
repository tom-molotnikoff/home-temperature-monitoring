import {API_BASE} from '../environment/Environment';

export type ApiError = { status: number; message: string, error: string | null };
export type ApiMessage = { message: string };

async function request<T>(path: string, opts: RequestInit = {}): Promise<T> {
  const headers = { 'Content-Type': 'application/json', ...(opts.headers || {}) };
  const res = await fetch(`${API_BASE}${path}`, { ...opts, headers });
  const text = await res.text();
  const data = text ? JSON.parse(text) : null;
  if (!res.ok) {
    throw {status: res.status, message: data?.message || res.statusText, error: data?.error || null};
  }
  return data as T;
}

export const headResponse = async (path: string): Promise<Response> => {
  const headers = { 'Content-Type': 'application/json' };
  return fetch(`${API_BASE}${path}`, { method: 'HEAD', headers });
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
export const head = (path: string) => headRequest(path);