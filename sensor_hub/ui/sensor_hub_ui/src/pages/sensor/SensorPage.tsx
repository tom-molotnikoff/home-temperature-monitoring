import { Box, Grid } from '@mui/material';
import PageContainer from '../../tools/PageContainer';
import { useSensorContext } from '../../hooks/useSensorContext';
import { useIsMobile } from '../../hooks/useMobile';
import SensorInfoCard from '../../components/SensorInfoCard';
import EditSensorDetails from '../../components/EditSensorDetails';
import SensorHealthHistory from '../../components/SensorHealthHistory';
import SensorHealthHistoryChartCard from '../../components/SensorHealthHistoryChartCard';
import SensorTemperatureDataCard from '../../components/SensorTemperatureDataCard';
import SensorRetentionCard from '../../components/SensorRetentionCard';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';

interface SensorPageProps {
  sensorId: number;
}

function SensorPage({ sensorId }: SensorPageProps) {
  const { sensors } = useSensorContext();
  const sensor = sensors.find(s => s.id === sensorId);
  const { user } = useAuth();
  const isMobile = useIsMobile();

  if (user === undefined) {
    return (
      <PageContainer titleText="Sensor" loading>
        <></>
      </PageContainer>
    );
  }

  if (!sensor) {
    return (
      <PageContainer titleText="Sensor Not Found">
        <Box sx={{ flexGrow: 1 }}>
          <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%' }}>
            <Grid size={12}>
              <h2>Sensor with ID {sensorId} not found.</h2>
            </Grid>
          </Grid>
        </Box>
      </PageContainer>
    );
  }

  return (
    <PageContainer titleText="Sensor">
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%' }}>
          {hasPerm(user, 'view_sensors') && (
            <>
              <Grid size={isMobile ? 12 : 6}><SensorInfoCard sensor={sensor} user={user} /></Grid>
              <Grid size={isMobile ? 12 : 6}><EditSensorDetails sensor={sensor} /></Grid>
            </>
          )}
          {hasPerm(user, 'view_readings') && (
            <>
              <Grid size={isMobile ? 12 : 6}><SensorHealthHistoryChartCard sensor={sensor} /></Grid>
              <Grid size={isMobile ? 12 : 6}><SensorTemperatureDataCard sensor={sensor} /></Grid>
            </>
          )}
          {hasPerm(user, 'view_sensors') && (
            <Grid size={isMobile ? 12 : 6}><SensorHealthHistory sensor={sensor} /></Grid>
          )}
          {hasPerm(user, 'manage_sensors') && (
            <Grid size={isMobile ? 12 : 6}><SensorRetentionCard sensor={sensor} /></Grid>
          )}
        </Grid>
      </Box>
    </PageContainer>
  );
}

export default SensorPage;
