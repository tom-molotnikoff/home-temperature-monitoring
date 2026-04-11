import { useEffect, useState, useCallback } from 'react';
import { Box, Chip, Stack, Typography } from '@mui/material';
import Grid from '@mui/material/Grid';
import { MqttStatsApi } from '../api/Mqtt';
import type { MQTTBrokerStats } from '../types/types';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import { logger } from '../tools/logger';
import WifiIcon from '@mui/icons-material/Wifi';
import WifiOffIcon from '@mui/icons-material/WifiOff';
import MessageIcon from '@mui/icons-material/Message';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import DevicesIcon from '@mui/icons-material/Devices';

function formatRelativeTime(iso: string | null): string {
  if (!iso) return 'Never';
  const date = new Date(iso);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  if (diffMs < 0) return 'Just now';
  const secs = Math.floor(diffMs / 1000);
  if (secs < 60) return `${secs}s ago`;
  const mins = Math.floor(secs / 60);
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

function formatUptime(iso: string | null): string {
  if (!iso) return '—';
  const date = new Date(iso);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  if (diffMs < 0) return 'Just connected';
  const secs = Math.floor(diffMs / 1000);
  if (secs < 60) return `${secs}s`;
  const mins = Math.floor(secs / 60);
  if (mins < 60) return `${mins}m ${secs % 60}s`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ${mins % 60}m`;
  return `${Math.floor(hrs / 24)}d ${hrs % 24}h`;
}

function BrokerStatCard({ stat }: { stat: MQTTBrokerStats }) {
  const totalErrors = stat.parse_errors + stat.processing_errors;
  return (
    <Grid size={{ xs: 12, sm: 6, md: 4 }}>
      <Box sx={{
        p: 2,
        borderRadius: 1,
        bgcolor: 'background.default',
        border: 1,
        borderColor: stat.connected ? 'success.main' : 'error.main',
        height: '100%',
      }}>
        <Stack direction="row" alignItems="center" spacing={1} mb={1.5}>
          {stat.connected
            ? <WifiIcon color="success" fontSize="small" />
            : <WifiOffIcon color="error" fontSize="small" />}
          <Typography variant="subtitle1" fontWeight={600} noWrap>
            {stat.broker_name || `Broker ${stat.broker_id}`}
          </Typography>
          <Chip
            size="small"
            label={stat.connected ? 'Connected' : 'Disconnected'}
            color={stat.connected ? 'success' : 'error'}
            variant="outlined"
          />
        </Stack>

        <Stack spacing={0.75}>
          <Stack direction="row" alignItems="center" spacing={1}>
            <MessageIcon sx={{ fontSize: 16, color: 'text.secondary' }} />
            <Typography variant="body2" color="text.secondary">
              Messages: <strong>{stat.messages_received.toLocaleString()}</strong>
            </Typography>
          </Stack>

          {totalErrors > 0 && (
            <Stack direction="row" alignItems="center" spacing={1}>
              <ErrorOutlineIcon sx={{ fontSize: 16, color: 'warning.main' }} />
              <Typography variant="body2" color="warning.main">
                Errors: <strong>{totalErrors}</strong>
                {stat.parse_errors > 0 && ` (${stat.parse_errors} parse)`}
                {stat.processing_errors > 0 && ` (${stat.processing_errors} processing)`}
              </Typography>
            </Stack>
          )}

          {stat.devices_discovered > 0 && (
            <Stack direction="row" alignItems="center" spacing={1}>
              <DevicesIcon sx={{ fontSize: 16, color: 'text.secondary' }} />
              <Typography variant="body2" color="text.secondary">
                Devices discovered: <strong>{stat.devices_discovered}</strong>
              </Typography>
            </Stack>
          )}

          <Typography variant="body2" color="text.secondary">
            Last message: {formatRelativeTime(stat.last_message_at)}
          </Typography>

          {stat.connected && (
            <Typography variant="body2" color="text.secondary">
              Uptime: {formatUptime(stat.connected_since)}
            </Typography>
          )}
        </Stack>
      </Box>
    </Grid>
  );
}

export default function MqttStatsCard() {
  const [stats, setStats] = useState<MQTTBrokerStats[]>([]);
  const [, setTick] = useState(0);

  const load = useCallback(async () => {
    try {
      const s = await MqttStatsApi.list();
      setStats(s ?? []);
    } catch (e) { logger.error(e); }
  }, []);

  useEffect(() => {
    load();
    const interval = setInterval(() => {
      load();
      setTick(t => t + 1);
    }, 10_000);
    return () => clearInterval(interval);
  }, [load]);

  return (
    <LayoutCard>
      <TypographyH2>MQTT Broker Stats</TypographyH2>
      {stats.length === 0 ? (
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          No MQTT broker statistics available. Connect a broker to see live stats.
        </Typography>
      ) : (
        <Grid container spacing={2} sx={{ mt: 0.5 }}>
          {stats.map(s => <BrokerStatCard key={s.broker_id} stat={s} />)}
        </Grid>
      )}
    </LayoutCard>
  );
}
