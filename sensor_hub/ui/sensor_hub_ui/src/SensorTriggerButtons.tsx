import { useState } from "react";

function SensorTriggerButtons({
  sensors,
  onButtonClick,
}: {
  sensors: string[];
  onButtonClick: (sensor: string) => Promise<void>;
}) {
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
    <div style={{ marginBottom: 16 }}>
      {sensors.map((sensor) => (
        <button
          key={sensor}
          style={{
            marginRight: 12,
            padding: "10px 28px",
            borderRadius: 6,
            background: loadingSensor === sensor ? "#6c63b5" : "#8884d8",
            color: "#fff",
            border: "none",
            fontSize: "1.1rem",
            fontWeight: 500,
            minWidth: 200,
            minHeight: 44,
            display: "inline-flex",
            alignItems: "center",
            justifyContent: "center",
            gap: 10, // use gap for spacing
            opacity: loadingSensor === sensor ? 0.7 : 1,
            cursor: loadingSensor === sensor ? "wait" : "pointer",
            position: "relative",
            boxShadow: "0 1px 4px rgba(0,0,0,0.04)",
            transition: "background 0.2s, opacity 0.2s",
          }}
          disabled={loadingSensor === sensor}
          onClick={() => handleClick(sensor)}
        >
          {/* Spinner and text in flex row, no marginRight */}
          {loadingSensor === sensor ? (
            <span
              style={{
                display: "inline-block",
                width: 20,
                height: 20,
                border: "3px solid #fff",
                borderTop: "3px solid #8884d8",
                borderRadius: "50%",
                animation: "spin 1s linear infinite",
                boxSizing: "border-box",
                verticalAlign: "middle",
              }}
            />
          ) : (
            <span
              style={{
                width: 20,
                height: 20,
                display: "inline-block",
                border: "3px solid transparent", // match spinner border
                borderRadius: "50%",
                boxSizing: "border-box",
                verticalAlign: "middle",
              }}
            />
          )}
          <span style={{ verticalAlign: "middle" }}>
            Take Reading: {sensor}
          </span>
          <style>
            {`
              @keyframes spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
              }
            `}
          </style>
        </button>
      ))}
    </div>
  );
}

export default SensorTriggerButtons;
