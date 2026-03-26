import {Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, Typography} from "@mui/material";
import {useEffect, useState} from "react";
import type {AlertHistory, AlertRule} from "../api/Alerts";
import { getAlertHistory } from "../api/Alerts";
import { logger } from '../tools/logger';

interface AlertHistoryDialogProps {
  open: boolean;
  onClose: () => void;
  selectedAlert: AlertRule | null;
}

export default function AlertHistoryDialog({open, onClose, selectedAlert}: AlertHistoryDialogProps) {
  const [historyLoading, setHistoryLoading] = useState(false);
  const [historyData, setHistoryData] = useState<AlertHistory[]>([]);

  useEffect(() => {
    if (!open || !selectedAlert) return;
    let cancelled = false;
    const fetchHistory = async () => {
      setHistoryLoading(true);
      try {
        const history = await getAlertHistory(selectedAlert.SensorID, 50);
        if (!cancelled) setHistoryData(history);
      } catch (e) {
        logger.error('Failed to load alert history', e);
        if (!cancelled) setHistoryData([]);
      } finally {
        if (!cancelled) setHistoryLoading(false);
      }
    };
    fetchHistory();
    return () => { cancelled = true; };
  }, [open, selectedAlert]);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>Alert History - {selectedAlert?.SensorName}</DialogTitle>
      <DialogContent>
        {historyLoading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
            Loading...
          </Box>
        ) : historyData.length === 0 ? (
          <Box sx={{ p: 2 }}>
            <Typography>No alert history found for this sensor.</Typography>
          </Box>
        ) : (
          <Box sx={{ mt: 1 }}>
            {historyData.map((h) => (
              <Box
                key={h.id}
                sx={{
                  p: 2,
                  mb: 1,
                  border: '1px solid',
                  borderColor: 'divider',
                  borderRadius: 1,
                }}
              >
                <Typography variant="body2">
                  <strong>Type:</strong> {h.alert_type}
                </Typography>
                <Typography variant="body2">
                  <strong>Value:</strong> {h.reading_value}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  <strong>Sent:</strong> {new Date(h.sent_at).toLocaleString()}
                </Typography>
              </Box>
            ))}
          </Box>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}