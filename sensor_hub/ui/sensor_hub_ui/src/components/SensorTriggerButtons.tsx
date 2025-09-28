import { useState } from "react";
import Button from "@mui/material/Button";
import { useIsMobile } from "../hooks/useMobile";
import { API_BASE } from "../environment/Environment";
import type { CSSProperties } from "@mui/material";

function SensorTriggerButtons({ sensors }: { sensors: string[] }) {
  const isMobile = useIsMobile();

  const triggerReading = async (sensor: string) => {
    const response = await fetch(`${API_BASE}/sensors/temperature/${sensor}`);
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
    <div
      style={
        isMobile ? mobileButtonContainerStyle : desktopButtonContainerStyle
      }
    >
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
    </div>
  );
}

const desktopButtonContainerStyle: CSSProperties = {
  display: "flex",
  marginBottom: 16,
  gap: 16,
};

const mobileButtonContainerStyle: CSSProperties = {
  ...desktopButtonContainerStyle,
  flexDirection: "column",
  gap: 16,
};

const buttonStyle: CSSProperties = {
  width: 190,
};

export default SensorTriggerButtons;
