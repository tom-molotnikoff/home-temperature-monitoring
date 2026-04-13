-- Migration 000014: Normalise voltage to Volts.
-- Zigbee2MQTT battery devices report in mV (e.g. 3100), mains devices in V
-- (e.g. 231.15). Standardise everything to V so the UI is consistent.

-- Update the measurement type default unit.
UPDATE measurement_types SET default_unit = 'V' WHERE name = 'voltage';

-- Convert existing millivolt readings to volts (any value > 500 is mV).
UPDATE readings
SET numeric_value = numeric_value / 1000.0
WHERE measurement_type_id IN (SELECT id FROM measurement_types WHERE name = 'voltage')
  AND numeric_value > 500;
