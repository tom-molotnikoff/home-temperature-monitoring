"""
Sensor abstract
"""

from w1thermsensor import W1ThermSensor
from datetime import datetime
import json


class DS18B20TempSensor:
    """
    DS18B20 interface class
    """

    reading = {
            "time": 0,
            "temperature": 0
        }

    def __init__(self, name):
        self.sensor = W1ThermSensor()

    def collect_data(self):
        """
        Collect data from the sensor and process into an object with the sensor name, time and temperature in C
        """
        temperature_in_celsius = self.sensor.get_temperature()
        now = datetime.now()
        now_str = now.strftime('%Y-%m-%d %H:%M:%S')
        self.reading = {
            "time": now_str,
            "temperature": temperature_in_celsius
        }
        return self.reading

    def __str__(self):
        return json.dumps(self.collect_data())

    def __repr__(self):
        return json.dumps(self.reading)
