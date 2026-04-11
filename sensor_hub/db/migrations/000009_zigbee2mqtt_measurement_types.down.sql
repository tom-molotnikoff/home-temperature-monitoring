-- Rollback: remove measurement types added in 000009.
-- Only deletes types that have no readings referencing them.
DELETE FROM measurement_types WHERE name IN (
    'link_quality', 'illuminance', 'energy', 'current',
    'co2', 'voc', 'formaldehyde', 'pm25', 'soil_moisture',
    'occupancy', 'water_leak', 'smoke', 'carbon_monoxide',
    'tamper', 'vibration', 'state'
) AND id NOT IN (SELECT DISTINCT measurement_type_id FROM readings);
