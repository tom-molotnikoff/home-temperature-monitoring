import { useIsMobile } from "../hooks/useMobile";

interface DesktopRowMobileColumnProps {
  children: React.ReactNode;
  testid?: string;
  desktopChanges?: React.CSSProperties;
  mobileChanges?: React.CSSProperties;
}

function DesktopRowMobileColumn({
  children,
  testid,
  desktopChanges,
  mobileChanges,
}: DesktopRowMobileColumnProps) {
  const isMobile = useIsMobile();

  const desktopCardLayoutStyle: React.CSSProperties = {
    display: "flex",
    gap: 16,
    ...desktopChanges,
  };

  const mobileCardLayoutStyle: React.CSSProperties = {
    display: "flex",
    gap: 16,
    alignItems: "center",
    flexDirection: "column",
    ...mobileChanges,
  };

  const finalStyle = isMobile ? mobileCardLayoutStyle : desktopCardLayoutStyle;

  return (
    <div style={{ ...finalStyle }} data-testid={testid}>
      {children}
    </div>
  );
}

export default DesktopRowMobileColumn;
