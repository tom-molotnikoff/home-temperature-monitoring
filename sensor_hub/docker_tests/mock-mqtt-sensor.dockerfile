FROM python:3.11-alpine

WORKDIR /app

RUN pip install --no-cache-dir paho-mqtt>=2.0

COPY mock-mqtt-sensor.py .

CMD ["python", "mock-mqtt-sensor.py"]
