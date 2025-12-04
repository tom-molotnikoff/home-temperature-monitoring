import {SensorContext} from "./SensorContextType.tsx";
import {useSensors} from "../hooks/useSensors.ts";


type SensorContextProviderProps = {
    children: React.ReactNode;
    type: string;
};


export function SensorContextProvider({children, type}: SensorContextProviderProps) {
    const sensors = useSensors({type});

    return (
        <SensorContext.Provider value={{sensors}}>
            {children}
        </SensorContext.Provider>
    );
}

