DELETE FROM sensor_health_history
WHERE id IN (
    SELECT id
    FROM (
        SELECT
            id,
            health_status,
            LAG(health_status) OVER (
                PARTITION BY sensor_id
                ORDER BY datetime(recorded_at), id
            ) AS previous_health_status
        FROM sensor_health_history
    ) dedup_candidates
    WHERE previous_health_status = health_status
);
