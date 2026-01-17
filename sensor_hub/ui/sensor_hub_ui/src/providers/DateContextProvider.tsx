import { DateTime } from "luxon";
import { useEffect, useState } from "react";
import { DateContext } from "./DateContext";
import { useIsMobile } from "../hooks/useMobile";

export function DateContextProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const isMobile = useIsMobile();
  
  const [startDate, setStartDate] = useState<DateTime | null>(
    DateTime.now().minus({ days: 7 }).startOf("day")
  );

  const [endDate, setEndDate] = useState<DateTime | null>(
    DateTime.now().plus({ days: 1 }).startOf("day")
  );

  const [invalidDate, setInvalidDate] = useState(false);

  // Adjust default range to 2 days on mobile (only on initial mount)
  useEffect(() => {
    if (isMobile) {
      setStartDate(DateTime.now().minus({ days: 2 }).startOf("day"));
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Empty deps - only run once on mount

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
