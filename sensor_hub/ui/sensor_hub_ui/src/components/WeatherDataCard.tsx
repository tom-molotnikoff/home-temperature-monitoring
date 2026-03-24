import type { CSSProperties } from 'react';
import { DateContextProvider } from '../providers/DateContextProvider';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import DateRangePicker from './DateRangePicker';
import WeatherChart from './WeatherGraph';
import { useIsMobile } from '../hooks/useMobile';

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: '100%',
  alignItems: 'center',
};

export default function WeatherDataCard() {
  const isMobile = useIsMobile();

  return (
    <DateContextProvider>
      <LayoutCard variant="secondary" changes={graphContainerStyle}>
        <TypographyH2>Sheffield Weather Data</TypographyH2>
        <DateRangePicker />
        <WeatherChart compact={isMobile} />
      </LayoutCard>
    </DateContextProvider>
  );
}
