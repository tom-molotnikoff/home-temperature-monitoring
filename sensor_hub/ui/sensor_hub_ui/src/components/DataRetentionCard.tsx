import { useState } from 'react';
import type { Sensor } from '../types/types';
import type { GridRowParams } from '@mui/x-data-grid';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import { DataGrid, type GridColDef } from '@mui/x-data-grid';
import { useIsMobile } from '../hooks/useMobile';
import { useSensorContext } from '../hooks/useSensorContext';
import { useProperties } from '../hooks/useProperties';
import { useAuth } from '../providers/AuthContext';
import { hasPerm } from '../tools/Utils';
import { Typography, Chip, Box, Menu, MenuItem } from '@mui/material';
import { formatRetention } from '../tools/retention';
import EditRetentionDialog from './EditRetentionDialog';

function DataRetentionCard() {
  const { sensors } = useSensorContext();
  const properties = useProperties();
  const { user } = useAuth();
  const isMobile = useIsMobile();

  const globalRetentionDays = parseInt(properties['sensor.data.retention.days'] || '90', 10);
  const globalRetentionHours = globalRetentionDays * 24;

  const activeSensors = sensors.filter((s) => s.status === 'active');

  const [menuAnchorEl, setMenuAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedSensor, setSelectedSensor] = useState<Sensor | null>(null);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [retentionOverrides, setRetentionOverrides] = useState<Record<number, number | null>>({});

  const displaySensors = activeSensors.map((s) => {
    if (s.id in retentionOverrides) {
      return { ...s, retentionHours: retentionOverrides[s.id] };
    }
    return s;
  });

  const handleRowClick = (params: GridRowParams, event: React.MouseEvent) => {
    const found = displaySensors.find((s) => s.id === params.row.id);
    setSelectedSensor(found ?? (params.row as Sensor));
    setMenuAnchorEl(event.currentTarget as HTMLElement);
  };

  const closeMenu = () => setMenuAnchorEl(null);
  const canManage = user && hasPerm(user, 'manage_sensors');

  const columns: GridColDef[] = [
    { field: 'name', headerName: 'Sensor', flex: 1, minWidth: 140 },
    { field: 'sensorDriver', headerName: 'Driver', flex: 1, minWidth: 120 },
    {
      field: 'retentionHours',
      headerName: 'Retention',
      flex: 1,
      minWidth: 140,
      renderCell: (params) => {
        const sensor = params.row as Sensor;
        if (sensor.retentionHours !== null) {
          return <Chip label={formatRetention(sensor.retentionHours)} color="primary" size="small" variant="outlined" />;
        }
        return <Typography variant="body2" color="text.secondary" sx={{ display: 'flex', alignItems: 'center', height: '100%' }}>Global default</Typography>;
      },
    },
    {
      field: 'effective',
      headerName: 'Effective',
      flex: 1,
      minWidth: 120,
      valueGetter: (_value: unknown, row: Sensor) => {
        const hours = row.retentionHours ?? globalRetentionHours;
        return formatRetention(hours);
      },
    },
  ];

  return (
    <>
      <LayoutCard variant="secondary" changes={{ alignItems: 'stretch', height: '100%', width: '100%' }}>
        <Box sx={{ width: '100%' }}>
          <TypographyH2>Sensor Retention Overview</TypographyH2>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
            Global default: {formatRetention(globalRetentionHours)}. Click a sensor to edit its retention policy.
          </Typography>
          <DataGrid
            rows={displaySensors}
            columns={columns}
            initialState={{
              pagination: { paginationModel: { pageSize: 10 } },
              columns: {
                columnVisibilityModel: isMobile ? { sensorDriver: false } : {},
              },
            }}
            pageSizeOptions={[5, 10, 25]}
            disableRowSelectionOnClick
            onRowClick={handleRowClick}
            autoHeight
            sx={{
              border: 0,
              '& .MuiDataGrid-cell': { fontSize: isMobile ? '0.9rem' : '1rem' },
              '& .MuiDataGrid-columnHeaders': { fontWeight: 'bold' },
              '.MuiDataGrid-row:hover': { cursor: 'pointer' },
            }}
          />
        </Box>

        <Menu anchorEl={menuAnchorEl} open={Boolean(menuAnchorEl)} onClose={closeMenu}>
          <MenuItem
            disabled={!canManage}
            onClick={() => { closeMenu(); setOpenEditDialog(true); }}
          >
            Edit Retention
          </MenuItem>
        </Menu>
      </LayoutCard>

      <EditRetentionDialog
        open={openEditDialog}
        onClose={() => setOpenEditDialog(false)}
        onSaved={(sensorId, retentionHours) => {
          setRetentionOverrides((prev) => ({ ...prev, [sensorId]: retentionHours }));
        }}
        sensor={selectedSensor}
        globalRetentionHours={globalRetentionHours}
      />
    </>
  );
}

export default DataRetentionCard;
