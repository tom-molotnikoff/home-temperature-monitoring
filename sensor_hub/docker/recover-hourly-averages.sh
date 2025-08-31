#!/bin/bash
# this script recovers hourly averages for all previous hours.
# this can be needed if the application was stopped for a period of time
# and the hourly averages were not calculated during that time.

# requires the container name as the first argument
MYSQL_ROOT_PASSWORD="password"

if [ -z "$1" ]; then
  echo "Usage: $0 <container_name>"
  exit 1
fi
docker exec -it $1 mysql -u root -p"$MYSQL_ROOT_PASSWORD" -D sensor_database -e "INSERT INTO hourly_avg_temperature (sensor_name, time, average_temperature) SELECT tr.sensor_name, DATE_FORMAT(tr.time, '%Y-%m-%d %H:00:00') AS hour, AVG(tr.temperature) AS avg_temp FROM temperature_readings tr GROUP BY tr.sensor_name, hour HAVING NOT EXISTS (SELECT 1 FROM hourly_avg_temperature hat WHERE hat.sensor_name = tr.sensor_name AND hat.time = hour);"