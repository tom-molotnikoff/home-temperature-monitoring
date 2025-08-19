# Reporting temperature reading from DS18B20 temperature sensor over 1 wire protocol
# Reading is ingested into Google Sheets
import os
from ds18b20_sensor import DS18B20TempSensor
from dotenv import load_dotenv
from flask import Flask, request, jsonify, abort

app = Flask(__name__)

load_dotenv()

TEMP_SENSOR_NAME = os.getenv("TEMP_SENSOR_NAME")
TEMP_SENSOR_SHEET_ID = os.getenv("TEMP_SENSOR_SHEET_ID")
SERVICE_ACCOUNT_KEY_PATH = os.getenv("SERVICE_ACCOUNT_KEY_PATH")
SHEET_NAME = os.getenv("SHEET_NAME")

@app.get("/temperature")
def get_temperature():
    try:
        sensor = DS18B20TempSensor(TEMP_SENSOR_NAME)
        return jsonify(sensor.collect_data_with_name())
    except Exception as e:
        abort(500)

@app.errorhandler(500)
def internal_error(error):
    return {"message": "couldn't take a reading"},500
