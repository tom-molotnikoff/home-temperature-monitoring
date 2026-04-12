"""Mock Zigbee2MQTT devices that publish to the Sensor Hub embedded MQTT broker.

Simulates three devices:
  - living-room-sensor: temperature, humidity, battery, linkquality
  - front-door:         contact (binary), battery
  - office-plug:        power, energy, current, state (binary)

Each device publishes every PUBLISH_INTERVAL seconds (default 5) to topics
matching the zigbee2mqtt/<device-name> convention.
"""

import json
import os
import random
import signal
import sys
import time

import paho.mqtt.client as mqtt

BROKER_HOST = os.environ.get("MQTT_BROKER_HOST", "sensor-hub")
BROKER_PORT = int(os.environ.get("MQTT_BROKER_PORT", "1883"))
PUBLISH_INTERVAL = int(os.environ.get("PUBLISH_INTERVAL", "5"))

# Mutable state for drifting values
state = {
    "living-room-sensor": {"temperature": 21.0, "humidity": 45.0, "battery": 95},
    "front-door": {"contact": True, "battery": 88},
    "office-plug": {"energy": 1.2, "state": "ON"},
}


def drift(value, low, high, step=0.3):
    """Random-walk a value within bounds."""
    value += random.uniform(-step, step)
    return round(max(low, min(high, value)), 2)


def build_living_room():
    s = state["living-room-sensor"]
    s["temperature"] = drift(s["temperature"], 16.0, 28.0, 0.2)
    s["humidity"] = drift(s["humidity"], 30.0, 70.0, 0.5)
    s["battery"] = max(0, min(100, s["battery"] + random.choice([-1, 0, 0, 0, 0])))
    return {
        "temperature": s["temperature"],
        "humidity": s["humidity"],
        "battery": s["battery"],
        "linkquality": random.randint(40, 255),
    }


def build_front_door():
    s = state["front-door"]
    # Toggle contact occasionally (~10% chance)
    if random.random() < 0.10:
        s["contact"] = not s["contact"]
    s["battery"] = max(0, min(100, s["battery"] + random.choice([-1, 0, 0, 0, 0])))
    return {
        "contact": s["contact"],
        "battery": s["battery"],
    }


def build_office_plug():
    s = state["office-plug"]
    # Toggle on/off occasionally (~5% chance)
    if random.random() < 0.05:
        s["state"] = "OFF" if s["state"] == "ON" else "ON"
    if s["state"] == "ON":
        power = round(random.uniform(40.0, 120.0), 1)
        current = round(power / 230.0, 3)
        s["energy"] = round(s["energy"] + power * PUBLISH_INTERVAL / 3_600_000, 4)
    else:
        power = 0.0
        current = 0.0
    return {
        "power": power,
        "energy": s["energy"],
        "current": current,
        "state": s["state"],
    }


DEVICES = {
    "living-room-sensor": build_living_room,
    "front-door": build_front_door,
    "office-plug": build_office_plug,
}


def on_connect(client, userdata, flags, rc, properties=None):
    if rc == 0:
        print(f"Connected to MQTT broker at {BROKER_HOST}:{BROKER_PORT}", flush=True)
    else:
        print(f"Connection failed with code {rc}", flush=True)


def main():
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2, client_id="mock-mqtt-sensor")
    client.on_connect = on_connect

    print(f"Connecting to {BROKER_HOST}:{BROKER_PORT}...", flush=True)

    # Retry connection until broker is available
    while True:
        try:
            client.connect(BROKER_HOST, BROKER_PORT, keepalive=60)
            break
        except (ConnectionRefusedError, OSError) as e:
            print(f"Broker not ready ({e}), retrying in 3s...", flush=True)
            time.sleep(3)

    client.loop_start()

    def shutdown(sig, frame):
        print("Shutting down...", flush=True)
        client.loop_stop()
        client.disconnect()
        sys.exit(0)

    signal.signal(signal.SIGTERM, shutdown)
    signal.signal(signal.SIGINT, shutdown)

    device_names = list(DEVICES.keys())
    stagger = PUBLISH_INTERVAL / len(device_names)

    print(f"Publishing {len(device_names)} devices every {PUBLISH_INTERVAL}s", flush=True)

    while True:
        for i, (name, builder) in enumerate(DEVICES.items()):
            topic = f"zigbee2mqtt/{name}"
            payload = json.dumps(builder())
            client.publish(topic, payload, qos=0)
            print(f"  → {topic}: {payload}", flush=True)
            if i < len(device_names) - 1:
                time.sleep(stagger)
        remaining = PUBLISH_INTERVAL - stagger * (len(device_names) - 1)
        time.sleep(max(0.1, remaining))


if __name__ == "__main__":
    main()
