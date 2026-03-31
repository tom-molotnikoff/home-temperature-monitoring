import { get, post, put, del } from './Client';
import type { Dashboard, CreateDashboardRequest, UpdateDashboardRequest, ShareDashboardRequest } from '../types/dashboard';
import type { ApiMessage } from './Client';

export const list = () => get<Dashboard[]>('/dashboards/');
export const getById = (id: number) => get<Dashboard>(`/dashboards/${id}`);
export const create = (req: CreateDashboardRequest) => post<{ id: number }>('/dashboards/', req);
export const update = (id: number, req: UpdateDashboardRequest) => put<ApiMessage>(`/dashboards/${id}`, req);
export const remove = (id: number) => del<ApiMessage>(`/dashboards/${id}`);
export const share = (id: number, req: ShareDashboardRequest) => post<ApiMessage>(`/dashboards/${id}/share`, req);
export const setDefault = (id: number) => put<ApiMessage>(`/dashboards/${id}/default`);
