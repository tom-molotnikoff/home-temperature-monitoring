import {createContext} from "react";
import type {Sensor} from "../gen/aliases";

type SensorContextValueType = {
  sensors: Sensor[];
  loaded: boolean;
};

export const SensorContext = createContext<SensorContextValueType>({sensors: [], loaded: false});
