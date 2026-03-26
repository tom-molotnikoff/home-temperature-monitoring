import { get, post, patch, del, type ApiMessage } from './Client';

export type ApiKey = {
  id: number;
  name: string;
  key_prefix: string;
  user_id: number;
  expires_at: string | null;
  revoked: boolean;
  last_used_at: string | null;
  created_at: string;
  updated_at: string;
};

export type CreateApiKeyRequest = {
  name: string;
  expires_at?: string;
};

export type CreateApiKeyResponse = {
  key: string;
  message: string;
};

export const listApiKeys = () => get<ApiKey[]>('/api-keys/');
export const createApiKey = (req: CreateApiKeyRequest) => post<CreateApiKeyResponse>('/api-keys/', req);
export const updateApiKeyExpiry = (id: number, expiresAt: string | null) =>
  patch<ApiMessage>(`/api-keys/${id}/expiry`, { expires_at: expiresAt });
export const revokeApiKey = (id: number) => post<ApiMessage>(`/api-keys/${id}/revoke`, {});
export const deleteApiKey = (id: number) => del<ApiMessage>(`/api-keys/${id}`);
