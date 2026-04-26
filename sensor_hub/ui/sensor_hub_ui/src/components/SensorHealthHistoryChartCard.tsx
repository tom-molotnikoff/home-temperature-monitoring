import { useState, type CSSProperties } from 'react';
import {
  Box, Dialog, DialogTitle, DialogContent, DialogActions,
  Button, TextField, IconButton,
} from '@mui/material';
import SettingsIcon from '@mui/icons-material/Settings';
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
  const [limit, setLimit] = useState(1000);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [draftLimit, setDraftLimit] = useState('1000');

  const handleOpen = () => {
    setDraftLimit(limit.toString());
    setSettingsOpen(true);
  };

  const handleSave = () => {
    const parsed = parseInt(draftLimit);
    setLimit(Number.isFinite(parsed) && parsed > 0 ? parsed : 1000);
    setSettingsOpen(false);
  };

  return (
    <LayoutCard variant="secondary" changes={graphContainerStyle}>
      <Box display="flex" alignItems="center" justifyContent="space-between" width="100%">
        <TypographyH2>Sensor Health History</TypographyH2>
        <IconButton onClick={handleOpen} size="small" title="Settings">
          <SettingsIcon />
        </IconButton>
      </Box>
      <SensorHealthHistoryChart sensor={sensor} limit={limit} />

      <Dialog open={settingsOpen} onClose={() => setSettingsOpen(false)} maxWidth="xs" fullWidth>
        <DialogTitle>Health Timeline Settings</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          <TextField
            label="History Limit"
            type="number"
            fullWidth
            sx={{ mt: 1 }}
            value={draftLimit}
            onChange={(e) => setDraftLimit(e.target.value)}
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
