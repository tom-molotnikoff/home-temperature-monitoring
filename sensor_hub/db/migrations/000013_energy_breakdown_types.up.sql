-- Migration 000013: Add energy breakdown measurement types for smart plugs.
INSERT OR IGNORE INTO measurement_types (name, display_name, category, default_unit) VALUES
    ('energy_today', 'Energy Today', 'numeric', 'kWh'),
    ('energy_month', 'Energy Month', 'numeric', 'kWh'),
    ('energy_yesterday', 'Energy Yesterday', 'numeric', 'kWh');
