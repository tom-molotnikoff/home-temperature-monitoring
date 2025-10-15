import type { CSSProperties } from "react";
import ColumnLayoutCardProps from "./LayoutCard.tsx";

interface PageContainerProps {
  children: React.ReactNode;
  titleText: string;
}

function PageContainer({ children, titleText }: PageContainerProps) {
  return (
    <div style={outerContainerStyle}>
      <h1 style={titleStyle}>{titleText}</h1>
      <ColumnLayoutCardProps changes={layoutCardStyleChanges}>
        {children}
      </ColumnLayoutCardProps>
    </div>
  );
}

const titleStyle: CSSProperties = {
  fontSize: 32,
  fontWeight: 700,
  marginBottom: 16,
};

const layoutCardStyleChanges: CSSProperties = {
  gap: 20,
  padding: 20,
  alignItems: "center",
  border: "none",
};

const outerContainerStyle: CSSProperties = {
  padding: 24,
  display: "block",
  textAlign: "center",
  background: "#f0f0f0ff",
  width: "100%",
};

export default PageContainer;
