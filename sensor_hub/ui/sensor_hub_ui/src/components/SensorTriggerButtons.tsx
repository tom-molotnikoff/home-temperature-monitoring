import { useState } from "react";
import Button from "@mui/material/Button";
import type { CSSProperties } from "@mui/material";
import DesktopRowMobileColumn from "../tools/DesktopRowMobileColumn";
import type {Sensor} from "../gen/aliases";
import { apiClient } from "../gen/client";

function SensorTriggerButtons({ sensors }: { sensors: Sensor[] }) {
  const [loadingSensor, setLoadingSensor] = useState<string | null>(null);

  const handleClick = async (sensor: string) => {
    setLoadingSensor(sensor);
    try {
      await apiClient.POST('/sensors/collect/{sensorName}', { params: { path: { sensorName: sensor } } });
    } finally {
      setLoadingSensor(null);
    }
  };

  return (
    <DesktopRowMobileColumn>
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
    </DesktopRowMobileColumn>
  );
}

const buttonStyle: CSSProperties = {
  width: 190,
};

export default SensorTriggerButtons;
