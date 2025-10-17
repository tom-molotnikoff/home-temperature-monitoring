import {SensorContext} from "./SensorContextType.tsx";
import {useSensors} from "../hooks/useSensors.ts";


type SensorContextProviderProps = {
    children: React.ReactNode;
    types: string[];
    refreshIntervalMs?: number;
};


export function SensorContextProvider({children, types, refreshIntervalMs = 3000}: SensorContextProviderProps) {
    const sensors = useSensors({types, refreshIntervalMs});

    return (
        <SensorContext.Provider value={{sensors}}>
            {children}
        </SensorContext.Provider>
    );
}

