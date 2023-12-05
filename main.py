# Reporting temperature reading from DS18B20 temperature sensor over 1 wire protocol
# Reading is ingested into Google Sheets
import os
from ds18b20_sensor import DS18B20TempSensor
from dotenv import load_dotenv

load_dotenv()

TEMP_SENSOR_NAME = os.getenv("TEMP_SENSOR_NAME")
TEMP_SENSOR_SHEET_ID = os.getenv("TEMP_SENSOR_SHEET_ID")


def main():
    sensor = DS18B20TempSensor(TEMP_SENSOR_NAME)
    print(sensor)


if __name__ == "__main__":
    main()
