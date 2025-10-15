import { useState } from "react";
import Button from "@mui/material/Button";
import { API_BASE } from "../environment/Environment";
import type { CSSProperties } from "@mui/material";
import DeskopRowMobileColumn from "../tools/DesktopRowMobileColumn";
import type {Sensor} from "../types/types.ts";

function SensorTriggerButtons({ sensors }: { sensors: Sensor[] }) {
  const triggerReading = async (sensor: string) => {
    const response = await fetch(`${API_BASE}/sensors/collect/${sensor}`, {
      method: "POST",
    });
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
          key={sensor.name}
          variant="contained"
          color="primary"
          onClick={() => handleClick(sensor.name)}
          disabled={loadingSensor === sensor.name}
          style={buttonStyle}
        >
          {`Trigger ${sensor.name}`}
        </Button>
      ))}
    </DeskopRowMobileColumn>
  );
}

const buttonStyle: CSSProperties = {
  width: 190,
};

export default SensorTriggerButtons;
