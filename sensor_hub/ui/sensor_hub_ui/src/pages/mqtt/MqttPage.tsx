import PageContainer from '../../tools/PageContainer';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';
import { Box, Grid } from '@mui/material';
import MqttBrokersCard from '../../components/MqttBrokersCard';
import MqttSubscriptionsCard from '../../components/MqttSubscriptionsCard';
import MqttStatsCard from '../../components/MqttStatsCard';
import PendingSensorsCard from '../../components/PendingSensorsCard';

export default function MqttPage() {
  const { user } = useAuth();

  return (
    <PageContainer titleText="MQTT" loading={user === undefined}>
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%' }}>
          {hasPerm(user, 'view_mqtt') && (
            <Grid size={12}><MqttStatsCard /></Grid>
          )}
          {hasPerm(user, 'view_sensors') && (
            <Grid size={12}><PendingSensorsCard /></Grid>
          )}
          {hasPerm(user, 'view_mqtt') && (
            <>
              <Grid size={12}><MqttBrokersCard /></Grid>
              <Grid size={12}><MqttSubscriptionsCard /></Grid>
            </>
          )}
        </Grid>
      </Box>
    </PageContainer>
  );
}
