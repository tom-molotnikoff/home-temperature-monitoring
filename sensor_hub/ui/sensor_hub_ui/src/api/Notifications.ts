import { get, post } from './Client';

export type NotificationCategory = 'threshold_alert' | 'user_management' | 'config_change';
export type NotificationSeverity = 'info' | 'warning' | 'error';

export interface Notification {
  id: number;
  category: NotificationCategory;
  severity: NotificationSeverity;
  title: string;
  message: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface UserNotification {
  id: number;
  user_id: number;
  notification_id: number;
  is_read: boolean;
  is_dismissed: boolean;
  read_at?: string;
  dismissed_at?: string;
  notification: Notification;
}

export interface ChannelPreference {
  user_id?: number;
  category: NotificationCategory;
  email_enabled: boolean;
  inapp_enabled: boolean;
}

export interface UnreadCount {
  count: number;
}

export const getNotifications = async (limit = 50, offset = 0, includeDismissed = false) => {
  return get<UserNotification[]>(`/notifications/?limit=${limit}&offset=${offset}&include_dismissed=${includeDismissed}`);
};

export const getUnreadCount = async () => {
  return get<UnreadCount>('/notifications/unread-count');
};

export const markAsRead = async (notificationId: number) => {
  return post<{ message: string }>(`/notifications/${notificationId}/read`);
};

export const dismissNotification = async (notificationId: number) => {
  return post<{ message: string }>(`/notifications/${notificationId}/dismiss`);
};

export const bulkMarkAsRead = async () => {
  return post<{ message: string }>('/notifications/bulk/read');
};

export const bulkDismiss = async () => {
  return post<{ message: string }>('/notifications/bulk/dismiss');
};

export const getChannelPreferences = async () => {
  return get<ChannelPreference[]>('/notifications/preferences');
};

export const setChannelPreference = async (pref: ChannelPreference) => {
  return post<{ message: string }>('/notifications/preferences', pref);
};
