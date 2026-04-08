-- Reverse config bag migration: restore url column from config

ALTER TABLE sensors ADD COLUMN url TEXT DEFAULT NULL;

UPDATE sensors SET url = json_extract(config, '$.url') WHERE json_extract(config, '$.url') IS NOT NULL;

ALTER TABLE sensors DROP COLUMN config;
