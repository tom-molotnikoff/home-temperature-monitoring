import type { CSSProperties } from "react";
import type { TemperatureReading } from "../types/types";

interface CurrentTemperatureReadingCardProps {
  reading: TemperatureReading;
}

function CurrentTemperatureReadingCard({
  reading,
}: CurrentTemperatureReadingCardProps) {
  return (
    <div style={shadowedCardStyle}>
      <div style={sensorNameStyle}>{reading.sensor_name}</div>
      <div style={temperatureStyle}>
        {reading.reading?.temperature ?? "N/A"}°C
      </div>
      <div style={timeStyle}>
        {reading.reading?.time
          ? new Date(
              reading.reading.time.replace(" ", "T")
            ).toLocaleTimeString()
          : "Unknown time"}
      </div>
    </div>
  );
}

const shadowedCardStyle: CSSProperties = {
  background: "#f7f7fa",
  borderRadius: 8,
  boxShadow: "0 2px 8px rgba(0,0,0,0.07)",
  padding: "18px 24px",
  minWidth: 190,
  display: "flex",
  flexDirection: "column",
  alignItems: "center",
};

const sensorNameStyle: CSSProperties = {
  fontWeight: 600,
  fontSize: 18,
  marginBottom: 8,
};

const temperatureStyle: CSSProperties = {
  fontSize: 28,
  fontWeight: 700,
  color: "#1976d2",
};

const timeStyle: CSSProperties = { fontSize: 13, color: "#666", marginTop: 8 };

export default CurrentTemperatureReadingCard;
