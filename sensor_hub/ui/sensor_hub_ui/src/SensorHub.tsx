import { BrowserRouter, Route, Routes } from "react-router";
import TemperatureDashboard from "./pages/TemperatureDashboard";
import LuxonLocalizationProvider from "./providers/LuxonLocalizationProvider";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import {SensorContextProvider} from "./providers/SensorContext.tsx";

function SensorHub() {
  return (
    <LuxonLocalizationProvider>
      <ThemeProvider theme={createTheme()}>
        <SensorContextProvider types={["Temperature"]} refreshIntervalMs={10000}>
          <CssBaseline />
          <BrowserRouter>
            <Routes>
              <Route path="/" element={<TemperatureDashboard />} />
            </Routes>
          </BrowserRouter>
        </SensorContextProvider>
      </ThemeProvider>
    </LuxonLocalizationProvider>
  );
}

export default SensorHub;
