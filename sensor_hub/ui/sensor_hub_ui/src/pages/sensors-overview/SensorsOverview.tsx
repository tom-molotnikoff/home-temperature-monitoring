import PageContainer from '../../tools/PageContainer';
import { useIsMobile } from '../../hooks/useMobile';
import { Grid, Box } from '@mui/material';
import SensorHealthCard from '../../components/SensorHealthCard';
import AddNewSensor from '../../components/AddNewSensor';
import SensorTypeCard from '../../components/SensorTypeCard';
import TotalReadingsForEachSensorCard from '../../components/TotalReadingsForEachSensorCard';
import AllSensorsCard from '../../components/AllSensorsCard';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';

function SensorsOverview() {
  const isMobile = useIsMobile();
  const { user } = useAuth();

  return (
    <PageContainer titleText="Sensors Overview" loading={user === undefined}>
      <Box sx={{ width: '100%', flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%', width: '100%', flexGrow: 1 }}>
          {hasPerm(user, 'manage_sensors') && (
            <Grid size={isMobile ? 12 : 4}><AddNewSensor /></Grid>
          )}
          {hasPerm(user, 'view_sensors') && (
            <>
              <Grid size={isMobile ? 12 : 4}><SensorHealthCard /></Grid>
              <Grid size={isMobile ? 12 : 4}><SensorTypeCard /></Grid>
              <Grid size={isMobile ? 12 : 8}><AllSensorsCard /></Grid>
              <Grid size={isMobile ? 12 : 4}><TotalReadingsForEachSensorCard /></Grid>
            </>
          )}
        </Grid>
      </Box>
    </PageContainer>
  );
}

export default SensorsOverview;
