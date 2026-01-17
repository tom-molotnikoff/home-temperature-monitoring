import { useState } from 'react';
import {
  Box,
  Typography,
  Button,
  Chip,
  IconButton,
  Menu,
  MenuItem,
  CircularProgress,
  Card,
  CardContent,
  Divider,
  Tabs,
  Tab,
} from '@mui/material';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import ErrorIcon from '@mui/icons-material/Error';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import PageContainer from '../../tools/PageContainer';
import LayoutCard from '../../tools/LayoutCard';
import { useNotifications } from '../../providers/NotificationContext';
import type { NotificationSeverity, NotificationCategory } from '../../api/Notifications';
import { useIsMobile } from '../../hooks/useMobile';

function getSeverityIcon(severity: NotificationSeverity) {
  switch (severity) {
    case 'info':
      return <InfoIcon color="info" />;
    case 'warning':
      return <WarningIcon color="warning" />;
    case 'error':
      return <ErrorIcon color="error" />;
    default:
      return <InfoIcon />;
  }
}

function getSeverityColor(severity: NotificationSeverity): 'info' | 'warning' | 'error' | 'default' {
  switch (severity) {
    case 'info':
      return 'info';
    case 'warning':
      return 'warning';
    case 'error':
      return 'error';
    default:
      return 'default';
  }
}

function getCategoryLabel(category: NotificationCategory): string {
  switch (category) {
    case 'threshold_alert':
      return 'Alert';
    case 'user_management':
      return 'User';
    case 'config_change':
      return 'Config';
    default:
      return category;
  }
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleString();
}

export default function NotificationsPage() {
  const {
    notifications,
    loading,
    markAsRead,
    dismiss,
    markAllAsRead,
    dismissAll,
    refresh,
  } = useNotifications();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedNotifId, setSelectedNotifId] = useState<number | null>(null);
  const [tabValue, setTabValue] = useState(0);
  const isMobile = useIsMobile();

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>, notifId: number) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedNotifId(notifId);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
    setSelectedNotifId(null);
  };

  const handleMarkAsRead = async () => {
    if (selectedNotifId) {
      await markAsRead(selectedNotifId);
    }
    handleMenuClose();
  };

  const handleDismiss = async () => {
    if (selectedNotifId) {
      await dismiss(selectedNotifId);
    }
    handleMenuClose();
  };

  const filteredNotifications = tabValue === 0
    ? notifications.filter(n => !n.is_read)
    : notifications;

  return (
    <PageContainer titleText="Notifications">
      <Box sx={{ flexGrow: 1 }}>
        <LayoutCard variant="secondary" changes={{ alignItems: 'stretch', height: '100%', width: '100%' }}>
          <Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2} flexWrap="wrap">
            <Typography variant="h4">Notifications</Typography>
            <Box display="flex" gap={1} flexWrap="wrap">
              <Button 
                variant="outlined" 
                onClick={() => refresh()}
                size={isMobile ? 'small' : 'medium'}
              >
                Refresh
              </Button>
              <Button 
                variant="outlined" 
                onClick={markAllAsRead}
                size={isMobile ? 'small' : 'medium'}
              >
                Mark All Read
              </Button>
              <Button 
                variant="outlined" 
                color="warning" 
                onClick={dismissAll}
                size={isMobile ? 'small' : 'medium'}
              >
                Dismiss All
              </Button>
            </Box>
          </Box>

          <Tabs value={tabValue} onChange={(_, v) => setTabValue(v)} sx={{ mb: 2 }}>
            <Tab label={`Unread (${notifications.filter(n => !n.is_read).length})`} />
            <Tab label={`All (${notifications.length})`} />
          </Tabs>

          {loading ? (
            <Box display="flex" justifyContent="center" p={4}>
              <CircularProgress />
            </Box>
          ) : filteredNotifications.length === 0 ? (
            <Box textAlign="center" py={6}>
              <CheckCircleIcon color="disabled" sx={{ fontSize: 64, mb: 2 }} />
              <Typography color="text.secondary">
                {tabValue === 0 ? 'No unread notifications' : 'No notifications'}
              </Typography>
            </Box>
          ) : (
            <Box>
              {filteredNotifications.map((notif, index) => (
                <div key={notif.notification_id}>
                  <Card
                    sx={{
                      mb: 1,
                      backgroundColor: notif.is_read ? 'transparent' : 'action.hover',
                    }}
                    variant="outlined"
                  >
                    <CardContent sx={{ display: 'flex', alignItems: 'flex-start', gap: 2, py: 2 }}>
                      <Box sx={{ mt: 0.5 }}>
                        {getSeverityIcon(notif.notification.severity)}
                      </Box>
                      <Box sx={{ flexGrow: 1 }}>
                        <Box display="flex" alignItems="center" gap={1} mb={0.5}>
                          <Typography variant="subtitle1" fontWeight={notif.is_read ? 'normal' : 'bold'}>
                            {notif.notification.title}
                          </Typography>
                          <Chip
                            label={getCategoryLabel(notif.notification.category)}
                            size="small"
                            color={getSeverityColor(notif.notification.severity)}
                            variant="outlined"
                          />
                          {!notif.is_read && (
                            <Chip label="New" size="small" color="primary" />
                          )}
                        </Box>
                        <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                          {notif.notification.message}
                        </Typography>
                        <Typography variant="caption" color="text.disabled">
                          {formatDate(notif.notification.created_at)}
                        </Typography>
                      </Box>
                      <IconButton
                        size="small"
                        onClick={(e) => handleMenuOpen(e, notif.notification_id)}
                      >
                        <MoreVertIcon />
                      </IconButton>
                    </CardContent>
                  </Card>
                  {index < filteredNotifications.length - 1 && <Divider sx={{ my: 1 }} />}
                </div>
              ))}
            </Box>
          )}

          <Menu
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={handleMenuClose}
          >
            {[
              <MenuItem key="mark-read" onClick={handleMarkAsRead}>Mark as Read</MenuItem>,
              <MenuItem key="dismiss" onClick={handleDismiss}>Dismiss</MenuItem>
            ]}
          </Menu>
        </LayoutCard>
      </Box>
    </PageContainer>
  );
}
