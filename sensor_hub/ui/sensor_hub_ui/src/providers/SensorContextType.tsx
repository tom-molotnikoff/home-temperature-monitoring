import {createContext} from "react";
import type {Sensor} from "../types/types.ts";

type SensorContextValueType = {
  sensors: Sensor[];
};

export const SensorContext = createContext<SensorContextValueType>({sensors: []});
