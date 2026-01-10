import { post, get, put, del, patch } from './Client';

export type User = { id: number; username: string; email?: string; roles: string[]; must_change_password: boolean };

export const User: unknown = undefined;

export const createUser = async (payload: { username: string; email?: string; password: string; roles: string[] }) => {
  return post<{ id: number }>("/users/", payload);
};

export const listUsers = async () => {
  return get<User[]>("/users/");
};

export const changePassword = async (userId: number, newPassword: string) => {
  return put('/users/password', { user_id: userId, new_password: newPassword });
};

export const deleteUser = async (userId: number) => {
  return del(`/users/${userId}`);
};

export const setMustChange = async (userId: number, mustChange: boolean) => {
  return patch(`/users/${userId}/must_change`, { must_change: mustChange });
};

export const setUserRoles = async (userId: number, roles: string[]) => {
  return post(`/users/${userId}/roles`, { roles });
};
