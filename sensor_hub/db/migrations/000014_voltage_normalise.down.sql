-- Revert voltage back to millivolts.
UPDATE measurement_types SET default_unit = 'mV' WHERE name = 'voltage';

-- Convert readings back: anything that looks like it was battery (< 5V) back to mV.
UPDATE readings
SET numeric_value = numeric_value * 1000.0
WHERE measurement_type_id IN (SELECT id FROM measurement_types WHERE name = 'voltage')
  AND numeric_value < 5;
