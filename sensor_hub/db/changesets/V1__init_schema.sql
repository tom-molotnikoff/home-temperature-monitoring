CREATE TABLE IF NOT EXISTS temperature_readings (
    id INT AUTO_INCREMENT,
    sensor_name TEXT NOT NULL,
    time DATETIME NOT NULL,
    temperature FLOAT(4) NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX idx_time ON temperature_readings (time DESC);

CREATE INDEX idx_sensor_name ON temperature_readings (sensor_name(16));

CREATE TABLE IF NOT EXISTS hourly_avg_temperature (
    id INT AUTO_INCREMENT,
    sensor_name VARCHAR(16) NOT NULL,
    time DATETIME NOT NULL,
    average_temperature FLOAT(4) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY unique_sensor_hour (sensor_name, time)
);

CREATE INDEX hourly_idx_time ON hourly_avg_temperature (time DESC);

CREATE INDEX hourly_idx_sensor_name ON hourly_avg_temperature (sensor_name(16));

CREATE EVENT IF NOT EXISTS hourly_average_temperature_event
ON SCHEDULE EVERY 1 HOUR
STARTS TIMESTAMP(CURRENT_DATE, SEC_TO_TIME((HOUR(NOW())+1)*3600 + 60))
DO
    INSERT INTO hourly_avg_temperature (sensor_name, time, average_temperature)
    SELECT
        tr.sensor_name,
        DATE_FORMAT(tr.time, '%Y-%m-%d %H:00:00') AS hour,
        ROUND(AVG(tr.temperature), 2) AS avg_temp
    FROM temperature_readings tr
    WHERE tr.time >= DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 HOUR), '%Y-%m-%d %H:00:00')
      AND tr.time < DATE_FORMAT(NOW(), '%Y-%m-%d %H:00:00')
    GROUP BY tr.sensor_name, hour
    HAVING NOT EXISTS (
        SELECT 1
        FROM hourly_avg_temperature hat
        WHERE hat.sensor_name = tr.sensor_name
            AND hat.time = hour
);