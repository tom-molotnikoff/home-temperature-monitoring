import LuxonLocalizationProvider from "./providers/LuxonLocalizationProvider";
import CssBaseline from "@mui/material/CssBaseline";
import ProviderWrapper from "./providers/ProviderWrapper.tsx";
import AppRoutes from "./navigation/AppRoutes.tsx";


function SensorHub() {
  return (
    <LuxonLocalizationProvider>
      <ProviderWrapper>
            <CssBaseline />
            <AppRoutes />
      </ProviderWrapper>
    </LuxonLocalizationProvider>
  );
}

export default SensorHub;
