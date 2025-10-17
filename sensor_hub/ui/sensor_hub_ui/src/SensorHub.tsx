import { BrowserRouter, Route, Routes } from "react-router";
import TemperatureDashboard from "./pages/temperature-dashboard/TemperatureDashboard.tsx";
import LuxonLocalizationProvider from "./providers/LuxonLocalizationProvider";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import {SensorContextProvider} from "./providers/SensorContext.tsx";
import SensorsOverview from "./pages/sensors-overview/SensorsOverview.tsx";
import {SidebarContextProvider} from "./providers/SidebarContextProvider.tsx";

const theme = createTheme({
  colorSchemes: { light: true, dark: true },
  cssVariables: {
    colorSchemeSelector: 'class'
  }
});

function SensorHub() {
  return (
    <LuxonLocalizationProvider>
      <ThemeProvider theme={theme}>
        <SidebarContextProvider>
          <SensorContextProvider types={["Temperature"]} refreshIntervalMs={10000}>
            <CssBaseline />
            <BrowserRouter>
              <Routes>
                <Route path="/" element={<TemperatureDashboard />} />
                <Route path="/sensors-overview" element={<SensorsOverview />} />
              </Routes>
            </BrowserRouter>
          </SensorContextProvider>
        </SidebarContextProvider>
      </ThemeProvider>
    </LuxonLocalizationProvider>
  );
}

export default SensorHub;
