import { useIsMobile } from "../hooks/useMobile";
import type { CSSProperties } from "@mui/material";
import { useCurrentTemperatures } from "../hooks/useCurrentTemperatures";

function CurrentTemperatures() {
  const currentTemperatures = useCurrentTemperatures();

  const sensorNames = Object.keys(currentTemperatures).sort((a, b) =>
    a.localeCompare(b)
  );

  const isMobile = useIsMobile();

  return (
    <div style={outerContainerStyle}>
      <h3>Current Temperatures</h3>
      <div style={isMobile ? mobileCardLayoutStyle : desktopCardLayoutStyle}>
        {sensorNames.map((sensor) => {
          const readingObj = currentTemperatures[sensor];
          return (
            <div key={sensor} style={shadowedCardStyle}>
              <div style={sensorNameStyle}>{readingObj.sensor_name}</div>
              <div style={temperatureStyle}>
                {readingObj.reading?.temperature ?? "N/A"}°C
              </div>
              <div style={timeStyle}>
                {readingObj.reading?.time
                  ? new Date(
                      readingObj.reading.time.replace(" ", "T")
                    ).toLocaleTimeString()
                  : "Unknown time"}
              </div>
            </div>
          );
        })}
        {sensorNames.length === 0 && <div style={loadingStyle}>Loading...</div>}
      </div>
    </div>
  );
}

const outerContainerStyle: CSSProperties = {
  marginBottom: 16,
  textAlign: "center",
  gap: 16,
};

const desktopCardLayoutStyle: CSSProperties = {
  display: "flex",
  gap: 16,
};

const mobileCardLayoutStyle: CSSProperties = {
  ...desktopCardLayoutStyle,
  flexDirection: "column",
};

const shadowedCardStyle: CSSProperties = {
  background: "#f7f7fa",
  borderRadius: 8,
  boxShadow: "0 2px 8px rgba(0,0,0,0.07)",
  padding: "18px 24px",
  minWidth: 180,
  display: "flex",
  flexDirection: "column",
  alignItems: "flex-start",
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

const loadingStyle: CSSProperties = {
  background: "#fffbe6",
  borderRadius: 8,
  padding: "18px 24px",
  color: "#888",
  minWidth: 180,
};

export default CurrentTemperatures;
