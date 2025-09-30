import type { CSSProperties } from "react";

interface ColumnLayoutCardProps {
  children: React.ReactNode;
  changes?: CSSProperties;
  variant?: "primary" | "secondary" | "tertiary";
  testid?: string;
}

function ColumnLayoutCard({
  children,
  changes,
  variant = "primary",
  testid,
}: ColumnLayoutCardProps) {
  return (
    <div
      style={{
        ...(variant === "primary"
          ? columnLayoutCardPrimaryStyle
          : variant === "secondary"
          ? columnLayoutCardSecondaryStyle
          : columnLayoutCardTertiaryStyle),
        ...changes,
      }}
      data-testid={testid}
    >
      {children}
    </div>
  );
}

const columnLayoutCardPrimaryStyle: CSSProperties = {
  padding: 18,
  border: "1px solid #ccc",
  background: "#f0f0f0ff",
  borderRadius: 12,
  gap: 4,
  display: "flex",
  flexDirection: "column",
};

const columnLayoutCardSecondaryStyle: CSSProperties = {
  ...columnLayoutCardPrimaryStyle,
  background: "#ffffffeb",
};

const columnLayoutCardTertiaryStyle: CSSProperties = {
  ...columnLayoutCardPrimaryStyle,
  background: "#7abbde31",
};

export default ColumnLayoutCard;
