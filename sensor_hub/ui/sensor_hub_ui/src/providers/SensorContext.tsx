import {SensorContext} from "./SensorContextType.tsx";
import {useSensors} from "../hooks/useSensors.ts";


type SensorContextProviderProps = {
    children: React.ReactNode;
};


export function SensorContextProvider({children}: SensorContextProviderProps) {
    const { sensors, loaded } = useSensors();

    return (
        <SensorContext.Provider value={{sensors, loaded}}>
            {children}
        </SensorContext.Provider>
    );
}

