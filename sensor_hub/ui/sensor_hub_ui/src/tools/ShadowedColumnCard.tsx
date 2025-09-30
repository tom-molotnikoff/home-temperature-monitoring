import type { CSSProperties } from "react";

interface ShadowedColumnCardProps {
  children: React.ReactNode;
  changes?: CSSProperties;
  variant?: "primary" | "secondary";
  testid?: string;
}

function ShadowedColumnCard({
  children,
  changes,
  variant = "primary",
  testid,
}: ShadowedColumnCardProps) {
  return (
    <div
      style={{
        ...(variant === "primary"
          ? shadowedColumnCardPrimaryStyle
          : shadowedColumnCardSecondaryStyle),
        ...changes,
      }}
      data-testid={testid}
    >
      {children}
    </div>
  );
}

const shadowedColumnCardPrimaryStyle: CSSProperties = {
  padding: 8,
  boxShadow: "0 5px 4px rgba(0,0,0,0.07)",
  background: "#fafafaff",
  borderRadius: 12,
  gap: 4,
  display: "flex",
  flexDirection: "column",
};

const shadowedColumnCardSecondaryStyle: CSSProperties = {
  ...shadowedColumnCardPrimaryStyle,
  background: "#f4f4f5ff",
};

export default ShadowedColumnCard;
