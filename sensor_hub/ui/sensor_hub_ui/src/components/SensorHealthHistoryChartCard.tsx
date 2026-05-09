import { type CSSProperties } from 'react';
import { Box } from '@mui/material';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import SensorHealthHistoryChart from './SensorHealthHistoryChart';
import type { Sensor } from '../gen/aliases';

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  minHeight: 400,
  alignItems: 'center',
};

interface SensorHealthHistoryChartCardProps {
  sensor: Sensor;
}

export default function SensorHealthHistoryChartCard({ sensor }: SensorHealthHistoryChartCardProps) {
  return (
    <LayoutCard variant="secondary" changes={graphContainerStyle}>
      <Box display="flex" alignItems="center" justifyContent="space-between" width="100%">
        <TypographyH2>Sensor Health History</TypographyH2>
      </Box>
      <SensorHealthHistoryChart sensor={sensor} />
    </LayoutCard>
  );
}
