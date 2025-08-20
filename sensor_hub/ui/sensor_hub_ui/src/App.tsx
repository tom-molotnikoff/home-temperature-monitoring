import { useEffect, useState } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  ResponsiveContainer,
  Legend,
} from "recharts";

type TemperatureReading = {
  sensor_name: string;
  reading: {
    temperature: number;
    time: string;
  };
};

type ChartEntry = {
  time: string;
  Upstairs: number | null;
  Downstairs: number | null;
};

const API_BASE = "http://localhost:8080";

const fetchReadings = async (
  start: string,
  end: string
): Promise<TemperatureReading[]> => {
  const response = await fetch(
    `${API_BASE}/readings/between?start=${start}&end=${end}`
  );
  if (!response.ok) {
    throw new Error("Failed to fetch readings");
  }
  return response.json();
};

function App() {
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  const sevenDaysAgo = new Date();
  sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);

  const [readings, setReadings] = useState<TemperatureReading[]>([]);
  const [startDate, setStartDate] = useState(
    sevenDaysAgo.toISOString().slice(0, 10)
  );
  const [endDate, setEndDate] = useState(tomorrow.toISOString().slice(0, 10));
  const [invalidDate, setInvalidDate] = useState(false);
  const [currentReadings, setCurrentReadings] = useState<{
    [sensor: string]: TemperatureReading;
  }>({});

  const triggerReading = async (sensor: string) => {
    const response = await fetch(`${API_BASE}/sensors/temperature/${sensor}`);
    if (!response.ok) {
      throw new Error(`Failed to trigger reading for ${sensor}`);
    }
  };

  const sensors: Array<"Upstairs" | "Downstairs"> = ["Upstairs", "Downstairs"];

  // Extract unique times from readings, replacing spaces with 'T' for ISO format
  // This is necessary because the readings may not be in chronological order or may have gaps.
  const times = Array.from(
    new Set((readings ?? []).map((r) => r.reading.time.replace(" ", "T")))
  );

  useEffect(() => {
    const ws = new WebSocket("ws://localhost:8080/ws/current-temperatures");
    ws.onmessage = (event) => {
      if (!event.data || event.data == "null") return;
      const data = JSON.parse(event.data);
      setCurrentReadings(data); // update your state
    };
    ws.onerror = (err) => {
      console.error("WebSocket error:", err);
    };
    return () => ws.close();
  }, []);

  useEffect(() => {
    if (!startDate || !endDate) {
      setInvalidDate(true);
      return;
    }

    // Validate date range
    const start = new Date(startDate);
    const end = new Date(endDate);
    if (isNaN(start.getTime()) || isNaN(end.getTime())) {
      setInvalidDate(true);
      return;
    }
    if (start >= end) {
      setInvalidDate(true);
      return;
    }

    // Fetch initial readings on mount and every time start or end date changes

    setInvalidDate(false);
    fetchReadings(startDate, endDate)
      .then((data) => {
        setReadings(data);
      })
      .catch((error) => {
        console.error("Error fetching readings:", error);
      });
  }, [startDate, endDate]);

  // Build merged data with nulls for missing sensors - A.I nonsense here:
  // Recharts needs a single array of objects where each object has all the values for each line (sensor) at a given time.
  // Your raw readings are not grouped by time and sensor, so you must "merge" them into this format.
  // If you skip this, Recharts cannot plot multiple lines correctly, and lines will be disconnected or missing.
  const mergedData: ChartEntry[] = times.map((time) => {
    const entry: ChartEntry = {
      time,
      Upstairs: null,
      Downstairs: null,
    };
    sensors.forEach((sensor) => {
      const found = readings.find(
        (r) =>
          r.sensor_name === sensor && r.reading.time.replace(" ", "T") === time
      );
      entry[sensor] = found ? found.reading.temperature : null;
    });
    return entry;
  });
  return (
    <div
      style={{
        width: "90vw",
        height: "80vh",
        margin: "40px auto",
        padding: 24,
        background: "#fff",
        borderRadius: 16,
        boxShadow: "0 2px 16px rgba(0,0,0,0.07)",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <h1 style={{ fontSize: 32, fontWeight: 700, marginBottom: 24 }}>
        Temperature Sensor Dashboard
      </h1>
      <>
        <div style={{ marginBottom: 16 }}>
          {sensors.map((sensor) => (
            <button
              key={sensor}
              style={{
                marginRight: 8,
                padding: "4px 12px",
                borderRadius: 4,
                background: "#8884d8",
                color: "#fff",
                border: "none",
              }}
              onClick={() => triggerReading(sensor)}
            >
              Take Reading: {sensor}
            </button>
          ))}
        </div>

        <div style={{ marginBottom: 16 }}>
          <h3>Current Temperatures</h3>
          <ul>
            {Object.values(currentReadings).map((readingObj, idx) => (
              <li key={idx}>
                {readingObj.sensor_name}:{" "}
                {readingObj.reading?.temperature ?? "N/A"}°C at{" "}
                {readingObj.reading?.time
                  ? new Date(
                      readingObj.reading.time.replace(" ", "T")
                    ).toLocaleTimeString()
                  : "Unknown time"}
              </li>
            ))}
            {Object.keys(currentReadings).length === 0 && (
              <li>No readings taken yet.</li>
            )}
          </ul>
        </div>
        <div style={{ marginBottom: 16 }}>
          <input
            type="date"
            value={startDate}
            style={{ marginRight: 8, padding: 4 }}
            onChange={(e) => setStartDate(e.target.value)}
          />
          <input
            type="date"
            value={endDate}
            style={{ marginRight: 8, padding: 4 }}
            onChange={(e) => setEndDate(e.target.value)}
          />
          {invalidDate && (
            <p style={{ color: "red" }}>Start date must be before end date.</p>
          )}
        </div>
        {!Array.isArray(readings) || readings.length === 0 ? (
          <p>No readings found for the selected date range.</p>
        ) : (
          <ResponsiveContainer width="100%" height="70%">
            <LineChart data={mergedData}>
              <CartesianGrid stroke="#eee" />
              <XAxis
                dataKey="time"
                tickFormatter={(t) => new Date(t).toLocaleTimeString()}
              />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line
                type="monotone"
                dataKey="Upstairs"
                stroke="#8884d8"
                strokeWidth={2}
                dot={false}
                connectNulls={true}
              />
              <Line
                type="monotone"
                dataKey="Downstairs"
                strokeWidth={2}
                dot={false}
                stroke="#82ca9d"
                connectNulls={true}
              />
            </LineChart>
          </ResponsiveContainer>
        )}
      </>
    </div>
  );
}

export default App;
