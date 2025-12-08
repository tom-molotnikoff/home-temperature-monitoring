CREATE TABLE sensor_health_history (
    id SERIAL PRIMARY KEY,
    sensor_id INT NOT NULL,
    health_status VARCHAR(20) NOT NULL,
    recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sensor_id) REFERENCES sensors(id)
);