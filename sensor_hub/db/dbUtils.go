package database

import (
	"database/sql"
	"example/sensorHub/types"
	"fmt"
)

func scanDbReading(rows *sql.Rows) ([]types.DbTempReading, error) {
	var readings []types.DbTempReading
	for rows.Next() {
		var reading types.DbTempReading
		err := rows.Scan(&reading.Id, &reading.SensorName, &reading.Time, &reading.Temperature)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		readings = append(readings, reading)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return readings, nil
}
