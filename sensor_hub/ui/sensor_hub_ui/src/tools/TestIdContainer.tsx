interface TestIdContainerProps {
  children: React.ReactNode;
  testid?: string;
}

function TestIdContainer({ children, testid }: TestIdContainerProps) {
  return <div data-testid={testid}>{children}</div>;
}

export default TestIdContainer;
