import { useEffect, useMemo, useState } from "react";
import "./App.css";
import CurrentTemperatures from "./CurrentTemperatures";
import type { TemperatureReading } from "./types";
import DateRangePicker from "./DateRangePicker";
import SensorTriggerButtons from "./SensorTriggerButtons";
import TemperatureGraph from "./TemperatureGraph";

const API_BASE = import.meta.env.VITE_API_BASE;
const WEBSOCKET_BASE = import.meta.env.VITE_WEBSOCKET_BASE;

function App() {
  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  const sevenDaysAgo = new Date();
  sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);

  const [useHourlyAverages, setUseHourlyAverages] = useState(true);

  const fetchReadings = async (
    start: string,
    end: string
  ): Promise<TemperatureReading[]> => {
    let response: Response;
    if (useHourlyAverages) {
      response = await fetch(
        `${API_BASE}/readings/hourly/between?start=${start}&end=${end}`
      );
    } else {
      response = await fetch(
        `${API_BASE}/readings/between?start=${start}&end=${end}`
      );
    }
    if (!response.ok) {
      throw new Error("Failed to fetch readings");
    }
    return response.json();
  };

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

  const sensors = useMemo(() => ["Downstairs", "Upstairs"], []);
  useEffect(() => {
    const ws = new WebSocket(`${WEBSOCKET_BASE}/ws/current-temperatures`);
    ws.onmessage = (event) => {
      if (!event.data || event.data === "null") return;
      const arr = JSON.parse(event.data);
      // Convert array to object keyed by sensor_name
      const obj: { [key: string]: TemperatureReading } = {};
      arr.forEach((reading: TemperatureReading) => {
        obj[String(reading.sensor_name)] = reading;
      });
      setCurrentReadings(obj);
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
  }, [startDate, endDate, useHourlyAverages]);

  return (
    <div
      style={{
        width: "100%",
        minHeight: 700,
        margin: "40px auto",
        padding: 24,
        background: "#fff",
        borderRadius: 16,
        boxShadow: "0 2px 16px rgba(0,0,0,0.07)",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        position: "relative",
      }}
    >
      <div
        style={{
          position: "absolute",
          top: 24,
          right: 24,
          display: "flex",
          alignItems: "center",
          gap: 8,
        }}
      >
        <label htmlFor="hourly-toggle" style={{ fontWeight: 500 }}>
          Hourly averages
        </label>
        <input
          id="hourly-toggle"
          type="checkbox"
          checked={useHourlyAverages}
          onChange={(e) => setUseHourlyAverages(e.target.checked)}
        />
      </div>
      <h1 style={{ fontSize: 32, fontWeight: 700, marginBottom: 24 }}>
        Temperature Sensor Dashboard
      </h1>
      <>
        <SensorTriggerButtons
          sensors={sensors}
          onButtonClick={triggerReading}
        />
        <CurrentTemperatures currentReadings={currentReadings} />
        <DateRangePicker
          startDate={startDate}
          endDate={endDate}
          onStartDateChange={(date) =>
            setStartDate(date ? date.toISOString().slice(0, 10) : "")
          }
          onEndDateChange={(date) =>
            setEndDate(date ? date.toISOString().slice(0, 10) : "")
          }
          invalidDate={invalidDate}
        />
        {Array.isArray(readings) && readings.length > 0 ? (
          <TemperatureGraph readings={readings} sensors={sensors} />
        ) : (
          <p>No readings found for the selected date range.</p>
        )}
      </>
    </div>
  );
}

export default App;
