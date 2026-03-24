import PageContainer from '../../tools/PageContainer';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';
import { Box, Grid } from '@mui/material';
import AlertRulesCard from '../../components/AlertRulesCard';
import NotificationsCard from '../../components/NotificationsCard';
import NotificationPreferencesCard from '../../components/NotificationPreferencesCard';
import OAuthConfigCard from '../../components/OAuthConfigCard';

export default function NotificationsPage() {
  const { user } = useAuth();

  return (
    <PageContainer titleText="Alerts & Notifications" loading={user === undefined}>
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%' }}>
          {hasPerm(user, 'view_alerts') && (
            <Grid size={12}><AlertRulesCard /></Grid>
          )}
          {hasPerm(user, 'manage_notifications') && (
            <Grid size={12}><NotificationPreferencesCard /></Grid>
          )}
          {hasPerm(user, 'manage_oauth') && (
            <Grid size={12}><OAuthConfigCard /></Grid>
          )}
          {hasPerm(user, 'view_notifications') && (
            <Grid size={12}><NotificationsCard /></Grid>
          )}
        </Grid>
      </Box>
    </PageContainer>
  );
}
