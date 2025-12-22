import type {Sensor} from "../types/types.ts";
import useSensorHealthHistory from "../hooks/useSensorHealthHistory.ts";
import {useEffect, useState} from "react";
import {DataGrid, type GridColDef} from "@mui/x-data-grid";
import LayoutCard from "../tools/LayoutCard.tsx";
import {TypographyH2} from "../tools/Typography.tsx";
import {Alert, Button, Snackbar} from "@mui/material";
import RefreshIcon from '@mui/icons-material/Refresh';

interface SensorHealthHistoryProps {
  sensor: Sensor,
}

function SensorHealthHistory({sensor}: SensorHealthHistoryProps) {
  const [healthHistory, refresh] = useSensorHealthHistory(sensor.name)
  const [isLoading, setIsLoading] = useState(true);
  const [snackbarOpen, setSnackbarOpen] = useState(false);

  useEffect(() => {
    if (healthHistory.length > 0) {
      setIsLoading(false);
    }
  }, [healthHistory]);

  const rows = healthHistory.map((entry) => ({
    id: entry.id,
    healthStatus: entry.healthStatus,
    recordedAt: entry.recordedAt.toLocaleString(),
  }));

  const columns: GridColDef[] = [
    { field: "healthStatus", headerName: "Health Status", flex: 1, minWidth: 150 },
    { field: "recordedAt", headerName: "Recorded At", flex: 1, minWidth: 200 },
  ]


  return (
    <LayoutCard variant="secondary" changes={{alignItems: "center", width: "100%", minHeight: 400}}>
      <TypographyH2>Sensor Health History</TypographyH2>
      <div
        style={{
          minHeight: 450,
          display: "flex",
          flexDirection: "column",
          width: "100%",
        }}
      >
        <DataGrid
          showToolbar
          rows={rows}
          columns={columns}
          initialState={{
            pagination: {
              paginationModel: { pageSize: 5, page: 0 },
            },
          }}
          loading={isLoading}
          sx={{
            backgroundColor: 'background.paper',
            borderRadius: 2,
            mt: 2,
            '& .MuiDataGrid-columnHeaders': { fontWeight: 'bold' },
          }}
        />
        <Button onClick={() => {
          setIsLoading(true);
          refresh().then(() => {
            setIsLoading(false);
            setSnackbarOpen(true);
          });
        }} variant="outlined" startIcon={<RefreshIcon />} sx={{ mt: 2, alignSelf: 'flex-end' }}>
          Refresh
        </Button>

      </div>
      <Snackbar
        open={snackbarOpen}
        onClose={() => setSnackbarOpen(false)}
        autoHideDuration={2000}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert sx={{ width: '100%' }}>
          Sensor health history refreshed.
        </Alert>
      </Snackbar>
    </LayoutCard>
  );
}

export default SensorHealthHistory;