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
  startDate: string | null;
  endDate: string | null;
  onStartDateChange: (date: Date | null) => void;
  onEndDateChange: (date: Date | null) => void;
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
          value={startDate ? DateTime.fromISO(startDate) : null}
          onChange={(date) =>
            onStartDateChange(date ? (date as DateTime).toJSDate() : null)
          }
        />
        <DatePicker
          label="End Date"
          value={endDate ? DateTime.fromISO(endDate) : null}
          onChange={(date) =>
            onEndDateChange(date ? (date as DateTime).toJSDate() : null)
          }
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
