import type { CSSProperties } from "react";

interface TypographyProps {
  children: React.ReactNode;
  testid?: string;
  changes?: CSSProperties;
}

export function TypographyH1({ children, testid, changes }: TypographyProps) {
  return (
    <h1 data-testid={testid} style={changes}>
      {children}
    </h1>
  );
}
export function TypographyH2({ children, testid, changes }: TypographyProps) {
  return (
    <h2 data-testid={testid} style={{ marginBottom: 8, ...changes }}>
      {children}
    </h2>
  );
}
export function TypographyH3({ children, testid, changes }: TypographyProps) {
  return (
    <h3 data-testid={testid} style={changes}>
      {children}
    </h3>
  );
}

export function TypographySecondaryHeading({
  children,
  testid,
  changes,
}: TypographyProps) {
  return (
    <div
      style={{
        fontSize: 28,
        fontWeight: 700,
        color: "#1976d2",
        ...changes,
      }}
      data-testid={testid}
    >
      {children}
    </div>
  );
}

export function TypographyMutedText({
  children,
  testid,
  changes,
}: TypographyProps) {
  return (
    <div
      style={{ fontSize: 13, color: "#666", marginTop: 8, ...changes }}
      data-testid={testid}
    >
      {children}
    </div>
  );
}
