import PageContainer from "../../tools/PageContainer.tsx";
import {useIsMobile} from "../../hooks/useMobile.ts";
import SensorSummaryCard from "../../components/SensorSummaryCard.tsx";
import {useSensorContext} from "../../hooks/useSensorContext.tsx";
import {Grid, Box} from "@mui/material";


function SensorsOverview () {

  const isMobile = useIsMobile();
  const { sensors } = useSensorContext();


  return (
    <PageContainer titleText="Sensors Overview">
        <Box sx={{ flexGrow: 1 }}>
          <Grid container spacing={2}>
            <Grid size={isMobile ? 12 : 4}>
              <SensorSummaryCard sensors={sensors}/>
            </Grid>
            <Grid size={isMobile ? 12 : 4}>
              <SensorSummaryCard sensors={sensors}/>
            </Grid>
            <Grid size={isMobile ? 12 : 4}>
              <SensorSummaryCard sensors={sensors}/>
            </Grid>
          </Grid>
        </Box>
    </PageContainer>
    );
}

export default SensorsOverview