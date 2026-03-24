import CurrentTemperatures from '../../components/CurrentTemperatures';
import IndoorTemperatureDataCard from '../../components/IndoorTemperatureDataCard';
import WeatherDataCard from '../../components/WeatherDataCard';
import TemperatureSensorsCard from '../../components/TemperatureSensorsCard';
import PageContainer from '../../tools/PageContainer';
import { useIsMobile } from '../../hooks/useMobile';
import { Box, Grid } from '@mui/material';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';

function TemperatureDashboard() {
  const { user } = useAuth();
  const isMobile = useIsMobile();

  return (
    <PageContainer titleText="Temperature Dashboard" loading={user === undefined}>
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%' }}>
          {hasPerm(user, 'view_readings') && (
            <>
              <Grid size={12}><IndoorTemperatureDataCard /></Grid>
              <Grid size={12}><WeatherDataCard /></Grid>
            </>
          )}
          {hasPerm(user, 'view_sensors') && (
            <Grid size={isMobile ? 12 : 6}><TemperatureSensorsCard /></Grid>
          )}
          {hasPerm(user, 'view_readings') && (
            <Grid size={isMobile ? 12 : 6}><CurrentTemperatures cardHeight="100%" /></Grid>
          )}
        </Grid>
      </Box>
    </PageContainer>
  );
}

export default TemperatureDashboard;
