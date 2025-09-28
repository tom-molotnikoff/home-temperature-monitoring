import { DatePicker } from "@mui/x-date-pickers";
import { useIsMobile } from "../hooks/useMobile";
import { useContext, type CSSProperties } from "react";
import { DateContext } from "../providers/DateContext";

function DateRangePicker() {
  const isMobile = useIsMobile();

  const { startDate, setStartDate, endDate, setEndDate, invalidDate } =
    useContext(DateContext);

  return (
    <div style={containerStyle}>
      <div
        className="date-range-picker"
        style={isMobile ? mobileLayoutStyle : desktopLayoutStyle}
      >
        <DatePicker
          label="Start Date"
          value={startDate}
          onChange={setStartDate}
        />
        <DatePicker label="End Date" value={endDate} onChange={setEndDate} />
      </div>
      {invalidDate && (
        <span className="error" style={errorStyle}>
          Invalid date range
        </span>
      )}
    </div>
  );
}

const containerStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  alignItems: "center",
};

const mobileLayoutStyle: CSSProperties = {
  display: "flex",
  flexDirection: "column",
  gap: 16,
};

const desktopLayoutStyle: CSSProperties = {
  display: "flex",
  gap: 24,
  justifyContent: "center",
  alignItems: "center",
  margin: "16px 0",
};

const errorStyle: CSSProperties = {
  color: "#d32f2f",
  marginLeft: 16,
  fontWeight: 500,
  fontSize: "1rem",
};

export default DateRangePicker;
