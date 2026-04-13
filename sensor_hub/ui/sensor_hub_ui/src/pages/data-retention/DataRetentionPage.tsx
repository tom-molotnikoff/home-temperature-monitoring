import { Box, Grid } from '@mui/material';
import PageContainer from '../../tools/PageContainer';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';
import { useIsMobile } from '../../hooks/useMobile';
import DataRetentionCard from '../../components/DataRetentionCard';

function DataRetentionPage() {
  const { user } = useAuth();
  const isMobile = useIsMobile();

  return (
    <PageContainer titleText="Data Retention" loading={user === undefined}>
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%' }}>
          {hasPerm(user, 'view_sensors') && (
            <Grid size={isMobile ? 12 : 12}><DataRetentionCard /></Grid>
          )}
        </Grid>
      </Box>
    </PageContainer>
  );
}

export default DataRetentionPage;
