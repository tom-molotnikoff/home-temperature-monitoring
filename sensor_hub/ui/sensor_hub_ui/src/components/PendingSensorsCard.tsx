import { useEffect, useState, useCallback } from 'react';
import { DataGrid } from '@mui/x-data-grid';
import type { GridColDef } from '@mui/x-data-grid';
import { Button, Box, Chip, Stack, Typography } from '@mui/material';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import CancelIcon from '@mui/icons-material/Cancel';
import { SensorsApi } from '../api/Sensors';
import type { Sensor } from '../types/types';
import LayoutCard from '../tools/LayoutCard';
import { useAuth } from '../providers/AuthContext';
import { hasPerm } from '../tools/Utils';
import { useIsMobile } from '../hooks/useMobile';
import { logger } from '../tools/logger';
import { TypographyH2 } from '../tools/Typography';

export default function PendingSensorsCard() {
  const [pending, setPending] = useState<Sensor[]>([]);
  const [dismissed, setDismissed] = useState<Sensor[]>([]);
  const [showDismissed, setShowDismissed] = useState(false);
  const { user } = useAuth();
  const isMobile = useIsMobile();

  const load = useCallback(async () => {
    try {
      const [p, d] = await Promise.all([
        SensorsApi.getByStatus('pending'),
        SensorsApi.getByStatus('dismissed'),
      ]);
      setPending(p || []);
      setDismissed(d || []);
    } catch (e) { logger.error(e); }
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleApprove = async (id: number) => {
    try {
      await SensorsApi.approve(id);
      await load();
    } catch (e) { logger.error('Failed to approve sensor', e); }
  };

  const handleDismiss = async (id: number) => {
    try {
      await SensorsApi.dismiss(id);
      await load();
    } catch (e) { logger.error('Failed to dismiss sensor', e); }
  };

  const canManage = user && hasPerm(user, 'manage_sensors');

  const pendingColumns: GridColDef[] = [
    { field: 'id', headerName: 'ID', width: 60 },
    { field: 'name', headerName: 'Device Name', flex: 1 },
    { field: 'sensorDriver', headerName: 'Driver', width: 160 },
    {
      field: 'actions', headerName: 'Actions', width: 200, sortable: false,
      renderCell: (params) => (
        <Stack direction="row" spacing={1}>
          <Button size="small" variant="contained" color="success" startIcon={<CheckCircleIcon />}
            onClick={(e) => { e.stopPropagation(); handleApprove(params.row.id); }}
            disabled={!canManage}>
            Approve
          </Button>
          <Button size="small" variant="outlined" color="warning" startIcon={<CancelIcon />}
            onClick={(e) => { e.stopPropagation(); handleDismiss(params.row.id); }}
            disabled={!canManage}>
            Dismiss
          </Button>
        </Stack>
      ),
    },
  ];

  const dismissedColumns: GridColDef[] = [
    { field: 'id', headerName: 'ID', width: 60 },
    { field: 'name', headerName: 'Device Name', flex: 1 },
    { field: 'sensorDriver', headerName: 'Driver', width: 160 },
    {
      field: 'actions', headerName: 'Actions', width: 140, sortable: false,
      renderCell: (params) => (
        <Button size="small" variant="outlined" startIcon={<CheckCircleIcon />}
          onClick={(e) => { e.stopPropagation(); handleApprove(params.row.id); }}
          disabled={!canManage}>
          Restore
        </Button>
      ),
    },
  ];

  const mobileHiddenFields = ['id', 'sensorDriver'];
  const filteredPendingCols = isMobile ? pendingColumns.filter(c => !mobileHiddenFields.includes(c.field)) : pendingColumns;
  const filteredDismissedCols = isMobile ? dismissedColumns.filter(c => !mobileHiddenFields.includes(c.field)) : dismissedColumns;

  return (
    <LayoutCard variant="secondary" changes={{ alignItems: 'stretch', height: '100%', width: '100%' }}>
      <Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2} sx={{ width: '100%' }}>
        <TypographyH2>
          Pending Sensors
          {pending.length > 0 && (
            <Chip label={pending.length} color="warning" size="small" sx={{ ml: 1 }} />
          )}
        </TypographyH2>
      </Box>

      {pending.length === 0 ? (
        <Typography variant="body2" color="text.secondary" sx={{ py: 2 }}>
          No pending sensors. When MQTT devices are auto-discovered, they will appear here for approval.
        </Typography>
      ) : (
        <div style={{ height: 300, width: '100%' }}>
          <DataGrid rows={pending} columns={filteredPendingCols} pageSizeOptions={[5, 10]}
            initialState={{ pagination: { paginationModel: { pageSize: 5 } } }}
            disableRowSelectionOnClick />
        </div>
      )}

      {dismissed.length > 0 && (
        <Box sx={{ mt: 2 }}>
          <Button size="small" variant="text" onClick={() => setShowDismissed(!showDismissed)}>
            {showDismissed ? 'Hide' : 'Show'} dismissed sensors ({dismissed.length})
          </Button>
          {showDismissed && (
            <div style={{ height: 250, width: '100%', marginTop: 8 }}>
              <DataGrid rows={dismissed} columns={filteredDismissedCols} pageSizeOptions={[5, 10]}
                initialState={{ pagination: { paginationModel: { pageSize: 5 } } }}
                disableRowSelectionOnClick />
            </div>
          )}
        </Box>
      )}
    </LayoutCard>
  );
}
