import { useState } from "react";
import Button from "@mui/material/Button";
import { useIsMobile } from "../hooks/useMobile";

function SensorTriggerButtons({
  sensors,
  onButtonClick,
}: {
  sensors: string[];
  onButtonClick: (sensor: string) => Promise<void>;
}) {
  const isMobile = useIsMobile();

  const [loadingSensor, setLoadingSensor] = useState<string | null>(null);

  const handleClick = async (sensor: string) => {
    setLoadingSensor(sensor);
    try {
      await onButtonClick(sensor);
    } finally {
      setLoadingSensor(null);
    }
  };

  return (
    <div
      style={
        isMobile
          ? { display: "flex", flexDirection: "column", gap: "16px" }
          : { display: "flex", marginBottom: 16 }
      }
    >
      {sensors.map((sensor) => (
        <Button
          key={sensor}
          variant="contained"
          color="primary"
          onClick={() => handleClick(sensor)}
          disabled={loadingSensor === sensor}
          style={{ marginRight: 8 }}
        >
          {`Trigger ${sensor}`}
        </Button>
      ))}
    </div>
  );
}

export default SensorTriggerButtons;
