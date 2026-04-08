import {SensorContext} from "./SensorContextType.tsx";
import {useSensors} from "../hooks/useSensors.ts";


type SensorContextProviderProps = {
    children: React.ReactNode;
    driver: string;
};


export function SensorContextProvider({children, driver}: SensorContextProviderProps) {
    const { sensors, loaded } = useSensors({driver});

    return (
        <SensorContext.Provider value={{sensors, loaded}}>
            {children}
        </SensorContext.Provider>
    );
}

