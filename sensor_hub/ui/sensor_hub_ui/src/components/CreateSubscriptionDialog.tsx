import { useState, useEffect } from 'react';
import {
  Button, Dialog, DialogActions, DialogContent, DialogTitle,
  TextField, FormControlLabel, Switch, FormControl, InputLabel, Select, MenuItem,
} from '@mui/material';
import { MqttBrokersApi, MqttSubscriptionsApi } from '../api/Mqtt';
import type { CreateSubscriptionPayload } from '../api/Mqtt';
import type { MQTTBroker } from '../types/types';
import { logger } from '../tools/logger';

interface Props {
  open: boolean;
  onClose: () => void;
  onCreated: () => Promise<void>;
}

const KNOWN_DRIVERS = [
  { value: 'mqtt-zigbee2mqtt', label: 'Zigbee2MQTT' },
];

export default function CreateSubscriptionDialog({ open, onClose, onCreated }: Props) {
  const [brokerId, setBrokerId] = useState<number>(0);
  const [topicPattern, setTopicPattern] = useState('');
  const [driverType, setDriverType] = useState('mqtt-zigbee2mqtt');
  const [enabled, setEnabled] = useState(true);
  const [brokers, setBrokers] = useState<MQTTBroker[]>([]);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!open) return;
    MqttBrokersApi.list()
      .then(b => {
        setBrokers(b || []);
        if (b.length > 0 && brokerId === 0) setBrokerId(b[0].id);
      })
      .catch(e => logger.error('Failed to load brokers', e));
  }, [open]);

  const reset = () => {
    setBrokerId(0); setTopicPattern(''); setDriverType('mqtt-zigbee2mqtt');
    setEnabled(true); setError('');
  };

  const handleCreate = async () => {
    setError('');
    const payload: CreateSubscriptionPayload = {
      broker_id: brokerId,
      topic_pattern: topicPattern,
      driver_type: driverType,
      enabled,
    };
    try {
      await MqttSubscriptionsApi.create(payload);
      reset();
      onClose();
      await onCreated();
    } catch (e: unknown) {
      const msg = (e as { message?: string })?.message || 'Failed to create subscription';
      setError(msg);
      logger.error('Failed to create subscription', e);
    }
  };

  const handleCancel = () => { reset(); onClose(); };

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
      <DialogTitle>Add MQTT Subscription</DialogTitle>
      <DialogContent>
        <FormControl fullWidth sx={{ mt: 1 }}>
          <InputLabel>Broker</InputLabel>
          <Select value={brokerId || ''} label="Broker" onChange={e => setBrokerId(Number(e.target.value))}>
            {brokers.map(b => (
              <MenuItem key={b.id} value={b.id}>{b.name} ({b.host}:{b.port})</MenuItem>
            ))}
          </Select>
        </FormControl>
        <TextField fullWidth label="Topic Pattern" value={topicPattern}
          onChange={e => setTopicPattern(e.target.value)} sx={{ mt: 1 }} required
          helperText="e.g. zigbee2mqtt/# or rtl_433/+/events" />
        <FormControl fullWidth sx={{ mt: 1 }}>
          <InputLabel>Driver</InputLabel>
          <Select value={driverType} label="Driver" onChange={e => setDriverType(e.target.value)}>
            {KNOWN_DRIVERS.map(d => (
              <MenuItem key={d.value} value={d.value}>{d.label}</MenuItem>
            ))}
          </Select>
        </FormControl>
        <FormControlLabel control={<Switch checked={enabled} onChange={e => setEnabled(e.target.checked)} />}
          label="Enabled" sx={{ mt: 1 }} />
        {error && <p style={{ color: 'red', marginTop: 8 }}>{error}</p>}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel}>Cancel</Button>
        <Button variant="contained" onClick={handleCreate} disabled={!topicPattern || !brokerId}>Create</Button>
      </DialogActions>
    </Dialog>
  );
}
