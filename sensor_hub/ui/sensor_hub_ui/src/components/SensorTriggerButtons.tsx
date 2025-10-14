import { useState } from "react";
import Button from "@mui/material/Button";
import { API_BASE } from "../environment/Environment";
import type { CSSProperties } from "@mui/material";
import DeskopRowMobileColumn from "../tools/DesktopRowMobileColumn";

function SensorTriggerButtons({ sensors }: { sensors: string[] }) {
  const triggerReading = async (sensor: string) => {
    const response = await fetch(`${API_BASE}/temperature/sensors/collect/${sensor}`);
    if (!response.ok) {
      throw new Error(`Failed to trigger reading for ${sensor}`);
    }
  };

  const [loadingSensor, setLoadingSensor] = useState<string | null>(null);

  const handleClick = async (sensor: string) => {
    setLoadingSensor(sensor);
    try {
      await triggerReading(sensor);
    } finally {
      setLoadingSensor(null);
    }
  };

  return (
    <DeskopRowMobileColumn>
      {sensors.map((sensor) => (
        <Button
          key={sensor}
          variant="contained"
          color="primary"
          onClick={() => handleClick(sensor)}
          disabled={loadingSensor === sensor}
          style={buttonStyle}
        >
          {`Trigger ${sensor}`}
        </Button>
      ))}
    </DeskopRowMobileColumn>
  );
}

const buttonStyle: CSSProperties = {
  width: 190,
};

export default SensorTriggerButtons;
