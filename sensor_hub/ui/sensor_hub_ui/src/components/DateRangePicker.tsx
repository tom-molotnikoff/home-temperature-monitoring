import { DatePicker } from "@mui/x-date-pickers";
import { DateTime } from "luxon";
import { useIsMobile } from "../hooks/useMobile";

function DateRangePicker({
  startDate,
  endDate,
  onStartDateChange,
  onEndDateChange,
  invalidDate,
}: {
  startDate: DateTime | null;
  endDate: DateTime | null;
  onStartDateChange: (date: DateTime | null) => void;
  onEndDateChange: (date: DateTime | null) => void;
  invalidDate?: boolean;
}) {
  const isMobile = useIsMobile();

  return (
    <div
      style={{ display: "flex", flexDirection: "column", alignItems: "center" }}
    >
      <div
        className="date-range-picker"
        style={
          isMobile
            ? {
                display: "flex",
                flexDirection: "column",
                gap: "16px",
              }
            : {
                display: "flex",
                gap: "24px",
                justifyContent: "center",
                alignItems: "center",
                margin: "16px 0",
              }
        }
      >
        <DatePicker
          label="Start Date"
          value={startDate}
          onChange={onStartDateChange}
        />
        <DatePicker
          label="End Date"
          value={endDate}
          onChange={onEndDateChange}
        />
      </div>
      {invalidDate && (
        <span
          className="error"
          style={{
            color: "#d32f2f",
            marginLeft: "16px",
            fontWeight: 500,
            fontSize: "1rem",
          }}
        >
          Invalid date range
        </span>
      )}
    </div>
  );
}

export default DateRangePicker;
