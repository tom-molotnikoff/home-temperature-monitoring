import { Box, Grid } from "@mui/material";
import PageContainer from "../../tools/PageContainer.tsx";
import { useSensorContext } from "../../hooks/useSensorContext.ts";
import SensorInfoCard from "../../components/SensorInfoCard.tsx";
import EditSensorDetails from "../../components/EditSensorDetails.tsx";
import SensorHealthHistory from "../../components/SensorHealthHistory.tsx";
import {DateContextProvider} from "../../providers/DateContextProvider.tsx";
import LayoutCard from "../../tools/LayoutCard.tsx";
import {TypographyH2} from "../../tools/Typography.tsx";
import DateRangePicker from "../../components/DateRangePicker.tsx";
import HourlyAveragesToggle from "../../components/HourlyAveragesToggle.tsx";
import TemperatureGraph from "../../components/TemperatureGraph.tsx";
import {type CSSProperties, useState} from "react";
import SensorHealthHistoryChart from "../../components/SensorHealthHistoryChart.tsx";
import {useAuth} from "../../providers/AuthContext.tsx";
import {hasPerm} from "../../tools/Utils.ts";

interface SensorPageProps {
  sensorId: number;
}

function SensorPage({sensorId}: SensorPageProps) {
  const {sensors} = useSensorContext();
  const [useHourlyAverages, setUseHourlyAverages] = useState(true);
  const sensor = sensors.find(s => s.id === sensorId);
  const { user } = useAuth();

  if (user === undefined) {
    return (
      <PageContainer titleText="Sensor">
        <Box sx={{ flexGrow: 1 }}>
          <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%" }}>
            <Grid size={12}>
              Loading...
            </Grid>
          </Grid>
        </Box>
      </PageContainer>
    );
  }

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
          {(hasPerm(user, "view_sensors") &&
            <Grid size={6}>
              <SensorInfoCard sensor={sensor} user={user} />
            </Grid>
          )}
          {(hasPerm(user, "view_sensors") &&
            <Grid size={6}>
              <EditSensorDetails sensor={sensor} />
            </Grid>
          )}
          {(hasPerm(user, "view_readings") &&
            <>
              <Grid size={6}>
                <LayoutCard variant="secondary" changes={graphContainerStyle}>
                  <TypographyH2>Sensor Health History</TypographyH2>
                  <SensorHealthHistoryChart sensor={sensor} limit={5000}/>
                </LayoutCard>
              </Grid>
              <Grid size={6}>
                {sensor.type === "Temperature" &&
                  <DateContextProvider>
                    <LayoutCard variant="secondary" changes={graphContainerStyle}>
                      <TypographyH2>Indoor Temperature Data</TypographyH2>
                      <DateRangePicker/>
                      <HourlyAveragesToggle
                        useHourlyAverages={useHourlyAverages}
                        setUseHourlyAverages={setUseHourlyAverages}/>
                      <TemperatureGraph
                          sensors={[sensor]}
                          useHourlyAverages={useHourlyAverages}/>
                    </LayoutCard>
                  </DateContextProvider>}
              </Grid>
            </>
          )}
          {(hasPerm(user, "view_sensors") &&
            <Grid size={6}>
              <SensorHealthHistory sensor={sensor} />
            </Grid>
          )}
        </Grid>
      </Box>
    </PageContainer>
  )
}

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: "100%",
  alignItems: "center",
};

export default SensorPage;