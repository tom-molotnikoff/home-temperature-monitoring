import type { CSSProperties } from "react";

interface PageContainerProps {
  children: React.ReactNode;
  titleText: string;
}

function PageContainer({ children, titleText }: PageContainerProps) {
  return (
    <div style={outerContainerStyle}>
      <div style={shadowedCardStyle}>
        <h1 style={titleStyle}>{titleText}</h1>
        {children}
      </div>
    </div>
  );
}

const titleStyle: CSSProperties = {
  fontSize: 32,
  fontWeight: 700,
};

const outerContainerStyle: CSSProperties = {
  padding: "40px 0",
};

const shadowedCardStyle: CSSProperties = {
  width: "90%",
  margin: "auto",
  padding: 12,
  gap: 16,
  boxShadow: "0 2px 16px rgba(0,0,0,0.07)",
  display: "flex",
  flexDirection: "column",
  borderRadius: 16,
  alignItems: "center",
};

export default PageContainer;
