import React from "react";
import { DateTime } from "luxon";

export type DateContextType = {
  startDate: DateTime | null;
  setStartDate: React.Dispatch<React.SetStateAction<DateTime | null>>;
  endDate: DateTime | null;
  setEndDate: React.Dispatch<React.SetStateAction<DateTime | null>>;
  invalidDate: boolean;
};

export const DateContext = React.createContext<DateContextType>({
  startDate: null,
  setStartDate: () => {},
  endDate: null,
  setEndDate: () => {},
  invalidDate: false,
});
