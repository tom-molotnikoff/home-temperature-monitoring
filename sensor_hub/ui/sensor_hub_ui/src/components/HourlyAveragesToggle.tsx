import ToggleWithLabel from "../tools/ToggleWithLabel";

interface HourlyAveragesToggleProps {
  useHourlyAverages: boolean;
  setUseHourlyAverages: (value: boolean) => void;
}

function HourlyAveragesToggle({
  useHourlyAverages,
  setUseHourlyAverages,
}: HourlyAveragesToggleProps) {
  return (
    <ToggleWithLabel
      label="Hourly averages"
      id="hourly-toggle"
      isChecked={useHourlyAverages}
      onToggle={setUseHourlyAverages}
      testid="hourly-averages-toggle"
    />
  );
}

export default HourlyAveragesToggle;
