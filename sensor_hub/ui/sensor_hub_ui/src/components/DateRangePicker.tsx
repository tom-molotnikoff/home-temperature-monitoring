import { DatePicker } from "@mui/x-date-pickers";
import { useContext, type CSSProperties } from "react";
import { DateContext } from "../providers/DateContext";
import TestIdContainer from "../tools/TestIdContainer";
import ErrorText from "../tools/ErrorText";
import DesktopRowMobileColumn from "../tools/DesktopRowMobileColumn";
import CenteredFlex from "../tools/CenteredFlex";
import { useIsMobile } from "../hooks/useMobile";

function DateRangePicker() {
  const { startDate, setStartDate, endDate, setEndDate, invalidDate } =
    useContext(DateContext);
  const isMobile = useIsMobile();

  return (
    <CenteredFlex>
      <DesktopRowMobileColumn
        testid="date-range-picker"
        desktopChanges={desktopLayoutStyleChanges}
        mobileChanges={mobileLayoutStyleChanges}
      >
        <TestIdContainer testid="start-date-picker">
          <DatePicker
            label="Start Date"
            value={startDate}
            onChange={setStartDate}
            slotProps={{
              textField: {
                fullWidth: isMobile,
              },
            }}
          />
        </TestIdContainer>
        <TestIdContainer testid="end-date-picker">
          <DatePicker 
            label="End Date" 
            value={endDate} 
            onChange={setEndDate}
            slotProps={{
              textField: {
                fullWidth: isMobile,
              },
            }}
          />
        </TestIdContainer>
      </DesktopRowMobileColumn>
      {invalidDate && (
        <ErrorText
          message="Invalid date range"
          testid="invalid-date-range-error"
        />
      )}
    </CenteredFlex>
  );
}

const desktopLayoutStyleChanges: CSSProperties = {
  gap: 24,
  justifyContent: "center",
  alignItems: "center",
  margin: "16px 0",
};

const mobileLayoutStyleChanges: CSSProperties = {
  width: "100%",
  alignItems: "stretch",
};

export default DateRangePicker;
