import {useContext} from "react";
import {SensorContext} from "../providers/SensorContextType.tsx";

export function useSensorContext() {
  return useContext(SensorContext);
}