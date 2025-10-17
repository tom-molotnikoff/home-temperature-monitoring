import type { CSSProperties } from "react";
import { Paper } from "@mui/material";

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
  let backgroundColor;
  if (variant === "primary" || variant === "tertiary") {
    backgroundColor = "background.default";
  } else if (variant === "secondary") {
    backgroundColor = "background.paper";
  }

  return (
    <Paper
      elevation={2}
      sx={{
        p: 2,
        border: 1,
        borderColor: "divider",
        borderRadius: 2,
        gap: 1,
        display: "flex",
        flexDirection: direction,
        backgroundColor,
        ...(changes || {}),
      }}
      data-testid={testid}
    >
      {children}
    </Paper>
  );
}

export default LayoutCard;
