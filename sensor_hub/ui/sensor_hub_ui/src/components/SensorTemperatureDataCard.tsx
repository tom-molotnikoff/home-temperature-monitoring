import { useState, type CSSProperties } from 'react';
import { DateContextProvider } from '../providers/DateContextProvider';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import DateRangePicker from './DateRangePicker';
import HourlyAveragesToggle from './HourlyAveragesToggle';
import TemperatureGraph from './TemperatureGraph';
import { useIsMobile } from '../hooks/useMobile';
import type { Sensor } from '../api/Sensors';

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: '100%',
  alignItems: 'center',
};

interface SensorTemperatureDataCardProps {
  sensor: Sensor;
}

export default function SensorTemperatureDataCard({ sensor }: SensorTemperatureDataCardProps) {
  const [useHourlyAverages, setUseHourlyAverages] = useState(true);
  const isMobile = useIsMobile();

  if (sensor.type !== 'Temperature') return null;

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
          sensors={[sensor]}
          useHourlyAverages={useHourlyAverages}
          compact={isMobile}
        />
      </LayoutCard>
    </DateContextProvider>
  );
}
