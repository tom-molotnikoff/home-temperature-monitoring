import type { CSSProperties } from 'react';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import SensorHealthHistoryChart from './SensorHealthHistoryChart';
import type { Sensor } from '../api/Sensors';

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: '100%',
  alignItems: 'center',
};

interface SensorHealthHistoryChartCardProps {
  sensor: Sensor;
}

export default function SensorHealthHistoryChartCard({ sensor }: SensorHealthHistoryChartCardProps) {
  return (
    <LayoutCard variant="secondary" changes={graphContainerStyle}>
      <TypographyH2>Sensor Health History</TypographyH2>
      <SensorHealthHistoryChart sensor={sensor} limit={5000} />
    </LayoutCard>
  );
}
