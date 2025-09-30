import type { CSSProperties } from "react";

interface ToggleWithLabelProps {
  label: string;
  isChecked: boolean;
  id: string;
  testid?: string;
  onToggle: (value: boolean) => void;
}

function ToggleWithLabel({
  label,
  id,
  testid,
  isChecked,
  onToggle,
}: ToggleWithLabelProps) {
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
      <label htmlFor={id} style={optionsTextStyle}>
        {label}
      </label>
      <input
        id={id}
        type="checkbox"
        checked={isChecked}
        data-testid={testid}
        onChange={(e) => onToggle(e.target.checked)}
      />
    </div>
  );
}

const optionsTextStyle: CSSProperties = { fontWeight: 500 };

export default ToggleWithLabel;
