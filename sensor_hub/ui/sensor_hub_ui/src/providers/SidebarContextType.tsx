import {createContext, type Dispatch, type SetStateAction} from "react";

type SidebarContextType = {
  open: boolean;
  setOpen: Dispatch<SetStateAction<boolean>>
};

export const SidebarContext = createContext<SidebarContextType>({open: false, setOpen: () => {}});
