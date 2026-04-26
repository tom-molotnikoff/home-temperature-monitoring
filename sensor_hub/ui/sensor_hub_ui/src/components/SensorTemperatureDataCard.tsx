import { useState, type CSSProperties } from 'react';
import {
  Box, Dialog, DialogTitle, DialogContent, DialogActions,
  Button, IconButton,
} from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers';
import { DateTime } from 'luxon';
import SettingsIcon from '@mui/icons-material/Settings';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import ReadingsChart from './ReadingsChart';
import { useIsMobile } from '../hooks/useMobile';
import type { Sensor } from '../gen/aliases';

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  minHeight: 400,
  alignItems: 'center',
};

interface SensorTemperatureDataCardProps {
  sensor: Sensor;
}

export default function SensorTemperatureDataCard({ sensor }: SensorTemperatureDataCardProps) {
  const isMobile = useIsMobile();

  const [startDate, setStartDate] = useState<DateTime | null>(DateTime.now().startOf('day'));
  const [endDate, setEndDate] = useState<DateTime | null>(DateTime.now().plus({ days: 1 }).startOf('day'));

  const [settingsOpen, setSettingsOpen] = useState(false);
  const [draftStart, setDraftStart] = useState<DateTime | null>(startDate);
  const [draftEnd, setDraftEnd] = useState<DateTime | null>(endDate);

  if (sensor.sensor_driver !== 'sensor-hub-http-temperature') return null;

  const handleOpen = () => {
    setDraftStart(startDate);
    setDraftEnd(endDate);
    setSettingsOpen(true);
  };

  const handleSave = () => {
    setStartDate(draftStart);
    setEndDate(draftEnd);
    setSettingsOpen(false);
  };

  return (
    <LayoutCard variant="secondary" changes={graphContainerStyle}>
      <Box display="flex" alignItems="center" justifyContent="space-between" width="100%">
        <TypographyH2>Indoor Temperature Data</TypographyH2>
        <IconButton onClick={handleOpen} size="small" title="Settings">
          <SettingsIcon />
        </IconButton>
      </Box>
      <ReadingsChart
        sensors={[sensor]}
        startDate={startDate}
        endDate={endDate}
        compact={isMobile}
      />

      <Dialog open={settingsOpen} onClose={() => setSettingsOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Temperature Chart Settings</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          <DatePicker
            label="Start Date"
            value={draftStart}
            onChange={setDraftStart}
            slotProps={{ textField: { fullWidth: true, sx: { mt: 1 } } }}
          />
          <DatePicker
            label="End Date"
            value={draftEnd}
            onChange={setDraftEnd}
            slotProps={{ textField: { fullWidth: true } }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSettingsOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleSave}>Apply</Button>
        </DialogActions>
      </Dialog>
    </LayoutCard>
  );
}
