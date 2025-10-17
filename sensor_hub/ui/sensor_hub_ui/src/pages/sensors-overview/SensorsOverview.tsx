import PageContainer from "../../tools/PageContainer.tsx";
import {useIsMobile} from "../../hooks/useMobile.ts";
import SensorSummaryCard from "../../components/SensorSummaryCard.tsx";
import {useSensorContext} from "../../hooks/useSensorContext.tsx";
import {Grid, Box} from "@mui/material";
import SensorHealthCard from "../../components/SensorHealthCard.tsx";


function SensorsOverview () {

  const isMobile = useIsMobile();
  const { sensors } = useSensorContext();


  return (
    <PageContainer titleText="Sensors Overview">
        <Box sx={{ flexGrow: 1 }}>
          <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%" }}>
            <Grid size={isMobile ? 12 : 4} >
              <SensorHealthCard />
            </Grid>
            <Grid size={12} >
              <SensorSummaryCard sensors={sensors} showReason={true} showType={true}/>
            </Grid>
          </Grid>
        </Box>
    </PageContainer>
    );
}

export default SensorsOverview