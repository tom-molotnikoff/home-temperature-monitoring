import { useState, type CSSProperties } from "react";
import CurrentTemperatures from "../components/CurrentTemperatures";
import DateRangePicker from "../components/DateRangePicker";
import SensorTriggerButtons from "../components/SensorTriggerButtons";
import TemperatureGraph from "../components/TemperatureGraph";
import { DateContextProvider } from "../providers/DateContextProvider";
import PageContainer from "../tools/PageContainer";
import HourlyAveragesToggle from "../components/HourlyAveragesToggle";
import ShadowedColumnCard from "../tools/ShadowedColumnCard";
import { TypographyH2 } from "../tools/Typography";

function TemperatureDashboard() {
  const [useHourlyAverages, setUseHourlyAverages] = useState(true);

  // Eventually this needs to be dynamic and fetched from the backend
  const sensors = ["Downstairs", "Upstairs"];

  return (
    <DateContextProvider>
      <PageContainer titleText="Temperature Dashboard">
        <ShadowedColumnCard>
          <CurrentTemperatures />
          <SensorTriggerButtons sensors={sensors} />
        </ShadowedColumnCard>

        <ShadowedColumnCard changes={graphContainerStyle}>
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
        </ShadowedColumnCard>
      </PageContainer>
    </DateContextProvider>
  );
}

const graphContainerStyle: CSSProperties = {
  width: "95%",
  alignItems: "center",
};

export default TemperatureDashboard;
