import {SensorContext} from "./SensorContextType.tsx";
import {useSensors} from "../hooks/useSensors.ts";


type SensorContextProviderProps = {
    children: React.ReactNode;
    type: string;
};


export function SensorContextProvider({children, type}: SensorContextProviderProps) {
    const { sensors, loaded } = useSensors({type});

    return (
        <SensorContext.Provider value={{sensors, loaded}}>
            {children}
        </SensorContext.Provider>
    );
}

