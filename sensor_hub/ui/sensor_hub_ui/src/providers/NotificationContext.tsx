import { createContext, useContext } from 'react';
import type { UserNotification, ChannelPreference } from '../api/Notifications';

export interface NotificationContextValue {
  notifications: UserNotification[];
  unreadCount: number;
  preferences: ChannelPreference[];
  loading: boolean;
  refresh: () => Promise<void>;
  markAsRead: (notificationId: number) => Promise<void>;
  dismiss: (notificationId: number) => Promise<void>;
  markAllAsRead: () => Promise<void>;
  dismissAll: () => Promise<void>;
  updatePreference: (pref: ChannelPreference) => Promise<void>;
}

export const NotificationContext = createContext<NotificationContextValue>({
  notifications: [],
  unreadCount: 0,
  preferences: [],
  loading: true,
  refresh: async () => {},
  markAsRead: async () => {},
  dismiss: async () => {},
  markAllAsRead: async () => {},
  dismissAll: async () => {},
  updatePreference: async () => {},
});

export const useNotifications = () => useContext(NotificationContext);
