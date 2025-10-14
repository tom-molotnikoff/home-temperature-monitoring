CREATE TABLE IF NOT EXISTS sensors (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    type VARCHAR(32) DEFAULT NULL,
    url VARCHAR(255) DEFAULT NULL
);

CREATE INDEX idx_sensor_type ON sensors (type);

INSERT INTO sensors (name)
SELECT DISTINCT sensor_name FROM temperature_readings;

ALTER TABLE temperature_readings ADD COLUMN sensor_id INT;
ALTER TABLE hourly_avg_temperature ADD COLUMN sensor_id INT;

UPDATE temperature_readings tr
JOIN sensors s ON tr.sensor_name = s.name
SET tr.sensor_id = s.id;

UPDATE hourly_avg_temperature hat
JOIN sensors s ON hat.sensor_name = s.name
SET hat.sensor_id = s.id;

ALTER TABLE temperature_readings MODIFY sensor_id INT NOT NULL;
ALTER TABLE hourly_avg_temperature MODIFY sensor_id INT NOT NULL;

ALTER TABLE temperature_readings
    ADD CONSTRAINT fk_temperature_readings_sensor
        FOREIGN KEY (sensor_id) REFERENCES sensors(id);

ALTER TABLE hourly_avg_temperature
    ADD CONSTRAINT fk_hourly_avg_temperature_sensor
        FOREIGN KEY (sensor_id) REFERENCES sensors(id);

ALTER TABLE temperature_readings DROP COLUMN sensor_name;
ALTER TABLE hourly_avg_temperature DROP COLUMN sensor_name;