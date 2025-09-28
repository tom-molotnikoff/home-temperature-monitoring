import { useState, type CSSProperties } from "react";
import CurrentTemperatures from "../components/CurrentTemperatures";
import DateRangePicker from "../components/DateRangePicker";
import SensorTriggerButtons from "../components/SensorTriggerButtons";
import TemperatureGraph from "../components/TemperatureGraph";
import { DateContextProvider } from "../providers/DateContextProvider";
import PageContainer from "../components/PageContainer";
import HourlyAveragesToggle from "../components/HourlyAveragesToggle";

function TemperatureDashboard() {
  const [useHourlyAverages, setUseHourlyAverages] = useState(true);

  // Eventually this needs to be dynamic and fetched from the backend
  const sensors = ["Downstairs", "Upstairs"];

  return (
    <DateContextProvider>
      <PageContainer titleText="Temperature Dashboard">
        <div style={shadowedCardStyle}>
          <CurrentTemperatures />
          <SensorTriggerButtons sensors={sensors} />
        </div>

        <div style={graphContainerStyle}>
          <h2>Temperature Over Time</h2>
          <DateRangePicker />
          <HourlyAveragesToggle
            useHourlyAverages={useHourlyAverages}
            setUseHourlyAverages={setUseHourlyAverages}
          />
          <TemperatureGraph
            sensors={sensors}
            useHourlyAverages={useHourlyAverages}
          />
        </div>
      </PageContainer>
    </DateContextProvider>
  );
}

const shadowedCardStyle: CSSProperties = {
  padding: 8,
  boxShadow: "0 5px 4px rgba(0,0,0,0.07)",
  background: "#fafafaff",
  borderRadius: 12,
  display: "flex",
  flexDirection: "column",
};

const graphContainerStyle: CSSProperties = {
  ...shadowedCardStyle,
  width: "95%",
  alignItems: "center",
};

export default TemperatureDashboard;
