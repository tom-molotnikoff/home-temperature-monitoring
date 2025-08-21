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
  return (
    <div
      className="date-range-picker"
      style={{
        display: "flex",
        gap: "24px",
        justifyContent: "center",
        alignItems: "center",
        margin: "16px 0",
      }}
    >
      <input
        type="date"
        value={startDate ? startDate : ""}
        onChange={(e) => onStartDateChange(new Date(e.target.value))}
        style={{
          fontSize: "1.1rem",
          padding: "8px 12px",
          borderRadius: "6px",
          border: "1px solid #ccc",
          boxShadow: "0 1px 4px rgba(0,0,0,0.04)",
        }}
      />
      <input
        type="date"
        value={endDate ? endDate : ""}
        onChange={(e) => onEndDateChange(new Date(e.target.value))}
        style={{
          fontSize: "1.1rem",
          padding: "8px 12px",
          borderRadius: "6px",
          border: "1px solid #ccc",
          boxShadow: "0 1px 4px rgba(0,0,0,0.04)",
        }}
      />
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
