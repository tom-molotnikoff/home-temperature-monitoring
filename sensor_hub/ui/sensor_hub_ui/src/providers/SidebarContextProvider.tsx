import {SidebarContext} from "./SidebarContextType.tsx";
import {useState} from "react";


type SidebarContextProviderProps = {
  children: React.ReactNode;
};


export function SidebarContextProvider({children}: SidebarContextProviderProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <SidebarContext.Provider value={{open: sidebarOpen, setOpen: setSidebarOpen}}>
      {children}
    </SidebarContext.Provider>
  );
}

