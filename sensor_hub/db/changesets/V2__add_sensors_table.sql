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
ALTER TABLE hourly_avg_temperature DROP INDEX unique_sensor_hour;
ALTER TABLE hourly_avg_temperature DROP COLUMN sensor_name;



ALTER TABLE hourly_avg_temperature ADD UNIQUE KEY unique_sensor_hour (sensor_id, time);
CREATE INDEX idx_sensor_id ON temperature_readings (sensor_id);
CREATE INDEX hourly_idx_sensor_id ON hourly_avg_temperature (sensor_id);

DROP EVENT IF EXISTS hourly_average_temperature_event;

CREATE EVENT IF NOT EXISTS hourly_average_temperature_event
ON SCHEDULE EVERY 1 HOUR
STARTS TIMESTAMP(CURRENT_DATE, SEC_TO_TIME((HOUR(NOW())+1)*3600 + 60))
DO
    INSERT INTO hourly_avg_temperature (sensor_id, time, average_temperature)
    SELECT
        tr.sensor_id,
        DATE_FORMAT(tr.time, '%Y-%m-%d %H:00:00') AS hour,
            ROUND(AVG(tr.temperature), 2) AS avg_temp
    FROM temperature_readings tr
    WHERE tr.time >= DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 HOUR), '%Y-%m-%d %H:00:00')
      AND tr.time < DATE_FORMAT(NOW(), '%Y-%m-%d %H:00:00')
    GROUP BY tr.sensor_id, hour
    HAVING NOT EXISTS (
        SELECT 1
        FROM hourly_avg_temperature hat
        WHERE hat.sensor_id = tr.sensor_id
            AND hat.time = hour
    );