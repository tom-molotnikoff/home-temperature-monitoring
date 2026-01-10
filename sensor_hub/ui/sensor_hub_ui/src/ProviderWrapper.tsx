import React from "react";
import {SensorContextProvider} from "./providers/SensorContext";
import { ThemeProvider } from "@mui/material";
import { SidebarContextProvider } from "./providers/SidebarContextProvider";
import LuxonLocalizationProvider from "./providers/LuxonLocalizationProvider";
import {createTheme} from "@mui/material/styles";
import AuthProvider from './providers/AuthProvider';

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
        <AuthProvider>
          <SidebarContextProvider>
            <SensorContextProvider type="Temperature">
              {children}
            </SensorContextProvider>
          </SidebarContextProvider>
        </AuthProvider>
      </ThemeProvider>
    </LuxonLocalizationProvider>
  )
}

export default ProviderWrapper
