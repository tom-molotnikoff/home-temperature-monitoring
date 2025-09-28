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
        <div style={optionsTopRightStyle}>
          <HourlyAveragesToggle
            useHourlyAverages={useHourlyAverages}
            setUseHourlyAverages={setUseHourlyAverages}
          />
        </div>

        <CurrentTemperatures />
        <SensorTriggerButtons sensors={sensors} />
        <DateRangePicker />
        <TemperatureGraph
          sensors={sensors}
          useHourlyAverages={useHourlyAverages}
        />
      </PageContainer>
    </DateContextProvider>
  );
}

const optionsTopRightStyle: CSSProperties = {
  position: "absolute",
  top: 24,
  right: 24,
  display: "flex",
  alignItems: "center",
  gap: 8,
};

export default TemperatureDashboard;
