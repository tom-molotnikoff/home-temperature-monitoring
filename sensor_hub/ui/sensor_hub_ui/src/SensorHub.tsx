import { BrowserRouter, Route, Routes } from "react-router";
import "./App.css";
import TemperatureDashboard from "./pages/TemperatureDashboard";
import LuxonLocalizationProvider from "./providers/LuxonLocalizationProvider";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";

function SensorHub() {
  return (
    <LuxonLocalizationProvider>
      <ThemeProvider theme={createTheme()}>
        <CssBaseline />
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<TemperatureDashboard />} />
          </Routes>
        </BrowserRouter>
      </ThemeProvider>
    </LuxonLocalizationProvider>
  );
}

export default SensorHub;
