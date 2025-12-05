import { DataGrid, type GridColDef } from "@mui/x-data-grid";
import { useCurrentTemperatures } from "../hooks/useCurrentTemperatures";
import { TypographyH2 } from "../tools/Typography";
import { useEffect, useState } from "react";
import { CircularProgress, Box } from "@mui/material";
import LayoutCard from "../tools/LayoutCard.tsx";
import { useIsMobile } from "../hooks/useMobile";

interface CurrentTemperaturesProps {
  cardHeight?: string | number;
}

function CurrentTemperatures({ cardHeight }: CurrentTemperaturesProps) {
  const isMobile = useIsMobile();
  const currentTemperatures = useCurrentTemperatures();
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (Object.keys(currentTemperatures).length > 0) {
      setIsLoading(false);
    }
  }, [currentTemperatures]);

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
          height: cardHeight,
          display: "flex",
          flexDirection: "column",
          paddingBottom: 10,
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
            columnVisibilityModel={columnVisibilityModel}
            sx={{
              backgroundColor: 'background.paper',
              borderRadius: 2,
              mt: 2,
              '& .MuiDataGrid-cell': { fontSize: isMobile ? '0.9rem' : '1rem' },
              '& .MuiDataGrid-columnHeaders': { fontWeight: 'bold' },
            }}
          />
        )}
      </div>
    </LayoutCard>
  );
}

export default CurrentTemperatures;
