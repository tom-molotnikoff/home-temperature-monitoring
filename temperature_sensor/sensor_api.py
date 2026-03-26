# Reporting temperature reading from DS18B20 temperature sensor over 1 wire protocol
import os
import socket
from ds18b20_sensor import DS18B20TempSensor
from dotenv import load_dotenv
from flask import Flask, request, jsonify, abort

load_dotenv()


def _init_telemetry():
    """Initialise OpenTelemetry tracing if OTEL_EXPORTER_OTLP_ENDPOINT is set."""
    endpoint = os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT")
    if not endpoint:
        return
    try:
        from opentelemetry import trace
        from opentelemetry.sdk.trace import TracerProvider
        from opentelemetry.sdk.trace.export import BatchSpanProcessor
        from opentelemetry.sdk.resources import Resource
        from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

        resource = Resource.create({
            "service.name": os.environ.get("OTEL_SERVICE_NAME", socket.gethostname())
        })
        provider = TracerProvider(resource=resource)
        exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
        provider.add_span_processor(BatchSpanProcessor(exporter))
        trace.set_tracer_provider(provider)
    except ImportError:
        pass


_init_telemetry()

app = Flask(__name__)

try:
    from opentelemetry.instrumentation.flask import FlaskInstrumentor
    FlaskInstrumentor().instrument_app(app)
except ImportError:
    pass

SERVICE_ACCOUNT_KEY_PATH = os.getenv("SERVICE_ACCOUNT_KEY_PATH")

@app.get("/temperature")
def get_temperature():
    try:
        sensor = DS18B20TempSensor()
        return jsonify(sensor.collect_data())
    except Exception as e:
        abort(500)

@app.errorhandler(500)
def internal_error(error):
    return {"message": "couldn't take a reading"},500
