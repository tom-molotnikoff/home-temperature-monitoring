import { useState, type CSSProperties } from 'react';
import { DateContextProvider } from '../providers/DateContextProvider';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import DateRangePicker from './DateRangePicker';
import HourlyAveragesToggle from './HourlyAveragesToggle';
import TemperatureGraph from './TemperatureGraph';
import { useSensorContext } from '../hooks/useSensorContext';
import { useIsMobile } from '../hooks/useMobile';

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: '100%',
  alignItems: 'center',
};

export default function IndoorTemperatureDataCard() {
  const [useHourlyAverages, setUseHourlyAverages] = useState(true);
  const { sensors } = useSensorContext();
  const isMobile = useIsMobile();

  return (
    <DateContextProvider>
      <LayoutCard variant="secondary" changes={graphContainerStyle}>
        <TypographyH2>Indoor Temperature Data</TypographyH2>
        <DateRangePicker />
        <HourlyAveragesToggle
          useHourlyAverages={useHourlyAverages}
          setUseHourlyAverages={setUseHourlyAverages}
        />
        <TemperatureGraph
          sensors={sensors}
          useHourlyAverages={useHourlyAverages}
          compact={isMobile}
        />
      </LayoutCard>
    </DateContextProvider>
  );
}
