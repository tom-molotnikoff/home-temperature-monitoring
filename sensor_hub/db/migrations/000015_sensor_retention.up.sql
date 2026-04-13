-- Migration 000015: Per-sensor data retention
-- Adds nullable retention_hours column to sensors table.
-- NULL means use the global default (sensor.data.retention.days config).
ALTER TABLE sensors ADD COLUMN retention_hours INTEGER DEFAULT NULL;
