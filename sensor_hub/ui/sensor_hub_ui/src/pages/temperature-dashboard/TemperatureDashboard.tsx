import {useState, type CSSProperties} from "react";
import CurrentTemperatures from "../../components/CurrentTemperatures.tsx";
import DateRangePicker from "../../components/DateRangePicker.tsx";
//import SensorTriggerButtons from "../../components/SensorTriggerButtons.tsx";
import TemperatureGraph from "../../components/TemperatureGraph.tsx";
import { DateContextProvider } from "../../providers/DateContextProvider.tsx";
import PageContainer from "../../tools/PageContainer.tsx";
import HourlyAveragesToggle from "../../components/HourlyAveragesToggle.tsx";
import LayoutCard from "../../tools/LayoutCard.tsx";
import { TypographyH2 } from "../../tools/Typography.tsx";
import { useSensorContext } from "../../hooks/useSensorContext.tsx";
import SensorSummaryCard from "../../components/SensorSummaryCard.tsx";
import {useIsMobile} from "../../hooks/useMobile.ts";
import {Box, Grid} from "@mui/material";

function TemperatureDashboard() {
  const [useHourlyAverages, setUseHourlyAverages] = useState(true);
  const { sensors } = useSensorContext();

  const temperatureSensors = sensors.filter(sensor => sensor.type === "Temperature");

  const isMobile = useIsMobile();

  return (
    <DateContextProvider>
      <PageContainer titleText="Temperature Dashboard">
        <Box sx={{ flexGrow: 1 }}>

          <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%" }}>
            {isMobile ? null : (
              <Grid size={12} sx={{width: "100%"}}>
                <LayoutCard variant="secondary" changes={graphContainerStyle}>
                  <TypographyH2>Temperature Over Time</TypographyH2>
                  <DateRangePicker />
                  <HourlyAveragesToggle
                    useHourlyAverages={useHourlyAverages}
                    setUseHourlyAverages={setUseHourlyAverages}
                  />
                  <TemperatureGraph
                    sensors={sensors}
                    useHourlyAverages={useHourlyAverages}
                  />
                </LayoutCard>
              </Grid>
            )}
            <Grid size={isMobile ? 12 : 6}>
              <SensorSummaryCard sensors={temperatureSensors} cardHeight={"100%"} showReason={false} showType={false} title="Temperature Sensors" />
            </Grid>
            <Grid size={isMobile ? 12 : 6}>
                <CurrentTemperatures cardHeight={"100%"} />
            </Grid>

          </Grid>
        </Box>
      </PageContainer>
    </DateContextProvider>
  );
}

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: "100%",
  alignItems: "center",
};

export default TemperatureDashboard;
