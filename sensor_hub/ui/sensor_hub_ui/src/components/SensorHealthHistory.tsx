import type {Sensor} from "../types/types.ts";
import useSensorHealthHistory from "../hooks/useSensorHealthHistory.ts";
import {useEffect, useState} from "react";
import {DataGrid, type GridColDef} from "@mui/x-data-grid";
import LayoutCard from "../tools/LayoutCard.tsx";
import {TypographyH2} from "../tools/Typography.tsx";
import {Alert, Button, Snackbar, TextField} from "@mui/material";
import RefreshIcon from '@mui/icons-material/Refresh';

interface SensorHealthHistoryProps {
  sensor: Sensor,
}

function SensorHealthHistory({sensor}: SensorHealthHistoryProps) {
  const [limit, setLimit] = useState(100);
  const [limitInput, setLimitInput] = useState("100");
  const [healthHistory, refresh] = useSensorHealthHistory(sensor.name, limit);
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
        <div style={{display: "flex", justifyContent: "flex-end", flexGrow: 1, width: "100%", marginTop: 16, gap: 16}}>
          <TextField
            label="Limit History Entries"
            type="number"
            defaultValue={100}
            onChange={(e) => setLimitInput(e.target.value)}
            sx={{ mt: 2, width: 200 }}
          />
          <Button
            onClick={() => {
            setIsLoading(true);
            setLimit(parseInt(limitInput));
            refresh().then(() => {
              setIsLoading(false);
              setSnackbarOpen(true);
            });
          }}
            variant="outlined" startIcon={<RefreshIcon />}
            sx={{
              mt: 2,
              alignSelf: 'center',
              height: "56px",
            }}
          >
            Refresh
          </Button>
        </div>


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