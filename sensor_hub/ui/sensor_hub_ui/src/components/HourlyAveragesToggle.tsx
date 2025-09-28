import type { CSSProperties } from "react";

interface HourlyAveragesToggleProps {
  useHourlyAverages: boolean;
  setUseHourlyAverages: (value: boolean) => void;
}

function HourlyAveragesToggle({
  useHourlyAverages,
  setUseHourlyAverages,
}: HourlyAveragesToggleProps) {
  return (
    <>
      <label htmlFor="hourly-toggle" style={optionsTextStyle}>
        Hourly averages
      </label>
      <input
        id="hourly-toggle"
        type="checkbox"
        checked={useHourlyAverages}
        onChange={(e) => setUseHourlyAverages(e.target.checked)}
      />
    </>
  );
}

const optionsTextStyle: CSSProperties = { fontWeight: 500 };

export default HourlyAveragesToggle;
