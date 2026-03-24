import PageContainer from '../../tools/PageContainer';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';
import { Box, Grid } from '@mui/material';
import UserManagementCard from '../../components/UserManagementCard';
import RolePermissionsCard from '../../components/RolePermissionsCard';

export default function UsersPage() {
  const { user } = useAuth();

  return (
    <PageContainer titleText="User Management" loading={user === undefined}>
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%' }}>
          {hasPerm(user, 'view_users') && (
            <Grid size={12}><UserManagementCard /></Grid>
          )}
          {hasPerm(user, 'view_roles') && (
            <Grid size={12}><RolePermissionsCard /></Grid>
          )}
        </Grid>
      </Box>
    </PageContainer>
  );
}
