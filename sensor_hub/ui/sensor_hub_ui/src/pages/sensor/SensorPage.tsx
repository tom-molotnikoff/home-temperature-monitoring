import { Box, Grid } from "@mui/material";
import PageContainer from "../../tools/PageContainer.tsx";
import { useSensorContext } from "../../hooks/useSensorContext.tsx";

interface SensorPageProps {
  sensorId: number;
}

function SensorPage({sensorId}: SensorPageProps) {
  const {sensors} = useSensorContext();
  const sensor = sensors.find(s => s.id === sensorId);

  return (
    <PageContainer titleText="Sensor">
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%" }}>
          <Grid size={12}>
            <div>
              <h2>Sensor Detail Page</h2>
              {sensor ? (
                <div>
                  <p><strong>Name:</strong> {sensor.name}</p>
                  <p><strong>Type:</strong> {sensor.type}</p>
                  <p><strong>API URL:</strong> {sensor.url}</p>
                </div>
              ) : (
                <p>Sensor not found.</p>
              )}
            </div>
          </Grid>
        </Grid>
      </Box>
    </PageContainer>
  )
}

export default SensorPage;