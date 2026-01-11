import { useState, type CSSProperties } from "react";
import CurrentTemperatures from "../../components/CurrentTemperatures.tsx";
import DateRangePicker from "../../components/DateRangePicker.tsx";
import TemperatureGraph from "../../components/TemperatureGraph.tsx";
import { DateContextProvider } from "../../providers/DateContextProvider.tsx";
import PageContainer from "../../tools/PageContainer.tsx";
import HourlyAveragesToggle from "../../components/HourlyAveragesToggle.tsx";
import LayoutCard from "../../tools/LayoutCard.tsx";
import { TypographyH2 } from "../../tools/Typography.tsx";
import { useSensorContext } from "../../hooks/useSensorContext.ts";
import SensorsDataGrid from "../../components/SensorsDataGrid.tsx";
import { useIsMobile } from "../../hooks/useMobile.ts";
import { Box, Grid } from "@mui/material";
import WeatherChart from "../../components/WeatherGraph.tsx";
import {useAuth} from "../../providers/AuthContext.tsx";
import {hasPerm} from "../../tools/Utils.ts";

function TemperatureDashboard() {
  const [useHourlyAverages, setUseHourlyAverages] = useState(true);
  const { sensors } = useSensorContext();
  const { user } = useAuth();

  const temperatureSensors = sensors.filter(
    (sensor) => sensor.type === "Temperature"
  );

  const isMobile = useIsMobile();

  if (user === undefined) {
    return (
      <PageContainer titleText="Temperature Dashboard">
        <Box sx={{ flexGrow: 1 }}>
          <Grid
            container
            spacing={2}
            alignItems="stretch"
            sx={{ minHeight: "100%" }}
          >
            <Grid size={12}>
              Loading...
            </Grid>
          </Grid>
        </Box>
      </PageContainer>
    );
  }

  return (
    <PageContainer titleText="Temperature Dashboard">
      <Box sx={{ flexGrow: 1 }}>
        <Grid
          container
          spacing={2}
          alignItems="stretch"
          sx={{ minHeight: "100%" }}
        >
          {isMobile ? null : (
            hasPerm(user, "view_readings") &&
              <>
                <Grid size={12} sx={{width: "98vw"}}>
                  <DateContextProvider>
                    <LayoutCard variant="secondary" changes={graphContainerStyle}>
                      <TypographyH2>Indoor Temperature Data</TypographyH2>
                      <DateRangePicker/>
                      <HourlyAveragesToggle
                        useHourlyAverages={useHourlyAverages}
                        setUseHourlyAverages={setUseHourlyAverages}/>
                      <TemperatureGraph
                        sensors={sensors}
                        useHourlyAverages={useHourlyAverages}/>
                    </LayoutCard>
                  </DateContextProvider>
                </Grid>
                <Grid size={12} sx={{width: "98vw"}}>
                  <DateContextProvider>
                    <LayoutCard variant="secondary" changes={graphContainerStyle}>
                      <TypographyH2>Sheffield Weather Data</TypographyH2>
                      <DateRangePicker/>
                      <WeatherChart/>
                    </LayoutCard>
                  </DateContextProvider>
                </Grid>
              </>
          )}
          {(hasPerm(user, "view_sensors") &&
            <Grid size={isMobile ? 12 : 6}>
              <SensorsDataGrid
                sensors={temperatureSensors}
                cardHeight={"100%"}
                showReason={false}
                showType={false}
                showEnabled={true}
                title="Temperature Sensors"
                user={user}
              />
            </Grid>
          )}

          { (hasPerm(user, "view_readings") &&
            <Grid size={isMobile ? 12 : 6}>
              <CurrentTemperatures cardHeight={"100%"} />
            </Grid>
          )}

        </Grid>
      </Box>
    </PageContainer>
  );
}

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: "100%",
  alignItems: "center",
};

export default TemperatureDashboard;
