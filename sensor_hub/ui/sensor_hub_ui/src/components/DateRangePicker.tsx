import { DatePicker } from "@mui/x-date-pickers";
import { useContext, type CSSProperties } from "react";
import { DateContext } from "../providers/DateContext";
import TestIdContainer from "../tools/TestIdContainer";
import ErrorText from "../tools/ErrorText";
import DesktopRowMobileColumn from "../tools/DesktopRowMobileColumn";
import CenteredFlex from "../tools/CenteredFlex";

function DateRangePicker() {
  const { startDate, setStartDate, endDate, setEndDate, invalidDate } =
    useContext(DateContext);

  return (
    <CenteredFlex>
      <DesktopRowMobileColumn
        testid="date-range-picker"
        desktopChanges={desktopLayoutStyleChanges}
      >
        <TestIdContainer testid="start-date-picker">
          <DatePicker
            label="Start Date"
            value={startDate}
            onChange={setStartDate}
          />
        </TestIdContainer>
        <TestIdContainer testid="end-date-picker">
          <DatePicker label="End Date" value={endDate} onChange={setEndDate} />
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

export default DateRangePicker;
