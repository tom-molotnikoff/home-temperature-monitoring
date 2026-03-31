import { DataGrid, type GridColDef } from "@mui/x-data-grid";
import { useCurrentTemperatures } from "../hooks/useCurrentTemperatures";
import { TypographyH2 } from "../tools/Typography";
import LayoutCard from "../tools/LayoutCard.tsx";
import { useIsMobile } from "../hooks/useMobile";
import { useSensorContext } from "../hooks/useSensorContext";
import EmptyState from "./EmptyState";
import ThermostatOutlinedIcon from "@mui/icons-material/ThermostatOutlined";

interface CurrentTemperaturesProps {
  cardHeight?: string | number;
}

function CurrentTemperatures({ cardHeight }: CurrentTemperaturesProps) {
  const isMobile = useIsMobile();
  const currentTemperatures = useCurrentTemperatures();
  const { loaded } = useSensorContext();

  const sensorNames = Object.keys(currentTemperatures).sort((a, b) =>
    a.localeCompare(b)
  );

  const rows = sensorNames.map((sensor) => {
    const reading = currentTemperatures[sensor];
    return {
      id: sensor,
      sensor_name: reading.sensor_name,
      temperature: reading.temperature,
      time: reading.time,
    };
  });

  const columns: GridColDef[] = [
    { field: "sensor_name", headerName: "Sensor Name", flex: 1, minWidth: 150 },
    {
      field: "temperature",
      headerName: "Temp (°C)",
      flex: 1,
      type: "number",
      minWidth: 90,
    },
    { field: "time", headerName: "Time", flex: 1, minWidth: 200 },
  ];

  const columnVisibilityModel = isMobile
    ? { time: false }
    : { time: true };

  return (
    <LayoutCard variant="secondary" changes={{ alignItems: "center", height: cardHeight, width: "100%" }}>
      <TypographyH2>Live Temperature</TypographyH2>
      <div
        style={{
          flex: 1,
          minHeight: 0,
          display: "flex",
          flexDirection: "column",
          paddingBottom: 10,
          width: "100%"
        }}
      >
        {loaded && rows.length === 0 ? (
          <EmptyState
            icon={<ThermostatOutlinedIcon sx={{ fontSize: 48 }} />}
            title="No live temperature data"
            description="Add and enable sensors to see live readings here."
            actionLabel="Go to Sensors"
            actionHref="/sensors-overview"
          />
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
            columnVisibilityModel={columnVisibilityModel}
            sx={{
              backgroundColor: 'background.paper',
              borderRadius: 2,
              mt: 2,
              '& .MuiDataGrid-cell': { fontSize: isMobile ? '0.9rem' : '1rem' },
              '& .MuiDataGrid-columnHeaders': { fontWeight: 'bold' },
            }}
            loading={!loaded}
          />
        )}
      </div>
    </LayoutCard>
  );
}

export default CurrentTemperatures;
