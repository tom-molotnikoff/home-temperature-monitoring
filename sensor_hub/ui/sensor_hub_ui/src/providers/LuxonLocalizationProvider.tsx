import { LocalizationProvider } from "@mui/x-date-pickers";
import { AdapterLuxon } from "@mui/x-date-pickers/AdapterLuxon";

function LuxonLocalizationProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <LocalizationProvider dateAdapter={AdapterLuxon} adapterLocale="en-gb">
      {children}
    </LocalizationProvider>
  );
}

export default LuxonLocalizationProvider;
