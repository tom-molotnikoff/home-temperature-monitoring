interface CenteredTextProps {
  children: React.ReactNode;
  testid?: string;
  changes?: React.CSSProperties;
}

function CenteredFlex({ children, testid, changes }: CenteredTextProps) {
  return (
    <div
      style={{
        textAlign: "center",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: 16,
        ...changes,
      }}
      data-testid={testid}
    >
      {children}
    </div>
  );
}

export default CenteredFlex;
