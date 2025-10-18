import type {Sensor, SensorHealthStatus} from "../types/types.ts";
import LayoutCard from "../tools/LayoutCard.tsx";
import {TypographyH2} from "../tools/Typography.tsx";
import {DataGrid, type GridColDef, type GridRowParams} from '@mui/x-data-grid';
import { useIsMobile } from "../hooks/useMobile";
import {useEffect, useState} from 'react';
import {Menu, MenuItem, type SnackbarCloseReason, Snackbar, Alert, Box, CircularProgress} from '@mui/material';
import {API_BASE} from "../environment/Environment.ts";
import {useNavigate} from "react-router";

interface SensorSummaryCardProps {
  sensors: Sensor[],
  cardHeight?: string | number;
  showReason: boolean;
  showType: boolean;
  showEnabled: boolean;
  title?: string;
}

type row = {
  id: string | number;
  name: string;
  type: string;
  url: string;
  healthStatus: SensorHealthStatus;
  healthReason: string | null;
  enabled: boolean;
} | null;

function SensorSummaryCard({ sensors, cardHeight, showReason, showType, title, showEnabled }: SensorSummaryCardProps) {
  const isMobile = useIsMobile();
  const [menuAnchorEl, setMenuAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedRow, setSelectedRow] = useState<row>(null);
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [alertSeverity, setAlertSeverity] = useState<'success' | 'error'>('success');
  const [alertMessage, setAlertMessage] = useState('');

  const [isLoading, setIsLoading] = useState(true);

  const navigate = useNavigate();

  useEffect(() => {
    if (sensors.length > 0) {
      setIsLoading(false);
    }
  }, [sensors]);

  const handleClose = (
    _event: React.SyntheticEvent | Event,
    reason?: SnackbarCloseReason,
  ) => {
    if (reason === 'clickaway') {
      return;
    }
    setSnackbarOpen(false);
  };
  const handleRowClick = (params: GridRowParams, event: React.MouseEvent) => {
    setSelectedRow(params.row);
    setMenuAnchorEl(event.currentTarget as HTMLElement);
  };

  const triggerReading = async (sensor: string) => {
    const response = await fetch(`${API_BASE}/sensors/collect/${sensor}`, {
      method: "POST",
    });
    if (!response.ok) {
      throw new Error(`Failed to trigger reading for ${sensor}`);
    }
    setSnackbarOpen(true);
  };


  const handleTriggerReading = async () => {
    handleMenuClose();
    try {
      if (selectedRow) {
        await triggerReading(selectedRow.name);
        setAlertSeverity('success');
        setAlertMessage('Reading triggered successfully');
      }
    } catch (err: unknown) {
      setAlertSeverity('error');
      if (err instanceof Error) {
        setAlertMessage(err.message);
      } else {
        setAlertMessage('Failed to trigger reading');
      }
    } finally {
      setSnackbarOpen(true);
    }
  }

  const handleViewDetails = () => {
    handleMenuClose();
    navigate(`/sensor/${selectedRow?.id}`);
  }

  const handleMenuClose = () => {
    setMenuAnchorEl(null);
    setSelectedRow(null);
  };

  const columns: GridColDef[] = [
    { field: 'name', headerName: 'Sensor Name', flex: 1, minWidth: 100 },
    { field: 'type', headerName: 'Type', flex: 1, minWidth: 100 },
    { field: 'url', headerName: 'API URL', flex: 2, minWidth: 200 },
    { field: 'healthStatus', headerName: 'Health Status', flex: 1, minWidth: 100 },
    { field: 'healthReason', headerName: 'Health Reason', flex: 2, minWidth: 200 },
    { field: 'enabled', headerName: 'Enabled', flex: 1, minWidth: 80, type: 'boolean' },
  ];

  const rows: row[] = sensors.map((sensor) => ({
    id: sensor.id,
    name: sensor.name,
    type: sensor.type,
    url: sensor.url,
    healthStatus: sensor.healthStatus,
    healthReason: sensor.healthReason,
    enabled: sensor.enabled,
  }));

  const columnVisibilityModel = {
    id: true,
    name: true,
    type: showType,
    url: !isMobile,
    healthStatus: true,
    healthReason: showReason,
    enabled: showEnabled,
  }

  return (
    <LayoutCard variant="secondary" changes={{alignItems: "center", height: cardHeight, width: "100%"}}>
      <TypographyH2>{title ? title : "Sensor Summary"}</TypographyH2>
      <div
        style={{
          height: cardHeight,
          paddingBottom: 10,
          display: "flex",
          flexDirection: "column",
          width: "100%"
        }}
      >
        {isLoading ? (
          <Box
            display="flex"
            justifyContent="center"
            alignItems="center"
            minHeight={200}
          >
            <CircularProgress />
          </Box>
        ) : (
          <DataGrid
            showToolbar
            rows={rows}
            columns={columns}
            pageSizeOptions={[5, 10, 25, 50, 100]}
            initialState={{
              pagination: {
                paginationModel: { pageSize: 5, page: 0 },
              },
            }}
            getRowHeight={showReason ? () => 'auto' : undefined}
            onRowClick={handleRowClick}
            columnVisibilityModel={columnVisibilityModel}
            sx={{
              backgroundColor: 'background.paper',
              borderRadius: 2,
              mt: 2,
              '& .MuiDataGrid-cell': { fontSize: isMobile ? '0.9rem' : '1rem' },
              '& .MuiDataGrid-columnHeaders': { fontWeight: 'bold' },
              '.MuiDataGrid-row:hover': { cursor: 'pointer' },
            }}
          />
        )}
        <Menu
          anchorEl={menuAnchorEl}
          open={Boolean(menuAnchorEl)}
          onClose={handleMenuClose}
        >
          <MenuItem onClick={handleTriggerReading}>Trigger Reading</MenuItem>
          <MenuItem onClick={handleViewDetails}>View Details</MenuItem>
        </Menu>
      </div>
      <Snackbar
        open={snackbarOpen}
        autoHideDuration={2000}
        onClose={handleClose}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert onClose={handleClose} severity={alertSeverity} sx={{ width: '100%' }}>
          {alertMessage}
        </Alert>
      </Snackbar>
    </LayoutCard>
  );
}


export default SensorSummaryCard