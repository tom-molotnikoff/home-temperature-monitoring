import type { CSSProperties } from "react";

interface LoadingContentBlockProps {
  children?: React.ReactNode;
  testid?: string;
  changes?: CSSProperties;
}

function LoadingContentBlock({
  children = "Loading...",
  testid,
  changes,
}: LoadingContentBlockProps) {
  return (
    <div
      style={{ ...loadingContentBlockStyle, ...changes }}
      data-testid={testid}
    >
      {children}
    </div>
  );
}

const loadingContentBlockStyle: CSSProperties = {
  background: "#fffbe6",
  borderRadius: 8,
  padding: "18px 24px",
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  color: "#888",
};

export default LoadingContentBlock;
