import PageContainer from "../../tools/PageContainer.tsx";
import { useIsMobile } from "../../hooks/useMobile.ts";
import SensorsDataGrid from "../../components/SensorsDataGrid.tsx";
import { useSensorContext } from "../../hooks/useSensorContext.ts";
import { Grid, Box } from "@mui/material";
import SensorHealthCard from "../../components/SensorHealthCard.tsx";
import AddNewSensor from "../../components/AddNewSensor.tsx";
import SensorTypeCard from "../../components/SensorTypeCard.tsx";
import TotalReadingsForEachSensorCard from "../../components/TotalReadingsForEachSensorCard.tsx";

function SensorsOverview() {
  const isMobile = useIsMobile();
  const { sensors } = useSensorContext();

  return (
    <PageContainer titleText="Sensors Overview">
      <Box sx={{ width: '100%', flexGrow: 1 }}>
        <Grid
          container
          spacing={2}
          alignItems="stretch"
          sx={{ minHeight: "100%", width: '100%', flexGrow: 1 }}
        >
          <Grid size={isMobile ? 12 : 4}>
            <AddNewSensor />
          </Grid>
          <Grid size={isMobile ? 12 : 4}>
            <SensorHealthCard />
          </Grid>
          <Grid size={isMobile ? 12 : 4} >
            <SensorTypeCard />
          </Grid>
          <Grid size={8}>
            <SensorsDataGrid
              cardHeight={"500px"}
              sensors={sensors}
              showReason={true}
              showType={true}
              showEnabled={true}
            />
          </Grid>
          <Grid size={4}>
            <TotalReadingsForEachSensorCard />
          </Grid>
        </Grid>
      </Box>
    </PageContainer>
  );
}

export default SensorsOverview;
