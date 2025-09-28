import { DateTime } from "luxon";
import { useEffect, useState } from "react";
import { DateContext } from "./DateContext";

export function DateContextProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [startDate, setStartDate] = useState<DateTime | null>(
    DateTime.now().minus({ days: 7 }).startOf("day")
  );

  const [endDate, setEndDate] = useState<DateTime | null>(
    DateTime.now().plus({ days: 1 }).startOf("day")
  );

  const [invalidDate, setInvalidDate] = useState(false);

  useEffect(() => {
    if (!startDate || !endDate) {
      setInvalidDate(true);
      return;
    }

    if (startDate >= endDate) {
      setInvalidDate(true);
      return;
    }

    setInvalidDate(false);
  }, [startDate, endDate]);

  return (
    <DateContext.Provider
      value={{ startDate, setStartDate, endDate, setEndDate, invalidDate }}
    >
      {children}
    </DateContext.Provider>
  );
}
