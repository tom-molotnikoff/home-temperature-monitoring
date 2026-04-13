-- Migration 000012: Add battery_low measurement type for Zigbee2MQTT devices.
INSERT OR IGNORE INTO measurement_types (name, display_name, category, default_unit) VALUES
    ('battery_low', 'Battery Low', 'binary', '');
