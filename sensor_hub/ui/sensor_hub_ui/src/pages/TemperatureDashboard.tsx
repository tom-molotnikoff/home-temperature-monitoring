import {useState, type CSSProperties} from "react";
import CurrentTemperatures from "../components/CurrentTemperatures";
import DateRangePicker from "../components/DateRangePicker";
import SensorTriggerButtons from "../components/SensorTriggerButtons";
import TemperatureGraph from "../components/TemperatureGraph";
import { DateContextProvider } from "../providers/DateContextProvider";
import PageContainer from "../tools/PageContainer";
import HourlyAveragesToggle from "../components/HourlyAveragesToggle";
import LayoutCard from "../tools/LayoutCard.tsx";
import { TypographyH2 } from "../tools/Typography";
import { useSensorContext } from "../hooks/useSensorContext";
import SensorSummaryCard from "../components/SensorSummaryCard.tsx";
import {useIsMobile} from "../hooks/useMobile.ts";

function TemperatureDashboard() {
  const [useHourlyAverages, setUseHourlyAverages] = useState(true);

  const { sensors } = useSensorContext();
  const isMobile = useIsMobile();

  return (
    <DateContextProvider>
      <PageContainer titleText="Temperature Dashboard">
        <LayoutCard variant="primary" direction={isMobile ? "column" : "row"} changes={{padding: 0, border: "none", gap: 20}}>
          <SensorSummaryCard sensors={sensors} />
          <LayoutCard variant="secondary">
            <CurrentTemperatures />
            <SensorTriggerButtons sensors={sensors} />
          </LayoutCard>
        </LayoutCard>
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
      </PageContainer>
    </DateContextProvider>
  );
}

const graphContainerStyle: CSSProperties = {
  width: "95%",
  alignItems: "center",
};

export default TemperatureDashboard;
