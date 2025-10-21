import React from "react";
import {SensorContextProvider} from "./SensorContext.tsx";
import { ThemeProvider } from "@mui/material";
import { SidebarContextProvider } from "./SidebarContextProvider.tsx";
import LuxonLocalizationProvider from "./LuxonLocalizationProvider.tsx";
import {createTheme} from "@mui/material/styles";
import {SensorTypes} from "../types/types.ts";

interface ProviderWrapperProps {
  children: React.ReactNode
}

function ProviderWrapper({ children }: ProviderWrapperProps) {
  const theme = createTheme({
    colorSchemes: { light: true, dark: true },
    cssVariables: {
      colorSchemeSelector: 'class'
    }
  });

  return (
    <LuxonLocalizationProvider>
      <ThemeProvider theme={theme}>
        <SidebarContextProvider>
          <SensorContextProvider types={SensorTypes} refreshIntervalMs={3000}>
            {children}
          </SensorContextProvider>
        </SidebarContextProvider>
      </ThemeProvider>
    </LuxonLocalizationProvider>
  )
}

export default ProviderWrapper