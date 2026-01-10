import { get } from './Client';

export type Role = { id: number; name: string };

export const listRoles = async (): Promise<Role[]> => {
  return get<Role[]>('/roles/');
};

