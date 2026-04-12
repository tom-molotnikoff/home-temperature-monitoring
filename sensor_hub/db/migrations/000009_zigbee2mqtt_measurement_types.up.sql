-- Migration 000009: Seed measurement types required by the Zigbee2MQTT driver.
-- Migration 000006 seeded a small initial set; this adds the rest so that
-- incoming MQTT readings can be resolved without errors.

INSERT OR IGNORE INTO measurement_types (name, display_name, category, default_unit) VALUES
    ('link_quality', 'Link Quality', 'numeric', 'lqi'),
    ('illuminance', 'Illuminance', 'numeric', 'lx'),
    ('energy', 'Energy', 'numeric', 'kWh'),
    ('current', 'Current', 'numeric', 'A'),
    ('co2', 'CO₂', 'numeric', 'ppm'),
    ('voc', 'VOC', 'numeric', 'ppb'),
    ('formaldehyde', 'Formaldehyde', 'numeric', 'mg/m³'),
    ('pm25', 'PM2.5', 'numeric', 'µg/m³'),
    ('soil_moisture', 'Soil Moisture', 'numeric', '%'),
    ('occupancy', 'Occupancy', 'binary', ''),
    ('water_leak', 'Water Leak', 'binary', ''),
    ('smoke', 'Smoke', 'binary', ''),
    ('carbon_monoxide', 'Carbon Monoxide', 'binary', ''),
    ('tamper', 'Tamper', 'binary', ''),
    ('vibration', 'Vibration', 'binary', ''),
    ('state', 'State', 'binary', '');
