import { Box, Grid } from "@mui/material";
import PageContainer from "../../tools/PageContainer.tsx";
import { useSensorContext } from "../../hooks/useSensorContext.ts";
import SensorInfoCard from "../../components/SensorInfoCard.tsx";
import EditSensorDetails from "../../components/EditSensorDetails.tsx";
import SensorHealthHistory from "../../components/SensorHealthHistory.tsx";

interface SensorPageProps {
  sensorId: number;
}

function SensorPage({sensorId}: SensorPageProps) {
  const {sensors} = useSensorContext();
  const sensor = sensors.find(s => s.id === sensorId);

  if (!sensor) {
    return (
      <PageContainer titleText="Sensor Not Found">
        <Box sx={{ flexGrow: 1 }}>
          <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%" }}>
            <Grid size={12}>
              <h2>Sensor with ID {sensorId} not found.</h2>
            </Grid>
          </Grid>
        </Box>
      </PageContainer>
    )
  }

  return (
    <PageContainer titleText="Sensor">
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%", width: "98vw" }}>
          <Grid size={6}>
            <SensorInfoCard sensor={sensor} />
          </Grid>
          <Grid size={6}>
            <EditSensorDetails sensor={sensor} />
          </Grid>
          <Grid size={6}>
            <SensorHealthHistory sensor={sensor} />
          </Grid>
        </Grid>
      </Box>
    </PageContainer>
  )
}

export default SensorPage;