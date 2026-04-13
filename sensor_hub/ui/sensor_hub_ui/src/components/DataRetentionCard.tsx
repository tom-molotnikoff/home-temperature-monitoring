import type { Sensor } from '../types/types';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import { DataGrid, type GridColDef } from '@mui/x-data-grid';
import { useIsMobile } from '../hooks/useMobile';
import { useSensorContext } from '../hooks/useSensorContext';
import { useProperties } from '../hooks/useProperties';
import { Typography, Chip, Box } from '@mui/material';
import { useNavigate } from 'react-router';

function formatRetention(hours: number): string {
  if (hours >= 168 && hours % 168 === 0) return `${hours / 168} week${hours / 168 !== 1 ? 's' : ''}`;
  if (hours >= 24 && hours % 24 === 0) return `${hours / 24} day${hours / 24 !== 1 ? 's' : ''}`;
  return `${hours} hour${hours !== 1 ? 's' : ''}`;
}

function DataRetentionCard() {
  const { sensors } = useSensorContext();
  const properties = useProperties();
  const isMobile = useIsMobile();
  const navigate = useNavigate();

  const globalRetentionDays = parseInt(properties['sensor.data.retention.days'] || '90', 10);
  const globalRetentionHours = globalRetentionDays * 24;

  const activeSensors = sensors.filter((s) => s.status === 'active');

  const columns: GridColDef[] = [
    { field: 'name', headerName: 'Sensor', flex: 1, minWidth: 140 },
    { field: 'sensorDriver', headerName: 'Driver', flex: 1, minWidth: 120, hideable: true },
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
        return <Typography variant="body2" color="text.secondary">Global default</Typography>;
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

  if (isMobile) {
    const hiddenFields = ['sensorDriver'];
    columns.forEach((col) => {
      if (hiddenFields.includes(col.field)) col.hideable = false;
    });
  }

  return (
    <LayoutCard variant="secondary" changes={{ alignItems: 'center', height: '100%', width: '100%' }}>
      <Box sx={{ width: '100%' }}>
        <TypographyH2>Sensor Retention Overview</TypographyH2>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Global default: {formatRetention(globalRetentionHours)}. Sensors with custom retention override this value.
        </Typography>
        <DataGrid
          rows={activeSensors}
          columns={columns}
          initialState={{
            pagination: { paginationModel: { pageSize: 10 } },
            columns: {
              columnVisibilityModel: isMobile ? { sensorDriver: false } : {},
            },
          }}
          pageSizeOptions={[5, 10, 25]}
          disableRowSelectionOnClick
          onRowClick={(params) => navigate(`/sensor/${params.row.id}`)}
          autoHeight
          sx={{
            border: 0,
            '& .MuiDataGrid-cell': { fontSize: isMobile ? '0.9rem' : '1rem' },
            '& .MuiDataGrid-columnHeaders': { fontWeight: 'bold' },
            '.MuiDataGrid-row:hover': { cursor: 'pointer' },
          }}
        />
      </Box>
    </LayoutCard>
  );
}

export default DataRetentionCard;
