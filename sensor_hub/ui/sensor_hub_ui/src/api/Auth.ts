import { post, get } from './Client';
import { setCsrfToken } from './Csrf';

export type LoginResponse = { must_change_password: boolean, csrf_token?: string };

export type UserType = { id: number; username: string; email?: string; roles: string[]; permissions?: string[] } | null;

export type authMeResponse = {
  user?: UserType,
  csrf_token?: string
}

export const login = async (username: string, password: string) => {
  const res = await post<LoginResponse>('/auth/login', { username, password });
  if (res.csrf_token) setCsrfToken(res.csrf_token);
  return res;
};

export const logout = async () => {
  setCsrfToken(null);
  return post('/auth/logout');
};

export const me = async () => {
  const res = await get<authMeResponse>('/auth/me');
  if (res.csrf_token) setCsrfToken(res.csrf_token);
  return res;
};
