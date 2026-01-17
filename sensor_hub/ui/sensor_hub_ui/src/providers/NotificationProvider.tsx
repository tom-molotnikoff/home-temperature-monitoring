import React, { useEffect, useState, useCallback } from 'react';
import { useAuth } from './AuthContext';
import { NotificationContext, type NotificationContextValue } from './NotificationContext';
import {
  getNotifications,
  getUnreadCount,
  markAsRead as apiMarkAsRead,
  dismissNotification,
  bulkMarkAsRead,
  bulkDismiss,
  getChannelPreferences,
  setChannelPreference,
  type UserNotification,
  type ChannelPreference,
  type Notification,
} from '../api/Notifications';
import { WEBSOCKET_BASE } from '../environment/Environment';

export default function NotificationProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  const [notifications, setNotifications] = useState<UserNotification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [preferences, setPreferences] = useState<ChannelPreference[]>([]);
  const [loading, setLoading] = useState(true);

  const hasPermission = user?.permissions?.includes('view_notifications');

  const refresh = useCallback(async () => {
    if (!user || !hasPermission) return;
    try {
      const [notifs, count, prefs] = await Promise.all([
        getNotifications(50, 0, false),
        getUnreadCount(),
        getChannelPreferences(),
      ]);
      setNotifications(notifs);
      setUnreadCount(count.count);
      setPreferences(prefs);
    } catch (err) {
      console.error('Failed to load notifications:', err);
    } finally {
      setLoading(false);
    }
  }, [user, hasPermission]);

  const markAsRead = useCallback(async (notificationId: number) => {
    await apiMarkAsRead(notificationId);
    setNotifications(prev =>
      prev.map(n =>
        n.notification_id === notificationId ? { ...n, is_read: true } : n
      )
    );
    setUnreadCount(prev => Math.max(0, prev - 1));
  }, []);

  const dismiss = useCallback(async (notificationId: number) => {
    await dismissNotification(notificationId);
    setNotifications(prev => prev.filter(n => n.notification_id !== notificationId));
    setUnreadCount(prev => Math.max(0, prev - 1));
  }, []);

  const markAllAsRead = useCallback(async () => {
    await bulkMarkAsRead();
    setNotifications(prev => prev.map(n => ({ ...n, is_read: true })));
    setUnreadCount(0);
  }, []);

  const dismissAll = useCallback(async () => {
    await bulkDismiss();
    setNotifications([]);
    setUnreadCount(0);
  }, []);

  const updatePreference = useCallback(async (pref: ChannelPreference) => {
    await setChannelPreference(pref);
    setPreferences(prev => {
      const idx = prev.findIndex(p => p.category === pref.category);
      if (idx >= 0) {
        const updated = [...prev];
        updated[idx] = pref;
        return updated;
      }
      return [...prev, pref];
    });
  }, []);

  // Initial load
  useEffect(() => {
    if (user === undefined) return;
    if (!user || !hasPermission) {
      setLoading(false);
      return;
    }
    refresh();
  }, [user, hasPermission, refresh]);

  // WebSocket subscription for real-time updates
  useEffect(() => {
    if (!user || !hasPermission) return;

    const ws = new WebSocket(`${WEBSOCKET_BASE}/notifications/ws`);
    
    ws.onmessage = (event) => {
      if (!event.data || event.data === 'null') return;
      try {
        const newNotif = JSON.parse(event.data) as Notification;
        const userNotif: UserNotification = {
          id: 0,
          user_id: user.id,
          notification_id: newNotif.id,
          is_read: false,
          is_dismissed: false,
          notification: newNotif,
        };
        setNotifications(prev => [userNotif, ...prev]);
        setUnreadCount(prev => prev + 1);
      } catch (err) {
        console.error('Failed to parse notification WebSocket message:', err);
      }
    };

    ws.onerror = (err) => {
      console.error('Notifications WebSocket error:', err);
    };

    ws.onclose = (event) => {
      console.debug('Notifications WebSocket closed', event);
    };

    return () => ws.close();
  }, [user, hasPermission]);

  const value: NotificationContextValue = {
    notifications,
    unreadCount,
    preferences,
    loading,
    refresh,
    markAsRead,
    dismiss,
    markAllAsRead,
    dismissAll,
    updatePreference,
  };

  return (
    <NotificationContext.Provider value={value}>
      {children}
    </NotificationContext.Provider>
  );
}
