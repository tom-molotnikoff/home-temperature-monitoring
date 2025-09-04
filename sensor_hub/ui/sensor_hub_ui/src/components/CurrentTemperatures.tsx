import { useIsMobile } from "../hooks/useMobile";
import type { TemperatureReading } from "../types/types";

function CurrentTemperatures({
  currentReadings,
}: {
  currentReadings: { [sensor: string]: TemperatureReading };
}) {
  const sensorNames = Object.keys(currentReadings).sort((a, b) =>
    a.localeCompare(b)
  );

  const isMobile = useIsMobile();

  return (
    <div style={{ marginBottom: 16 }}>
      <h3 style={{ marginBottom: 16 }}>Current Temperatures</h3>
      <div
        style={
          isMobile
            ? { display: "flex", flexDirection: "column", gap: "16px" }
            : {
                display: "flex",
                gap: "16px",
                flexWrap: "wrap",
                justifyContent: "flex-start",
              }
        }
      >
        {sensorNames.map((sensor) => {
          const readingObj = currentReadings[sensor];
          return (
            <div
              key={sensor}
              style={{
                background: "#f7f7fa",
                borderRadius: 8,
                boxShadow: "0 2px 8px rgba(0,0,0,0.07)",
                padding: "18px 24px",
                minWidth: 180,
                display: "flex",
                flexDirection: "column",
                alignItems: "flex-start",
              }}
            >
              <div style={{ fontWeight: 600, fontSize: 18, marginBottom: 8 }}>
                {readingObj.sensor_name}
              </div>
              <div style={{ fontSize: 28, fontWeight: 700, color: "#1976d2" }}>
                {readingObj.reading?.temperature ?? "N/A"}°C
              </div>
              <div style={{ fontSize: 13, color: "#666", marginTop: 8 }}>
                {readingObj.reading?.time
                  ? new Date(
                      readingObj.reading.time.replace(" ", "T")
                    ).toLocaleTimeString()
                  : "Unknown time"}
              </div>
            </div>
          );
        })}
        {sensorNames.length === 0 && (
          <div
            style={{
              background: "#fffbe6",
              borderRadius: 8,
              padding: "18px 24px",
              color: "#888",
              minWidth: 180,
            }}
          >
            Loading...
          </div>
        )}
      </div>
    </div>
  );
}

export default CurrentTemperatures;
