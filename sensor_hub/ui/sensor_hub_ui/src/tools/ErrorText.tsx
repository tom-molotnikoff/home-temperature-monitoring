import type { CSSProperties } from "react";

interface ErrorTextProps {
  message: string;
  testid?: string;
  changes?: CSSProperties;
}

function ErrorText({ message, testid, changes }: ErrorTextProps) {
  return (
    <div style={{ ...errorTextStyle, ...changes }} data-testid={testid}>
      {message}
    </div>
  );
}

const errorTextStyle: CSSProperties = {
  color: "#d32f2f",
  fontWeight: 500,
  fontSize: "1rem",
};

export default ErrorText;
