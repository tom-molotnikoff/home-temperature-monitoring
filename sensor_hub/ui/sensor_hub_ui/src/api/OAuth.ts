import { get, post } from './Client';

export interface OAuthStatus {
  configured: boolean;
  needs_auth: boolean;
  token_valid: boolean;
  token_expiry?: string;
  refresher_active: boolean;
  last_refresh_at?: string;
  last_error?: string;
}

export interface OAuthAuthorizeResponse {
  auth_url: string;
  state: string;
}

export const getOAuthStatus = async (): Promise<OAuthStatus> => {
  return get<OAuthStatus>('/oauth/status');
};

export const getOAuthAuthorizeURL = async (): Promise<OAuthAuthorizeResponse> => {
  return get<OAuthAuthorizeResponse>('/oauth/authorize');
};

export const submitOAuthCode = async (code: string, state: string): Promise<void> => {
  return post<void>('/oauth/submit-code', { code, state });
};

export const reloadOAuthConfig = async (): Promise<void> => {
  return post<void>('/oauth/reload', {});
};
