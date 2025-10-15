import type { CSSProperties } from "react";

interface LayoutCardProps {
  children: React.ReactNode;
  changes?: CSSProperties;
  variant?: "primary" | "secondary" | "tertiary";
  testid?: string;
  direction?: "row" | "column";
}

function LayoutCard({
  children,
  changes,
  variant = "primary",
  testid,
  direction = "column",
}: LayoutCardProps) {
  return (
    <div
      style={{
        ...(variant === "primary"
          ? layoutCardPrimaryStyle
          : variant === "secondary"
          ? layoutCardSecondaryStyle
          : layoutCardTertiaryStyle),
        flexDirection: direction === "column" ? "column" : "row",
        ...changes,
      }}
      data-testid={testid}
    >
      {children}
    </div>
  );
}

const layoutCardPrimaryStyle: CSSProperties = {
  padding: 18,
  border: "1px solid #ccc",
  background: "#f0f0f0ff",
  borderRadius: 12,
  gap: 4,
  display: "flex",
  flexDirection: "column",
};

const layoutCardSecondaryStyle: CSSProperties = {
  ...layoutCardPrimaryStyle,
  background: "#ffffffeb",
};

const layoutCardTertiaryStyle: CSSProperties = {
  ...layoutCardPrimaryStyle,
  background: "#7abbde31",
};

export default LayoutCard;
