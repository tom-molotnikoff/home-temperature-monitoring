import { useIsMobile } from "../hooks/useMobile";
import type { CSSProperties } from "@mui/material";
import { useCurrentTemperatures } from "../hooks/useCurrentTemperatures";
import CurrentTemperatureReadingCard from "./CurrentTemperatureReadingCard";

function CurrentTemperatures() {
  const currentTemperatures = useCurrentTemperatures();

  const sensorNames = Object.keys(currentTemperatures).sort((a, b) =>
    a.localeCompare(b)
  );

  const isMobile = useIsMobile();

  return (
    <div style={outerContainerStyle}>
      <h2>Current Temperatures</h2>
      <div style={isMobile ? mobileCardLayoutStyle : desktopCardLayoutStyle}>
        {sensorNames.map((sensor) => {
          const readingObj = currentTemperatures[sensor];
          return (
            <CurrentTemperatureReadingCard key={sensor} reading={readingObj} />
          );
        })}
        {sensorNames.length === 0 && (
          <div style={loadingStyle}>Checking Temperatures...</div>
        )}
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

const loadingStyle: CSSProperties = {
  background: "#fffbe6",
  borderRadius: 8,
  padding: "18px 24px",
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  color: "#888",
  minWidth: 395,
  minHeight: 142,
};

export default CurrentTemperatures;
